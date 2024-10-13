package certificate

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/mckern/spiry/internal/spiry"
)

const defaultTLSPort = "443"

type Certificate struct {
	addr     string
	name     string
	insecure bool
	raw      *x509.Certificate
}

var _ spiry.ExpiringResource = (*Certificate)(nil)

type Command struct {
	DomainName string `name:"name" short:"n" help:"request TLS certificate for domain <name> instead of <address>"`
	Insecure   bool   `name:"insecure" short:"k" help:"allow insecure server connections"`
	Addr       string `arg:"" name:"address" help:"address to retrieve TLS certificate from"`
}

func (c *Command) Run(globals *spiry.Command) (err error) {
	cert, err := New(c.Addr)
	if err != nil {
		return err
	}

	if c.DomainName != "" {
		cert, err = NewWithName(c.DomainName, c.Addr)
		if err != nil {
			return err
		}
	}

	output, err := globals.Render(cert)
	fmt.Println(output)

	return err
}

func New(address string) (cert *Certificate, err error) {
	addr, err := parseAddr(address)
	if err != nil {
		return cert, err
	}
	return &Certificate{addr: addr}, err
}

func NewWithName(name string, addr string) (*Certificate, error) {
	if !govalidator.IsDNSName(name) {
		slog.Debug("invalid DNS name given", "name", name)
		return nil, fmt.Errorf("%q is an invalid DNS name", name)
	}

	addr, err := parseAddr(addr)
	if err != nil {
		return nil, err
	}
	return &Certificate{addr: addr, name: name}, err
}

func (c *Certificate) Expiry() (time.Time, error) {
	// if NotAfter already has a valid value, use it
	if c.raw != nil && !c.raw.NotAfter.IsZero() {
		return c.raw.NotAfter, nil
	}

	cert, err := c.getCert()
	if err != nil {
		// no cert to read time from, so use time.Time's zero value
		return time.Time{},
			fmt.Errorf("unable to retrieve certificate for %v: %w", c.addr, err)
	}
	c.raw = cert

	return c.raw.NotAfter, err
}

func (c *Certificate) Name() (name string) {
	if c.name != "" {
		return c.name
	}

	name, _, _ = net.SplitHostPort(c.addr)
	return
}

func (c *Certificate) getCert() (cert *x509.Certificate, err error) {
	tlsConfig := &tls.Config{
		// this is intentionally done to allow
		// retrieval of any TLS certificate -- we only
		// care about its expiration date.
		InsecureSkipVerify: true,
		ServerName:         c.Name()}

	dialer := &net.Dialer{
		Timeout: time.Millisecond * time.Duration(1000),
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", c.addr, tlsConfig)
	if err != nil {
		return
	}

	defer conn.Close()
	certs := conn.ConnectionState().PeerCertificates
	cert = certs[0]
	return
}

func parseAddr(addr string) (parsedAddress string, err error) {
	if govalidator.IsURL(addr) {
		parsedAddress, err = parseAsURL(addr)
		if err == nil {
			return
		}
	}

	isIP := govalidator.IsIP(addr)
	isDNSName := govalidator.IsDNSName(addr)

	slog.Debug("looking for IP address or DNS name",
		"address", addr,
		"ip", isIP,
		"dnsname", isDNSName)
	if isIP || isDNSName {
		slog.Debug("attempting to parse address with default TLS port",
			"address", addr,
			"port", defaultTLSPort)
		defaultAddr := net.JoinHostPort(addr, defaultTLSPort)
		parsedAddress, err = parseAsHostPort(defaultAddr)
		if err == nil {
			return
		}
	}

	slog.Debug("attempting to parse as Host:Port pair", "address", addr)
	parsedAddress, err = parseAsHostPort(addr)
	return
}

func parseAsHostPort(addr string) (parsedAddress string, err error) {
	// test to see if this is already a valid host:port pair
	slog.Debug("trying to parse address as Host:Port pair",
		"address", addr)
	name, port, err := net.SplitHostPort(addr)
	if err != nil {
		slog.Debug("failed to parse address as Host:Port pair", "address", addr)
		return parsedAddress,
			fmt.Errorf("cannot parse %q as Host:Port pair: %w", addr, err)
	}

	if (!govalidator.IsIP(name) && !govalidator.IsDNSName(name)) ||
		!govalidator.IsPort(port) {
		slog.Debug("unable to parse address as valid Host:Port pair",
			"name", name,
			"port", port)
		return parsedAddress, fmt.Errorf("%q is an invalid Host:Port pair", name)
	}

	parsedAddress = net.JoinHostPort(name, port)
	if parsedAddress == addr {
		return parsedAddress, nil
	}

	return
}

func parseAsURL(addr string) (parsedAddress string, err error) {
	slog.Debug("trying to parse address as URL", "address", addr)
	u, err := url.Parse(addr)

	// either addr failed to parse entirely, and url.Parse explicitly rejected it,
	// or addr failed to parse correctly but url.Parse did not return an error,
	// leaving the pieces of data that we care about unusable
	if err != nil || (u.Scheme == "" && u.Host == "") {
		slog.Debug("failed to parse address", "address", addr)
		return parsedAddress, fmt.Errorf("failed to parse address %q", addr)
	}

	// the address parsed, but the domain name passed is invalid
	if !govalidator.IsDNSName(u.Hostname()) {
		slog.Debug("invalid DNS name", "domain", u.Hostname())
		return parsedAddress, fmt.Errorf("domain %q is an invalid DNS name", u.Hostname())
	}

	// host and port were both defined after parsing,so they will be explicitly used
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
