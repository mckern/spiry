package main

import (
	"fmt"
	"math"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/araddon/dateparse"
	"github.com/likexian/whois-go"
	whoisparser "github.com/likexian/whois-parser-go"
	flag "github.com/mckern/pflag"
)

// Basic information about `spiry` itself
const name = "spiry"

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

func main() {
	if len(os.Args) <= 1 {
		flag.Usage()
		os.Exit(1)
	}

	record, err := whois.Whois(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	fmt.Printf("%v", record)

	result, err := whoisparser.Parse(record)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	if err == nil {
		expiry, _ := dateparse.ParseAny(result.Registrar.ExpirationDate)
		now := time.Now()
		days := expiry.Sub(now).Hours() / 24
		hours := math.Mod(expiry.Sub(now).Hours(), 24)

		fmt.Printf("(%v) %.0f days, %.0f hours\n", expiry, days, hours)
	}
}
