package main

import (
	"fmt"
	"math"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/araddon/dateparse"
	whois "github.com/likexian/whois-go"
	flag "github.com/mckern/pflag"
	whoisparser "github.com/mckern/whois-parser-go"
	"golang.org/x/net/publicsuffix"
)

// Basic information about `spiry` itself
const (
	name = "spiry"
	iana = "whois.iana.org"
)

var buildDate string
var gitCommit string
var versionNumber string

var version = fmt.Sprintf("%s (%s) %s", whoami, name, versionNumber)
var whoami = path.Base(os.Args[0])

func init() {
	var versionFlag bool
	var helpFlag bool

	flag.BoolVarP(&helpFlag, "help", "h", false, "show this help")
	flag.BoolVarP(&versionFlag, "version", "v", false, "print version number")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s: print number of days until a domain name expires\n\n", whoami)
		fmt.Fprintf(flag.CommandLine.Output(), "usage: %s [-hv] [-b address]\n", whoami)
		flag.PrintDefaults()
	}

	flag.Parse()

	if helpFlag {
		flag.CommandLine.SetOutput(os.Stdout)
		flag.Usage()
		os.Exit(0)
	}

	if versionFlag {
		prettyDate, _ := dateparse.ParseAny(buildDate)

		fmt.Printf("%s\n\n", version)
		fmt.Printf("git commit hash: %s\n", gitCommit)
		fmt.Printf("build date: %s\n", prettyDate)
		fmt.Printf("runtime: %s/%s\n", runtime.GOOS, runtime.GOARCH)

		os.Exit(0)
	}
}

func tld(domain string) (string, error) {
	var err error
	eTLD, icann := publicsuffix.PublicSuffix(domain)

	if !icann {
		err = fmt.Errorf("domain '%v' is unmanaged and cannot be looked up", domain)
	}

	return eTLD, err
}

func lookupTLDserver(tld string) (string, error) {
	var err error
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

func daysToExpiry(expiryDate time.Time) (days, hours float64) {
	now := time.Now()

	days = expiryDate.Sub(now).Hours() / 24
	hours = math.Mod(expiryDate.Sub(now).Hours(), 24)
	return
}

func main() {
	if len(os.Args) <= 1 {
		flag.Usage()
		os.Exit(1)
	}

	domain := os.Args[1]
	tld, err := tld(domain)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	tldServer, err := lookupTLDserver(tld)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	record, err := whois.Whois(domain, tldServer)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	result, err := whoisparser.Parse(record)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	if err == nil {
		expiry, _ := dateparse.ParseAny(result.Registrar.ExpirationDate)
		days, hours := daysToExpiry(expiry)

		fmt.Printf("(%v) %.0f days, %.0f hours\n", expiry, days, hours)
	}
}
