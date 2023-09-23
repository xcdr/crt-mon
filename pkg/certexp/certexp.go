package certexp

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

// Represents details of check.
type ExpirationDetail struct {
	Issuer  string
	Subject string
	Date    time.Time
	Days    int
}

// Represents expiration error message and code.
type ExpirationError struct {
	Message string
	Code    int
}

func (e *ExpirationError) Error() string {
	return e.Message
}

// Domain represents info about domain parsed from config file and its addresses.
type Domain struct {
	Name        string   `yaml:"domain"`
	Addresses   []net.IP `yaml:"addresses"`
	Port        int      `yaml:"port"`
	SkipResolve bool     `yaml:"skip_resolve"`
}

/*
Resolves DNS records for domain.

- IPv6: when false resolves only IPv4 addresses.
*/
func (d *Domain) Resolve(IPv6 bool) error {
	var error ExpirationError

	if !d.SkipResolve {
		addresses, err := net.LookupIP(d.Name)

		if err != nil {
			error.Code = 4
			error.Message = err.Error()

			return &error
		}

		for _, addr := range addresses {
			if addr.To4() == nil && !IPv6 {
				// Skip IPv6 addresses
				continue
			}

			d.Addresses = append(d.Addresses, addr)
		}
	}

	return nil
}

// Represents info about checked host.
type HostInfo struct {
	Name    string
	Address net.IP
	Port    int
}

// Represents result of check against single IP.
type CheckResult struct {
	Address net.IP
	Expiry  ExpirationDetail
	Error   ExpirationError
}

// Represents list of check results for host.
type Check struct {
	Host   HostInfo
	Result []CheckResult
}

// Returns new check for specified host.
func NewCheck(host HostInfo) *Check {
	return &Check{Host: host}
}

// Verify expiration against checked host.
func (check *Check) Process() error {
	var error ExpirationError
	var today time.Time = time.Now()

	cfg := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         check.Host.Name}

	var expiry ExpirationDetail

	error.Code = 0
	error.Message = ""

	conn, err := tls.Dial("tcp", fmt.Sprintf("[%s]:%d", check.Host.Address.String(), check.Host.Port), cfg)

	if err != nil {
		error.Code = 3
		error.Message = err.Error()
	} else {
		// Must be deferred only after error handled and connected!
		defer conn.Close()

		err = conn.VerifyHostname(check.Host.Name)
		if err != nil {
			error.Code = 2
			error.Message = err.Error()
		} else {
			expiry.Date = conn.ConnectionState().PeerCertificates[0].NotAfter
			expiry.Days = int(expiry.Date.Sub(today).Hours() / 24)
			expiry.Subject = conn.ConnectionState().PeerCertificates[0].Subject.String()
			expiry.Issuer = conn.ConnectionState().PeerCertificates[0].Issuer.String()

			if expiry.Days <= 0 {
				error.Code = 1
			}
		}
	}

	check.Result = append(check.Result, CheckResult{Address: check.Host.Address, Expiry: expiry, Error: error})

	return nil
}
