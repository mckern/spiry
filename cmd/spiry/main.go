package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/araddon/dateparse"
	"github.com/mckern/spiry/internal/console"
	"github.com/mckern/spiry/internal/spiry"
	flag "github.com/spf13/pflag"
)

// Basic information about `spiry` itself, and
// the canonical root-level whois address FOR THE WORLD
const (
	name = "spiry"
	// ISO8601 is not one of Go's built-in formats :(
	ISO8601 = "2006-01-02T15:04:05-0700"
)

var (
	// metadata about the program itself, derived from compile-time variables
	buildDate     string
	gitCommit     string
	versionNumber string
	whoami        = path.Base(os.Args[0])
	// formats      map[string]string
)

var (
	// and the runtime flags, which kinda-sorta have to be global
	flags        = flag.NewFlagSet(whoami, flag.ExitOnError)
	bareFlag     bool
	jsonFlag     bool
	unixFlag     bool
	rfc1123zFlag bool
	rfc3339Flag  bool
	helpFlag     bool
	versionFlag  bool
)

func init() {
	flags.SortFlags = false

	flags.BoolVarP(&bareFlag,
		"bare", "b", false,
		"only display expiration date")

	flags.BoolVarP(&jsonFlag,
		"json", "j", false,
		"display output as JSON")

	flags.BoolVarP(&unixFlag,
		"unix", "u", false,
		"display expiration date as UNIX time")

	flags.BoolVarP(&rfc1123zFlag,
		"rfc1123z", "r", false,
		"display expiration date as RFC1123Z timestamp")

	flags.BoolVarP(&rfc3339Flag,
		"rfc3339", "R", false,
		"display expiration date as RFC3339 timestamp")

	flags.BoolVarP(&versionFlag,
		"version", "v", false,
		"display version information and exit")

	flags.BoolVarP(&helpFlag,
		"help", "h", false,
		"display this help and exit")

	// If this program is aliased to a different name, use that in
	// help output because it's what a user would expect to see.
	sayMyNameSayMyName := whoami
	if whoami == name {
		sayMyNameSayMyName = name
	}

	flags.Usage = func() {
		fmt.Fprintf(flags.Output(), "%s: look up domain name expiration\n\n", sayMyNameSayMyName)
		fmt.Fprintf(flags.Output(), "usage: %s [-b|-j] [-u|-r|-R] [-h|-v] <domain>\n", whoami)
		flags.PrintDefaults()
		fmt.Fprintln(flags.Output(),
			"\nenvironment variables:\n"+
				"  SPIRY_DEBUG:   print debug messages")
	}

	// parse every argument passed, except the name of the calling program
	err := flags.Parse(os.Args[1:])
	console.Fatal(err)

	// user asked for help
	if helpFlag {
		console.Warn("--help requested, all other flags ignored")

		flags.SetOutput(os.Stdout)
		flags.Usage()
		os.Exit(0)
	}

	// user asked for version information
	if versionFlag {
		console.Warn("--version requested, all other flags ignored")

		prettyDate, _ := dateparse.ParseAny(buildDate)

		fmt.Printf("%s\t%s\n", name, versionNumber)
		fmt.Print("Copyright (C) 2019 by Ryan McKern <ryan@mckern.sh>\n")
		fmt.Print("Web site: https://github.com/mckern/spiry\n")
		fmt.Print("Build information:\n")
		fmt.Printf("    git commit ref: %s\n", gitCommit)
		fmt.Printf("    build date:     %s\n", prettyDate.Format(ISO8601))
		fmt.Printf("\n%s comes with ABSOLUTELY NO WARRANTY.\n"+
			"This is free software, and you are welcome to redistribute\n"+
			"it under certain conditions. See the Parity Public License\n"+
			"(version 7.0.0) for details.\n", whoami)

		os.Exit(0)
	}

	// too many arguments passed
	if len(flags.Args()) > 1 {
		fmt.Fprintf(os.Stderr, "ERROR: too many arguments\n\n")
		flags.Usage()
		os.Exit(1)
	}

	// too few arguments passed
	if len(flags.Args()) < 1 {
		fmt.Fprintf(os.Stderr, "ERROR: too few arguments; domain name required\n\n")
		flags.Usage()
		os.Exit(1)
	}

	// mutually exclusive output flags used
	if bareFlag && jsonFlag {
		fmt.Fprintf(os.Stderr, "ERROR: cannot use --bare and --json together\n\n")
		flags.Usage()
		os.Exit(1)
	}
}

func main() {
	domain := spiry.Domain{Name: flags.Arg(0)}

	rootDomain, err := domain.Root()
	console.Fatal(err)
	console.Debug(fmt.Sprintf("found root domain %q for FQDN %q", rootDomain, domain.Name))

	tld, err := domain.TLD()
	console.Fatal(err)
	console.Debug(fmt.Sprintf("found eTLD %q for root domain %q", tld, rootDomain))

	tldServer, err := domain.CanonicalWhoisServer()
	console.Fatal(err)
	console.Debug(fmt.Sprintf("found canonical whois server %q for eTLD %q\n", tldServer, tld))

	expiry, err := domain.Expiry()
	console.Fatal(err)

	timeFmt := expiry.Format(ISO8601)
	if unixFlag {
		timeFmt = strconv.FormatInt(expiry.Unix(), 10)
	} else if rfc1123zFlag {
		timeFmt = expiry.Format(time.RFC1123Z)
	} else if rfc3339Flag {
		timeFmt = expiry.Format(time.RFC3339)
	}

	// default output formatting first
	output := fmt.Sprintf("%s\t%s", rootDomain, timeFmt)

	// refine output formatting if a user requested
	// something besides the default values
	if bareFlag {
		output = timeFmt
	}

	if jsonFlag {
		jsonStruct := make(map[string]string)
		jsonStruct["domain"] = rootDomain
		jsonStruct["expiry"] = timeFmt

		json, err := json.MarshalIndent(jsonStruct, "", "  ")
		console.Fatal(err)

		output = string(json)
	}

	fmt.Println(output)
}
