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

type Certificate struct {
	addr string
	name string
	raw  x509.Certificate
}

func New(address string) (cert *Certificate, err error) {
	addr, err := parseAddr(address)
	if err != nil {
		return cert, err
	}
	return &Certificate{addr: addr}, err
}

func NewWithName(name string, address string) (*Certificate, error) {
	if !isDomainName(name) {
		return nil, fmt.Errorf("certificate: invalid name %q given", name)
	}

	addr, err := parseAddr(address)
	if err != nil {
		return nil, err
	}
	return &Certificate{addr: addr, name: name}, err
}

func (c *Certificate) Expiry() (time.Time, error) {
	var err error

	if !c.raw.NotAfter.IsZero() {
		return c.raw.NotAfter, err
	}

	c.raw, err = c.getCert()
	if err != nil {
		return c.raw.NotAfter, fmt.Errorf("unable to retrieve certificate for %v: %w", c.addr, err)
	}

	return c.raw.NotAfter, err
}

func (c *Certificate) Name() (name string) {
	if c.name != "" {
		return c.name
	}

	name, _, _ = net.SplitHostPort(c.addr)
	return
}

func (c *Certificate) getCert() (cert x509.Certificate, err error) {
	tlsConfig := &tls.Config{
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
	console.Debug(fmt.Sprintf("trying to parse address %v", addr))
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

// copied wholesale from stdlib's net.isDomainName() func,
// because it's not exported by stdlib.
// ----
// isDomainName checks if a string is a presentation-format domain name
// (currently restricted to hostname-compatible "preferred name" LDH labels and
// SRV-like "underscore labels"; see golang.org/issue/12421).
func isDomainName(s string) bool {
	// The root domain name is valid. See golang.org/issue/45715.
	if s == "." {
		return true
	}

	// See RFC 1035, RFC 3696.
	// Presentation format has dots before every label except the first, and the
	// terminal empty label is optional here because we assume fully-qualified
	// (absolute) input. We must therefore reserve space for the first and last
	// labels' length octets in wire format, where they are necessary and the
	// maximum total length is 255.
	// So our _effective_ maximum is 253, but 254 is not rejected if the last
	// character is a dot.
	l := len(s)
	if l == 0 || l > 254 || l == 254 && s[l-1] != '.' {
		return false
	}

	last := byte('.')
	nonNumeric := false // true once we've seen a letter or hyphen
	partlen := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		default:
			return false
		case 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || c == '_':
			nonNumeric = true
			partlen++
		case '0' <= c && c <= '9':
			// fine
			partlen++
		case c == '-':
			// Byte before dash cannot be dot.
			if last == '.' {
				return false
			}
			partlen++
			nonNumeric = true
		case c == '.':
			// Byte before dot cannot be dot, dash.
			if last == '.' || last == '-' {
				return false
			}
			if partlen > 63 || partlen == 0 {
				return false
			}
			partlen = 0
		}
		last = c
	}
	if last == '-' || partlen > 63 {
		return false
	}

	return nonNumeric
}
