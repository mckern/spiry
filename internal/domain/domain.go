package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/araddon/dateparse"
	"github.com/likexian/whois"
	whoisparser "github.com/likexian/whois-parser"
	"github.com/mckern/spiry/internal/console"
	"golang.org/x/net/publicsuffix"
)

type domain struct {
	name        string
	WhoisServer string
	expiryDate  time.Time
}

func New(name string) *domain {
	return &domain{name: name}
}

func (d *domain) Name() string {
	return d.name
}

// Root returns the root domain (example.com, example.net, etc.) of a
// given fully-qualified domain name.
// It returns a String if successful, otherwise it will
// return an empty String and any errors encountered.
func (d *domain) Root() (string, error) {
	root, err := publicsuffix.EffectiveTLDPlusOne(d.name)
	if err != nil {
		return "", err
	}

	console.Debug(fmt.Sprintf("found root domain %q for FQDN %q", root, d.name))
	return root, err
}

// TLD returns the top-level domain (.com, .net, etc.) of a
// given fully-qualified domain name according to the semi-canonical
// list maintained at https://publicsuffix.org/.
// It returns a String if successful, otherwise it will
// return an empty String and any errors encountered.
func (d *domain) TLD() (string, error) {
	root, err := d.Root()
	if err != nil {
		return "",
			fmt.Errorf("unable to look up eTLD for domain %v: %w", d.name, err)
	}

	etld, icannManaged := publicsuffix.PublicSuffix(d.name)

	// domain is not actually managed according to https://publicsuffix.org/
	// so we should give up now
	if !icannManaged {
		return "",
			fmt.Errorf("eTLD root %q is not publicly managed and cannot be looked up using `whois`",
				root)
	}

	console.Debug(fmt.Sprintf("found eTLD %q for root domain %q", etld, d.name))
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
				d.name, err)
	}

	// derive root of domain, so we aren't trying to
	// query subdomains
	root, err := d.Root()
	if err != nil {
		return ex,
			fmt.Errorf("unable to find domain root for %v: %w",
				d.name, err)
	}

	// query for whois data
	record, err := whois.Whois(root, d.WhoisServer)
	if err != nil {
		return ex,
			fmt.Errorf("(expiry) whois request for domain %v failed: %w",
				root, err)
	}

	result, err := whoisparser.Parse(record)
	if err != nil {
		errorMsg := fmt.Errorf("parsing whois record for domain %v failed: %w", root, err)

		if errors.Is(err, whoisparser.ErrNotFoundDomain) {
			errorMsg = fmt.Errorf("domain record %q not found", root)
		} else if errors.Is(err, whoisparser.ErrDomainDataInvalid) {
			errorMsg = fmt.Errorf("whois record %q is invalid", root)
		}

		return ex, errorMsg
	}

	d.expiryDate, _ = dateparse.ParseAny(result.Domain.ExpirationDate)
	return d.expiryDate, err
}
