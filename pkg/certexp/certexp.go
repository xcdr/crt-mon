package certexp

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

type ExpirationError struct {
	Message string
	Code    int
}

func (e *ExpirationError) Error() string {
	return e.Message
}

type ExpirationDetail struct {
	Issuer  string
	Subject string
	Date    time.Time
	Days    int
}

type ProcessResult struct {
	Address net.IP
	Expiry  ExpirationDetail
	Error   ExpirationError
}

type HostInfo struct {
	Port int
	Name string
}

type Check struct {
	Host   HostInfo
	Result []ProcessResult
}

func NewCheck(host HostInfo) *Check {
	return &Check{Host: host}
}

func (check *Check) Expiration(IPv6 bool) error {
	var error ExpirationError
	var today time.Time = time.Now()

	addresses, err := net.LookupIP(check.Host.Name)
	if err != nil {
		error.Code = 4
		error.Message = err.Error()

		return &error
	}

	cfg := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         check.Host.Name}

	for _, addr := range addresses {
		if addr.To4() == nil && IPv6 == false {
			// Skip IPv6 addresses
			continue
		}

		var expiry ExpirationDetail

		error.Code = 0
		error.Message = ""

		conn, err := tls.Dial("tcp", fmt.Sprintf("[%s]:%d", addr.String(), check.Host.Port), cfg)

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

		check.Result = append(check.Result, ProcessResult{Address: addr, Expiry: expiry, Error: error})
	}

	return nil
}
