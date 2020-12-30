package domain

import (
	"fmt"
	"time"

	"github.com/araddon/dateparse"
	whois "github.com/likexian/whois-go"
	whoisparser "github.com/likexian/whois-parser-go"
	"github.com/mckern/spiry/internal/console"
	"golang.org/x/net/publicsuffix"
)

// const iana = "whois.iana.org"

type domain struct {
	Name        string
	WhoisServer string
	expiryDate  time.Time
}

func New(name string) *domain {
	return &domain{Name: name}
}

// Root returns the root domain (example.com, example.net, etc.) of a
// given fully-qualified domain name.
// It returns a String if successful, otherwise it will
// return an empty String and any errors encountered.
func (d *domain) Root() (string, error) {
	root, err := publicsuffix.EffectiveTLDPlusOne(d.Name)
	if err != nil {
		return "", err
	}

	console.Debug(fmt.Sprintf("found root domain %q for FQDN %q", root, d.Name))
	return root, err
}

// TLD returns the top-level domain (.com, .net, etc.) of a
// given fully-qualified domain name according to the semi-canonical
// list maintained at https://publicsuffix.org/.
// It returns a String if successfull, otherwise it will
// return an empty String and any errors encountered.
func (d *domain) TLD() (string, error) {
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

// Expiry returns the expiration date of a given fully-qualified
// domain name according to public DNS records.
// It returns a time.Time value if successfull, otherwise it will
// return any errors encountered.
func (d *domain) Expiry() (ex time.Time, err error) {
	if !d.expiryDate.IsZero() {
		return d.expiryDate, nil
	}

	// ensure this is not a private or invalid domain
	_, err = d.TLD()
	if err != nil {
		return ex,
			fmt.Errorf("unable to find eTLD for domain %v: %w",
				d.Name, err)
	}

	// derive root of domain, so we aren't trying to
	// query subdomains
	root, err := d.Root()
	if err != nil {
		return ex,
			fmt.Errorf("unable to find domain root for %v: %w",
				d.Name, err)
	}

	// query for whois data
	record, err := whois.Whois(root, d.WhoisServer)
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
