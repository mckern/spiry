package certificate

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/mckern/spiry/internal/console"
)

const defaultTLSPort = "443"

type certificate struct {
	Addr string
	raw  x509.Certificate
}

func New(address string) (cert *certificate, err error) {
	addr, err := parseAddr(address)
	if err != nil {
		return cert, err
	}
	return &certificate{Addr: addr}, err
}

func (c *certificate) Expiry() (time.Time, error) {
	var err error

	if !c.raw.NotAfter.IsZero() {
		return c.raw.NotAfter, err
	}

	c.raw, err = c.getCert()
	if err != nil {
		return c.raw.NotAfter, fmt.Errorf("unable to retrieve certificate: %w", err)
	}

	return c.raw.NotAfter, err
}

func (c *certificate) Name() (name string) {
	name, _, _ = net.SplitHostPort(c.Addr)
	return
}

func (c *certificate) getCert() (cert x509.Certificate, err error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         c.Name()}

	dialer := &net.Dialer{
		Timeout: time.Millisecond * time.Duration(1000),
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", c.Addr, tlsConfig)
	if err != nil {
		return
	}

	defer conn.Close()
	certs := conn.ConnectionState().PeerCertificates
	cert = *certs[0]
	return
}

func parseAddr(addr string) (parsedAddress string, err error) {
	parsedAddress, err = parseAsURL(addr)
	if err == nil {
		return
	}
	console.Debug(fmt.Sprintf("cannot parse %q as URL: %v", addr, err.Error()))

	parsedAddress, err = parseAsHostPort(addr)
	if err == nil {
		return
	}
	console.Debug(fmt.Sprintf("cannot parse %q as Host:Port pair: %v", addr, err.Error()))

	defaultAddr := net.JoinHostPort(addr, defaultTLSPort)
	console.Debug(fmt.Sprintf("attempting to parse %q with default port as %q", addr, defaultAddr))
	parsedAddress, err = parseAsHostPort(defaultAddr)
	return
}

func parseAsHostPort(addr string) (parsedAddress string, err error) {
	// test to see if this is already a valid host:port pair
	name, port, err := net.SplitHostPort(addr)
	if err == nil {
		parsedAddress = net.JoinHostPort(name, port)
		if parsedAddress == addr {
			console.Debug(fmt.Sprintf("address %q is already a host:port pair", parsedAddress))
			return
		}
	}
	return
}

func parseAsURL(addr string) (parsedAddress string, err error) {
	u, err := url.Parse(addr)

	// addr failed to parse entirely, and url.Parse rejected it
	if err != nil {
		msg := fmt.Sprintf("failed to parse address %q", addr)
		console.Debug(msg)
		return parsedAddress, fmt.Errorf(msg)
	}

	// addr failed to parse correctly, and the pieces of data
	// that we care about are not usable
	if u.Scheme == "" && u.Host == "" {
		msg := fmt.Sprintf("failed to parse address %q correctly", addr)
		console.Debug(msg)
		return parsedAddress, fmt.Errorf(msg)
	}

	// host and port were both defined, and will be used explicitly
	if u.Host != "" && u.Port() != "" {
		parsedAddress = net.JoinHostPort(u.Hostname(), u.Port())
		return
	}

	// attempt to derive an explicit port from the URL scheme
	// as no port has been specified
	port, err := net.LookupPort("tcp", u.Scheme)
	if err == nil {
		parsedAddress = net.JoinHostPort(u.Hostname(), fmt.Sprint(port))
	}

	return
}
