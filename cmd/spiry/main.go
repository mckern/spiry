package main

import (
	"fmt"
	"os"
	"path"

	"github.com/mckern/spiry/internal/console"
	"github.com/mckern/spiry/internal/spiry"

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

func init() {
	flags.SortFlags = false

	flags.BoolVarP(&bareFlag,
		"bare", "b", false,
		"display expiration date as ISO8601 timestamp")

	flags.BoolVarP(&humanFlag,
		"human-readable", "H", false,
		"display a human-readable number of days until expiration")

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
			console.Warn("--help requested, all other flags ignored")
		}

		flags.SetOutput(os.Stdout)
		flags.Usage()
		os.Exit(0)
	}

	// user asked for version information
	if versionFlag {
		if bareFlag || humanFlag || helpFlag {
			console.Warn("--version requested, all other flags ignored")
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
	console.Fatal(err)
	console.Debug(fmt.Sprintf("found root domain %q for FQDN %q", rootDomain, domain))

	tld := spiry.ETLD(rootDomain)
	console.Debug(fmt.Sprintf("found eTLD %q for root domain %q", tld, rootDomain))

	tldServer, err := spiry.LookupTLDserver(tld)
	console.Fatal(err)
	console.Debug(fmt.Sprintf("found canonical whois server %q for eTLD %q\n", tldServer, tld))

	record, err := whois.Whois(rootDomain, tldServer)
	console.Fatal(err)
	console.Debug(fmt.Sprintf("canonical whois server %q returned a record for root domain %q\n", tldServer, rootDomain))

	// whoisparser does not seem to reliably catch domains that report
	// as not-found, so we've got to manually look for those
	if whoisparser.IsNotFound(record) {
		console.Fatal(fmt.Errorf("canonical whois server %q reports domain %q as unregistered", tldServer, domain))
	}

	result, err := whoisparser.Parse(record)
	console.Fatal(err)
	console.Debug(fmt.Sprintf("successfully parsed record for root domain %q\n", rootDomain))

	expiry, _ := dateparse.ParseAny(result.Registrar.ExpirationDate)
	output := fmt.Sprintf("%s\t%s", domain, expiry.Format(iso8601))

	if bareFlag {
		output = expiry.Format(iso8601)
	}

	if humanFlag {
		days, hours := spiry.HumanReadableExpiry(expiry)
		output = fmt.Sprintf("%s expires in %d days, %d hours", domain, days, hours)
	}

	fmt.Println(output)
}
