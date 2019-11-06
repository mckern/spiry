package main

import (
	"fmt"
	"math"
	"os"
	"path"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	whois "github.com/likexian/whois-go"
	flag "github.com/mckern/pflag"
	whoisparser "github.com/mckern/whois-parser-go"
	"golang.org/x/net/publicsuffix"
)

// Basic information about `spiry` itself, and
// the canonical root-level whois address FOR THE WORLD
const (
	name = "spiry"
	iana = "whois.iana.org"
	// ISO8601 is not one of Go's built-in formats :(
	iso8601 = "2006-01-02T15:04:05-0700"
)

var (
	// metadata about the program itself, derived from compile-time variables
	buildDate     string
	gitCommit     string
	versionNumber string
	whoami        = path.Base(os.Args[0])

	// and the runtime flags, which kinda-sorta have to be global
	flags       = flag.NewFlagSet(whoami, flag.ExitOnError)
	bareFlag    bool
	helpFlag    bool
	humanFlag   bool
	versionFlag bool
)

func debug(msg string) {
	value, defined := os.LookupEnv("SPIRY_DEBUG")
	if defined && value != "" {
		fmt.Fprintf(os.Stderr, "DEBUG: %s\n", strings.Trim(msg, "\n"))
	}
}

func warn(msg string) {
	value, defined := os.LookupEnv("SPIRY_DEBUG")
	if defined && value != "" {
		fmt.Fprintf(os.Stderr, "WARNING: %s\n", strings.Trim(msg, "\n"))
	}
}

func fatalErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err.Error())
		os.Exit(1)
	}
}

func eTLD(domain string) string {
	eTLD, icannManaged := publicsuffix.PublicSuffix(domain)

	// domain is not actually managed according to https://publicsuffix.org/
	// so we should give up now
	if !icannManaged {
		fatalErr(fmt.Errorf("eTLD root %q is not publicly managed and cannot be looked up using `whois`", eTLD))
	}

	return eTLD
}

func lookupTLDserver(tld string) (string, error) {
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

func humanReadableExpiry(expiryDate time.Time) (days, hours int64) {
	now := time.Now()
	days = int64(expiryDate.Sub(now).Hours() / 24)
	hours = int64(math.Mod(expiryDate.Sub(now).Hours(), 24))
	return
}

func init() {
	flags.SortFlags = false

	flags.BoolVarP(&bareFlag,
		"bare", "b", false,
		"display the bare expiration date in some mish-mash unix format that might be RFCish?")

	flags.BoolVarP(&humanFlag,
		"human-readable", "H", false,
		"print the human-readable number of days until expiration")

	flags.BoolVarP(&versionFlag,
		"version", "v", false,
		"display version information and exit")

	flags.BoolVarP(&helpFlag,
		"help", "h", false,
		"display this help and exit")

	flags.Usage = func() {
		fmt.Fprintf(flags.Output(), "%s: print number of days until a domain name expires\n\n", whoami)
		fmt.Fprintf(flags.Output(), "usage: %s [-h|-v|-b|-H] <domain>\n", whoami)
		flags.PrintDefaults()
		fmt.Fprintln(flags.Output(),
			"\nenvironment variables:\n"+
				"  SPIRY_DEBUG:   print debug messages")
	}

	flags.Parse(os.Args[1:])

	// user asked for help
	if helpFlag {
		if bareFlag || humanFlag || versionFlag {
			warn("--help requested, all other flags ignored")
		}

		flags.SetOutput(os.Stdout)
		flags.Usage()
		os.Exit(0)
	}

	// user asked for version information
	if versionFlag {
		if bareFlag || humanFlag || helpFlag {
			warn("--version requested, all other flags ignored")
		}

		prettyDate, _ := dateparse.ParseAny(buildDate)

		fmt.Printf("%s\t%s\n", whoami, versionNumber)
		fmt.Print("Copyright (C) 2019 by Ryan McKern <ryan@mckern.sh>\n")
		fmt.Print("Web site: https://github.com/mckern/spiry\n")
		fmt.Print("Build information:\n")
		fmt.Printf("    git commit ref: %s\n", gitCommit)
		fmt.Printf("    build date:     %s\n", prettyDate.Format(iso8601))
		fmt.Printf("\n%s comes with ABSOLUTELY NO WARRANTY.\n"+
			"This is free software, and you are welcome to redistribute\n"+
			"it under certain conditions. See the Parity Public License\n"+
			"(version 6.0.0) for details.\n", whoami)

		os.Exit(0)
	}

	// too many arguments passed
	if len(flags.Args()) > 1 {
		fmt.Fprintf(os.Stderr, "ERROR: too many arguments\n\n")
		flags.Usage()
		os.Exit(1)
	}

	// mutually exclusive output flags used
	if bareFlag && humanFlag {
		fmt.Fprintf(os.Stderr, "ERROR: cannot use --bare and --human-readable together\n\n")
		flags.Usage()
		os.Exit(1)
	}
}

func main() {
	domain := flags.Arg(0)
	rootDomain, err := publicsuffix.EffectiveTLDPlusOne(domain)
	fatalErr(err)
	debug(fmt.Sprintf("found root domain %q for FQDN %q", rootDomain, domain))

	tld := eTLD(rootDomain)
	debug(fmt.Sprintf("found eTLD %q for root domain %q", tld, rootDomain))

	tldServer, err := lookupTLDserver(tld)
	fatalErr(err)
	debug(fmt.Sprintf("found canonical whois server %q for eTLD %q\n", tldServer, tld))

	record, err := whois.Whois(rootDomain, tldServer)
	fatalErr(err)
	debug(fmt.Sprintf("canonical whois server %q returned a record for root domain %q\n", tldServer, rootDomain))

	// whoisparser does not seem to reliably catch domains that report
	// as not-found, so we've got to manually look for those
	if whoisparser.IsNotFound(record) {
		fatalErr(fmt.Errorf("canonical whois server %q reports domain %q as unregistered", tldServer, domain))
	}

	result, err := whoisparser.Parse(record)
	fatalErr(err)
	debug(fmt.Sprintf("successfully parsed record for root domain %q\n", rootDomain))

	expiry, _ := dateparse.ParseAny(result.Registrar.ExpirationDate)
	output := fmt.Sprintf("%s\t%s", domain, expiry.Format(iso8601))

	if bareFlag {
		output = expiry.Format(iso8601)
	}

	if humanFlag {
		days, hours := humanReadableExpiry(expiry)
		output = fmt.Sprintf("%s expires in %d days, %d hours", domain, days, hours)
	}

	fmt.Println(output)
}
