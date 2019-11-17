package spiry

import (
	"fmt"
	"math"
	"time"

	"github.com/likexian/whois-go"
	"github.com/mckern/spiry/internal/console"
	whoisparser "github.com/mckern/whois-parser-go"
	"golang.org/x/net/publicsuffix"
)

const (
	iana = "whois.iana.org"
)

func ETLD(domain string) string {
	eTLD, icannManaged := publicsuffix.PublicSuffix(domain)

	// domain is not actually managed according to https://publicsuffix.org/
	// so we should give up now
	if !icannManaged {
		console.Fatal(fmt.Errorf("eTLD root %q is not publicly managed and cannot be looked up using `whois`", eTLD))
	}

	return eTLD
}

func LookupTLDserver(tld string) (string, error) {
	record, err := whois.Whois(tld, iana)
	if err != nil {
		return "", err
	}

	result, err := whoisparser.Parse(record)
	if err != nil {
		return "", err
	}

	server := result.Registrar.WhoisServer
	return server, err
}

func HumanReadableExpiry(expiryDate time.Time) (days, hours int64) {
	now := time.Now()
	days = int64(expiryDate.Sub(now).Hours() / 24)
	hours = int64(math.Mod(expiryDate.Sub(now).Hours(), 24))
	return
}
