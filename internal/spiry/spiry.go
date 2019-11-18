package spiry

import (
	"fmt"
	"time"

	"github.com/araddon/dateparse"
	whois "github.com/likexian/whois-go"
	"github.com/mckern/spiry/internal/console"
	whoisparser "github.com/mckern/whois-parser-go"
	"golang.org/x/net/publicsuffix"
)

const (
	iana = "whois.iana.org"
)

type Domain struct {
	Name        string
	whoisServer string
	expiryDate  time.Time
}

func (d *Domain) Root() (string, error) {
	root, err := publicsuffix.EffectiveTLDPlusOne(d.Name)
	if err != nil {
		return "", err
	}
	return root, err
}

func (d *Domain) TLD() (string, error) {
	root, err := d.Root()
	if err != nil {
		return "",
			fmt.Errorf("unable to look up eTLD for domain %v: %w", d.Name, err)
	}

	etld, icannManaged := publicsuffix.PublicSuffix(d.Name)

	// domain is not actually managed according to https://publicsuffix.org/
	// so we should give up now
	if !icannManaged {
		return "",
			fmt.Errorf("eTLD root %q is not publicly managed and cannot be looked up using `whois`",
				root)
	}

	return etld, err
}

func (d *Domain) CanonicalWhoisServer() (string, error) {
	if len(d.whoisServer) != 0 {
		return d.whoisServer, nil
	}

	tld, err := d.TLD()
	if err != nil {
		return "",
			fmt.Errorf("unable to look up canonical whois server for domain %v: %w",
				d.Name, err)
	}

	record, err := whois.Whois(tld, iana)
	if err != nil {
		return "",
			fmt.Errorf("(whoisServer) whois request for domain %q failed: %w",
				tld, err)
	}

	result, err := whoisparser.Parse(record)
	if err != nil {
		return "",
			fmt.Errorf("parsing whois record for domain %v failed: %w",
				tld, err)
	}

	d.whoisServer = result.Registrar.WhoisServer
	return d.whoisServer, err
}

func (d *Domain) Expiry() (ex time.Time, err error) {
	if !d.expiryDate.IsZero() {
		return d.expiryDate, nil
	}

	root, err := d.Root()
	if err != nil {
		return ex,
			fmt.Errorf("unable to find domain root for %v: %w",
				d.Name, err)
	}

	addr, err := d.CanonicalWhoisServer()
	if err != nil {
		return ex,
			fmt.Errorf("cannot look up canonical whois server for domain %v: %w",
				d.Name, err)
	}

	record, err := whois.Whois(root, addr)
	if err != nil {
		return ex,
			fmt.Errorf("(expiry) whois request for domain %v failed: %w",
				root, err)
	}

	// whoisparser does not seem to reliably catch domains that report
	// as not-found, so we've got to manually look for those
	if whoisparser.IsNotFound(record) {
		console.Fatal(fmt.Errorf("canonical whois server %q reports domain %q as unregistered", addr, root))
	}

	result, err := whoisparser.Parse(record)
	if err != nil {
		return ex,
			fmt.Errorf("parsing whois record for domain %v failed: %w",
				root, err)
	}

	d.expiryDate, _ = dateparse.ParseAny(result.Registrar.ExpirationDate)
	return d.expiryDate, err
}
