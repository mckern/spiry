package domain

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/asaskevich/govalidator"
	"github.com/likexian/whois"
	whoisparser "github.com/likexian/whois-parser"
	"github.com/mckern/spiry/internal/spiry"
	"golang.org/x/net/idna"
	"golang.org/x/net/publicsuffix"
)

var IncompleteTLDs = []string{
	"ac", "ad", "al", "an", "ao", "aq", "ar", "aw", "ba",
	"bf", "bh", "bm", "bs", "bt", "bv", "bw", "bz", "cd",
	"cg", "ck", "cm", "cr", "cu", "cv", "cw", "cy", "dj",
	"do", "eg", "er", "et", "fj", "fk", "fm", "ga", "gb",
	"ge", "gf", "gh", "gm", "gn", "gp", "gq", "gr", "gt",
	"gu", "gw", "hm", "jm", "jo", "kh", "km", "kn", "kp",
	"kw", "ky", "lb", "lc", "lk", "lr", "ls", "mc", "mh",
	"mil", "mk", "mm", "mq", "mr", "mt", "mv", "mw", "mz",
	"ne", "ni", "np", "nr", "pa", "pg", "ph", "pk", "pn",
	"ps", "py", "rw", "sd", "sj", "sl", "sr", "sv", "sz",
	"td", "tg", "tj", "to", "tp", "tt", "va", "vi", "vn",
	"vu", "ye", "za", "zm", "zw",

	// punycode conversions of IDNs
	"xn--0zwm56d", "xn--11b5bs3a9aj6g", "xn--45brj9c",
	"xn--80akhbyknj4f", "xn--90a3ac", "xn--9t4b11yi5a",
	"xn--deba0ad", "xn--fpcrj9c3d", "xn--fzc2c9e2c",
	"xn--g6w251d", "xn--gecrj9c", "xn--h2brj9c",
	"xn--hgbk6aj7f53bba", "xn--hlcj6aya9esc7a", "xn--jxalpdlp",
	"xn--kgbechtv", "xn--l1acc", "xn--mgbayh7gpa",
	"xn--mgbbh1a71e", "xn--mgbc0a9azcg", "xn--pgbs0dh",
	"xn--s9brj9c", "xn--wgbh1c", "xn--xkc2al3hye2a",
	"xn--xkc2dl3a5ee0h", "xn--zckzah",
}

type Domain struct {
	name        string
	WhoisServer string
	expiryDate  time.Time
}

var _ spiry.ExpiringResource = (*Domain)(nil)

type Command struct {
	DomainName string `arg:"" name:"domain" help:"top-level domain name to look up"`
	ServerAddr string `name:"server" short:"s" help:"use <server> as specific whois server"`
}

func (d *Command) Run(globals *spiry.Command) (err error) {
	domainName, err := New(d.DomainName)
	if err != nil {
		return
	}

	output, err := globals.Render(domainName)
	if err != nil {
		return err
	}

	fmt.Println(output)
	return
}

func New(name string) (*Domain, error) {
	name = strings.ToLower(name)
	if !govalidator.IsDNSName(name) {
		slog.Debug("invalid DNS name given", "name", name)
		return nil, fmt.Errorf("%q is an invalid DNS name", name)
	}

	return &Domain{name: name}, nil
}

func (d *Domain) Name() string {
	return d.name
}

// Root returns the root domain (example.com, example.net, etc.) of a
// given fully-qualified domain name.
// It returns a String if successful, otherwise it will
// return an empty String and any errors encountered.
func (d *Domain) Root() (string, error) {
	root, err := publicsuffix.EffectiveTLDPlusOne(d.name)
	if err != nil {
		slog.Debug("unable to find root domain from FQDN",
			"fqdn", d.name)
		return "", err
	}

	return root, err
}

// TLD returns the top-level domain (.com, .net, etc.) of a
// given fully-qualified domain name according to the semi-canonical
// list maintained at https://publicsuffix.org/.
// It returns a String if successful, otherwise it will
// return an empty String and any errors encountered.
func (d *Domain) TLD() (string, error) {
	root, err := d.Root()
	if err != nil {
		return "",
			fmt.Errorf("unable to look up eTLD for domain %v: %w", d.name, err)
	}

	// note that publicsuffix.PublicSuffix assumes case-sensitive comparison
	// and all of its reference domains are lowercase.
	// while all domains were cast to lowercase in New(),
	// they're cast to lowercase here  just in case.
	etld, icannManaged := publicsuffix.PublicSuffix(strings.ToLower(d.name))
	// domain is not actually managed according to https://publicsuffix.org/
	// so we should give up now
	if !icannManaged {
		return "",
			fmt.Errorf("eTLD root %q is not publicly managed and cannot be looked up using the whois network",
				root)
	}

	// check for internationalized domains and convert them
	// to their ascii equivalent for comparison; if they don't map,
	// then give up.
	etld, err = idna.ToASCII(etld)
	if err != nil {
		return "", err
	}

	// check if the domain comes from a known-incomplete registry,
	// and let the user know they may not get what they're looking for.
	if slices.Contains(IncompleteTLDs, etld) {
		fmt.Fprintf(os.Stderr,
			"warning: TLD %q returns incomplete WHOIS data; you may not be able to look up an expiration date\n", etld)
	}

	return etld, err
}

// Expiry returns the expiration date of a given fully-qualified
// domain name according to public DNS records.
// It returns a time.Time value if successful, otherwise it will
// return any errors encountered.
func (d *Domain) Expiry() (ex time.Time, err error) {
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

		// figure out what kind of error was returned and whether
		// it warrants additional information/context
		if errors.Is(err, whoisparser.ErrNotFoundDomain) {
			errorMsg = fmt.Errorf("domain record %q not found", root)
		} else if errors.Is(err, whoisparser.ErrReservedDomain) {
			errorMsg = fmt.Errorf("reserved domain record %q cannot be looked up", root)
		}

		return ex, errorMsg
	}

	d.expiryDate, err = dateparse.ParseAny(result.Domain.ExpirationDate)
	return d.expiryDate, err
}
