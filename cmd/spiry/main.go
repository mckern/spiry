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
	url  = "https://github.com/mckern/spiry"
	// ISO8601 is not one of Go's built-in formats :(
	ISO8601 = "2006-01-02T15:04:05-0700"
)

var (
	// metadata about the program itself, derived from compile-time variables
	buildDate     string
	gitCommit     string
	versionNumber string
)

var (
	// and the runtime flags, which kinda-sorta have to be global
	flags        = flag.NewFlagSet(path.Base(os.Args[0]), flag.ExitOnError)
	bareFlag     bool
	jsonFlag     bool
	unixFlag     bool
	rfc1123zFlag bool
	rfc3339Flag  bool
	helpFlag     bool
	versionFlag  bool
)

// Fprint formats using the default formats for its operands and writes to w.
// Spaces are added between operands when neither is a string.
// It returns the number of bytes written and any write error encountered.

// flagsAreMutuallyExclusive takes any number of booleans and returns
// true if 0 or 1 of them are true; otherwise it returns false.
func flagsAreMutuallyExclusive(f ...bool) bool {
	counter := 0
	for _, flagValue := range f {
		if flagValue {
			if counter += 1; counter > 1 {
				return false
			}
		}
	}
	return true
}

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

	flags.Usage = func() {
		// If this program is aliased to a different name, use that in
		// help output because it's what a user would expect to see.
		fmt.Fprintf(flags.Output(), "%s: look up domain name expiration\n\n", flags.Name())
		fmt.Fprintf(flags.Output(), "usage: %s [-b|-j] [-u|-r|-R] [-h|-v] <domain>\n", flags.Name())
		flags.PrintDefaults()
		fmt.Fprintln(flags.Output(),
			"\nenvironment variables:\n"+
				"  SPIRY_DEBUG:   print debug messages")
	}

	// parse every argument passed, except the name of the calling program
	err := flags.Parse(os.Args[1:])
	console.Fatal(err)

	// user asked for help; give them what
	// they asked for and exit successfully
	if helpFlag {
		console.Warn("--help requested, all other flags ignored")

		flags.SetOutput(os.Stdout)
		flags.Usage()
		os.Exit(0)
	}

	// user asked for version information; give them what
	// they asked for and exit successfully
	if versionFlag {
		console.Warn("--version requested, all other flags ignored")

		prettyDate, _ := dateparse.ParseAny(buildDate)

		fmt.Printf("%s\t%s\n", name, versionNumber)
		fmt.Print("Copyright (C) 2019 by Ryan McKern <ryan@mckern.sh>\n")
		fmt.Printf("Web site: %s\n", url)
		fmt.Print("Build information:\n")
		fmt.Printf("    git commit ref: %s\n", gitCommit)
		fmt.Printf("    build date:     %s\n", prettyDate.Format(ISO8601))
		fmt.Printf("\n%s comes with ABSOLUTELY NO WARRANTY.\n"+
			"This is free software, and you are welcome to redistribute\n"+
			"it under certain conditions. See the Parity Public License\n"+
			"(version 7.0.0) for details.\n", name)

		os.Exit(0)
	}

	// initialize an error collector, so we can display all error
	// output and save users the hassle of rerunning multiple times,
	// one error message at a time
	var errors []string

	// mutually exclusive output flags used
	if !flagsAreMutuallyExclusive(bareFlag, jsonFlag) {
		errors = append(errors, "cannot use --bare and --json together")
	}

	if !flagsAreMutuallyExclusive(rfc1123zFlag, rfc3339Flag, unixFlag) {
		errors = append(errors, "cannot use --rfc1123z, --rfc3339, and --unix together")
	}

	// too many arguments passed
	if len(flags.Args()) > 1 {
		errors = append(errors, "too many arguments")
	}

	// too few arguments passed
	if len(flags.Args()) < 1 {
		errors = append(errors, "too few arguments; domain name required")
	}

	// check our error collection and if there's anything
	// in there, print it, display Usage(), and exit 1
	if len(errors) > 0 {
		for _, programError := range errors {
			console.Error(programError)
		}

		fmt.Fprintln(os.Stderr)

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

	// define a default time format
	timeFmt := expiry.Format(ISO8601)
	if unixFlag {
		timeFmt = strconv.FormatInt(expiry.Unix(), 10)
	} else if rfc1123zFlag {
		timeFmt = expiry.Format(time.RFC1123Z)
	} else if rfc3339Flag {
		timeFmt = expiry.Format(time.RFC3339)
	}

	// define a default output formatting
	output := fmt.Sprintf("%s\t%s", rootDomain, timeFmt)

	// redefine output formatting if a user requested
	// something besides the default values
	if bareFlag {
		output = timeFmt
	} else if jsonFlag {
		jsonStruct := make(map[string]string)
		jsonStruct["domain"] = rootDomain
		jsonStruct["expiry"] = timeFmt

		json, err := json.MarshalIndent(jsonStruct, "", "  ")
		console.Fatal(err)

		output = string(json)
	}

	// print constructed output, because we're done
	fmt.Println(output)
}
