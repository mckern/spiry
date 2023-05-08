package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mckern/spiry/internal/certificate"
	"github.com/mckern/spiry/internal/console"
	"github.com/mckern/spiry/internal/domain"

	"github.com/alecthomas/kong"
	"github.com/araddon/dateparse"
)

// Basic information about `spiry` itself, and
// the canonical root-level whois address FOR THE WORLD
const (
	name    = "spiry"
	url     = "https://github.com/mckern/spiry"
	ISO8601 = "2006-01-02T15:04:05-0700"
)

var (
	// metadata about the program itself, derived from compile-time variables
	buildDate     string
	gitCommit     string
	versionNumber string
)

type resource interface {
	Name() string
	Expiry() (time.Time, error)
}

func versionMsg() string {
	prettyDate, _ := dateparse.ParseAny(buildDate)

	msg := fmt.Sprintf("%s\t%s\n", name, versionNumber)
	msg += "\nCopyright (C) 2019 by Ryan McKern <ryan@mckern.sh>\n"
	msg += fmt.Sprintf("Web site: %s\n", url)
	msg += "Build information:\n"
	msg += fmt.Sprintf("    git commit ref: %s\n", gitCommit)
	msg += fmt.Sprintf("    build date:     %s\n", prettyDate.Format(time.RFC3339))
	msg += fmt.Sprintf("\n%s comes with ABSOLUTELY NO WARRANTY.\n"+
		"This is free software, and you are welcome to redistribute\n"+
		"it under certain conditions. See the MIT License for details.\n", name)

	return msg
}

// ///////// Domain command struct ////////////

type DomainCmd struct {
	DomainName string `arg:"" name:"domain" help:"top-level domain name to look up"`
	ServerAddr string `name:"server" short:"s" help:"use <server> as specific whois server"`
}

func (d *DomainCmd) Run(globals *Globals) (err error) {
	domainName := domain.New(d.DomainName)
	output, err := globals.render(domainName)
	if err == nil {
		fmt.Println(output)
	}

	return err
}

// ///////// Certificate command struct ////////////

type CertCmd struct {
	Addr       string `arg:"" name:"address" help:"address to retrieve TLS certificate from"`
	DomainName string `name:"name" short:"n" help:"request TLS certificate for domain <name> instead of <address>"`
}

func (c *CertCmd) Run(globals *Globals) (err error) {
	cert, err := certificate.New(c.Addr)
	if err != nil {
		return err
	}

	if c.DomainName != "" {
		cert, err = certificate.NewWithName(c.DomainName, c.Addr)
		if err != nil {
			return err
		}
	}

	output, err := globals.render(cert)
	fmt.Println(output)

	return err
}

// ///////// Global command structs ////////////

type Globals struct {
	Debug   bool             `short:"D" help:"Enable debug mode"`
	Version kong.VersionFlag `name:"version" short:"v" help:"display version information and exit"`

	// output formatting flags are mutually exclusive
	BareFlag bool   `name:"bare" short:"b" xor:"output" help:"only display expiration date"`
	JsonFlag bool   `name:"json" short:"j" xor:"output" help:"display output as JSON"`
	Output   string `kong:"-"`

	// time formatting flags are mutually exclusive
	UnixFlag     bool   `name:"unix" short:"u" xor:"time" help:"display expiration date as UNIX timestamp"`
	Rfc1123zFlag bool   `name:"rfc1123z" short:"r" xor:"time" help:"display expiration date as RFC1123Z timestamp"`
	Rfc3339Flag  bool   `name:"rfc3339" short:"R" xor:"time" help:"display expiration date as RFC3339 timestamp"`
	Time         string `kong:"-"`
}

func (g *Globals) render(res resource) (output string, err error) {
	expiry, err := res.Expiry()
	if err != nil {
		return output, err
	}

	// define a default time format
	timeFmt := expiry.Format(ISO8601)
	if g.UnixFlag {
		timeFmt = strconv.FormatInt(expiry.Unix(), 10)
	} else if g.Rfc1123zFlag {
		timeFmt = expiry.Format(time.RFC1123Z)
	} else if g.Rfc3339Flag {
		timeFmt = expiry.Format(time.RFC3339)
	}

	// define a default output formatting
	output = fmt.Sprintf("%s\t%s", res.Name(), timeFmt)

	// redefine output formatting if a user requested
	// something besides the default values
	if g.BareFlag {
		output = timeFmt
	} else if g.JsonFlag {
		jsonStruct := map[string]string{
			"domainName": res.Name(),
			"expiry":     timeFmt,
		}

		jsonOut, err := json.MarshalIndent(jsonStruct, "", "  ")
		if err != nil {
			return output, err
		}

		output = string(jsonOut)
	}

	return output, err
}

type SubCommands struct {
	Domain      DomainCmd `cmd:"domain" help:"look up domain expiration date"`
	Certificate CertCmd   `cmd:"certificate" help:"look up TLS certificate expiration date"`
}

var cli struct {
	Globals
	SubCommands
}

func main() {
	console.DebugVariable = "SPIRY_DEBUG"

	ctx := kong.Parse(&cli,
		kong.Name(name),
		kong.Description("TLS & WHOIS expiration date lookup"),
		kong.ConfigureHelp(kong.HelpOptions{
			FlagsLast:           true,
			NoExpandSubcommands: false,
			Compact:             true,
			Summary:             true,
		}),
		kong.UsageOnError(),
		kong.Vars{"version": strings.TrimSpace(versionMsg())},
	)

	console.Debug(fmt.Sprintf("config context: %+v\n", ctx))

	// Call the Run() method of the selected parsed command.
	err := ctx.Run(&cli.Globals)

	ctx.FatalIfErrorf(err)
	// if err != nil {
	// 	fmt.Fprintln(os.Stderr, err)
	// 	os.Exit(1)
	// }
}
