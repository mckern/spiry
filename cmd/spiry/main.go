package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/araddon/dateparse"
	"github.com/mckern/spiry/internal/console"
	"github.com/mckern/spiry/internal/domain"
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
	flags        = flag.NewFlagSet(path.Base(os.Args[0]), flag.ContinueOnError)
	serverAddr   string
	bareFlag     bool
	jsonFlag     bool
	unixFlag     bool
	rfc1123zFlag bool
	rfc3339Flag  bool
	helpFlag     bool
	versionFlag  bool
)

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

func versionMsg() string {
	prettyDate, _ := dateparse.ParseAny(buildDate)

	msg := fmt.Sprintf("%s\t%s\n", name, versionNumber)
	msg += "Copyright (C) 2019 by Ryan McKern <ryan@mckern.sh>\n"
	msg += fmt.Sprintf("Web site: %s\n", url)
	msg += "Build information:\n"
	msg += fmt.Sprintf("    git commit ref: %s\n", gitCommit)
	msg += fmt.Sprintf("    build date:     %s\n", prettyDate.Format(ISO8601))
	msg += fmt.Sprintf("\n%s comes with ABSOLUTELY NO WARRANTY.\n"+
		"This is free software, and you are welcome to\n"+
		"redistribute it under certain conditions.\n"+
		"See the MIT License for details.\n", name)

	return msg
}

func init() {
	// flags are explicitly sorted here, and we don't need
	// pflag sorting them alphabetically on us.
	flags.SortFlags = false

	flags.StringVarP(&serverAddr,
		"server", "s", "",
		"use <server> as specific whois server")

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
		fmt.Fprintf(flags.Output(),
			"%s: look up domain name expiration\n\n",
			flags.Name())
		fmt.Fprintf(flags.Output(),
			"usage: %s [-b|-j] [-u|-r|-R] [-h|-v] [-s <server>] <domain>\n",
			flags.Name())

		flags.PrintDefaults()
		fmt.Fprintln(flags.Output(),
			"\nenvironment variables:\n"+
				"  SPIRY_DEBUG:   print debug messages")
	}

	// initialize an error collector, so we can display all error
	// output and save users the hassle of rerunning multiple times,
	// one error message at a time
	var errMsgs []string

	// parse every argument passed, except the name of the calling program
	err := flags.Parse(os.Args[1:])
	console.Fatal(err, func() {
		console.Error(errors.New("failed to parse unknown flags, cowardly aborting"))
	})

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

		fmt.Print(versionMsg())
		os.Exit(0)
	}

	// mutually exclusive output flags used
	if !flagsAreMutuallyExclusive(bareFlag, jsonFlag) {
		errMsgs = append(errMsgs, "cannot use --bare and --json together")
	}

	if !flagsAreMutuallyExclusive(rfc1123zFlag, rfc3339Flag, unixFlag) {
		errMsgs = append(errMsgs, "cannot use --rfc1123z, --rfc3339, and --unix together")
	}

	if len(flags.Args()) > 1 { // too many arguments passed
		errMsgs = append(errMsgs, "too many arguments")
	} else if len(flags.Args()) == 0 { // too few arguments passed
		errMsgs = append(errMsgs, "too few arguments; domain name required")
		console.Warn(fmt.Sprintf("%v", errMsgs))
	}

	// check our error collection and if there's anything
	// in there, print it, display Usage(), and exit 1
	if len(errMsgs) > 0 {
		for _, errMsg := range errMsgs {
			console.Error(errors.New(errMsg))
		}
		// print a newline on stderr for some whitespace
		fmt.Fprintln(os.Stderr)
		flags.Usage()
		os.Exit(1)
	}
}

func main() {
	domain := domain.New(flags.Arg(0))
	domain.WhoisServer = serverAddr

	rootDomain, err := domain.Root()
	console.Fatal(err)

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

		jsonOut, err := json.MarshalIndent(jsonStruct, "", "  ")
		console.Fatal(err)

		output = string(jsonOut)
	}

	// print constructed output, because we're done
	fmt.Println(output)
}
