package certexp

import (
	"crypto/tls"
	"fmt"
	"net"
	"strings"
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

// Represents info about domain parsed from config file and its addresses.
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

	if d.Name == "" {
		return nil
	}

	if !d.SkipResolve {
		addresses, err := net.LookupIP(d.Name)

		if err != nil {
			error.Code = 4
			error.Message = err.Error()
			d.Addresses = append(d.Addresses, nil)

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
func (check *Check) Process(timeout int) error {
	var error ExpirationError
	var today time.Time = time.Now()

	cfg := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         check.Host.Name}

	var expiry ExpirationDetail

	if check.Host.Address == nil {
		error.Code = 4
		error.Message = "domain resolve error"
		check.Result = append(check.Result, CheckResult{Address: check.Host.Address, Expiry: expiry, Error: error})

		return nil
	}

	error.Code = 0
	error.Message = ""

	// conn, err := tls.Dial("tcp", fmt.Sprintf("[%s]:%d", check.Host.Address.String(), check.Host.Port), cfg)
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: time.Duration(timeout) * time.Second}, "tcp", fmt.Sprintf("[%s]:%d", check.Host.Address.String(), check.Host.Port), cfg)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "connect: network is unreachable"):
			error.Code = 3
			error.Message = "connect: network is unreachable"
		case strings.Contains(err.Error(), "tls: failed to verify certificate:"):
			error.Code = 2
			error.Message = strings.Replace(err.Error(), "tls: failed to verify certificate: ", "", 1)
		default:
			error.Code = 5
			error.Message = err.Error()
		}

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
