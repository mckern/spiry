package spiry

import (
	"fmt"
	"time"

	"github.com/araddon/dateparse"
	"github.com/likexian/whois-go"
	whoisparser "github.com/likexian/whois-parser-go"
	"github.com/mckern/spiry/internal/console"
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

// Root will return the root domain of a given domain.
// It returns a String if successful, otherwise it will
// return an empty string and any errors encountered.
func (d *Domain) Root() (string, error) {
	root, err := publicsuffix.EffectiveTLDPlusOne(d.Name)
	if err != nil {
		return "", err
	}

	console.Debug(fmt.Sprintf("found root domain %q for FQDN %q", root, d.Name))
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

	console.Debug(fmt.Sprintf("found eTLD %q for root domain %q", etld, d.Name))
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

	d.whoisServer = result.Domain.WhoisServer
	console.Debug(fmt.Sprintf("found canonical whois server %q for eTLD %q\n", d.whoisServer, tld))
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

	record, err := whois.Whois(root)
	if err != nil {
		return ex,
			fmt.Errorf("(expiry) whois request for domain %v failed: %w",
				root, err)
	}

	// whoisparser does not seem to reliably catch domains that report
	// as not-found, so we've got to manually look for those
	if whoisparser.IsNotFound(record) {
		console.Fatal(fmt.Errorf("whois reports domain %q as unregistered or expired", root))
	}

	result, err := whoisparser.Parse(record)

	if err != nil {
		return ex,
			fmt.Errorf("parsing whois record for domain %v failed: %w",
				root, err)
	}

	d.expiryDate, _ = dateparse.ParseAny(result.Domain.ExpirationDate)
	return d.expiryDate, err
}
