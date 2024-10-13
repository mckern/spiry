package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"github.com/araddon/dateparse"

	"github.com/mckern/spiry/internal/certificate"
	"github.com/mckern/spiry/internal/domain"
	"github.com/mckern/spiry/internal/spiry"
)

// Basic information about `spiry` itself
const (
	name = "spiry"
	url  = "https://github.com/mckern/spiry"
)

var (
	// metadata about the program itself, derived from compile-time variables
	buildDate     string
	gitCommit     string
	versionNumber string
)

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

var cli struct {
	spiry.Command
	Domain      domain.Command      `cmd:"domain" help:"look up domain expiration date"`
	Certificate certificate.Command `cmd:"certificate" help:"look up TLS certificate expiration date"`
}

func main() {
	ctx := kong.Parse(&cli,
		kong.Name(name),
		kong.Description("TLS & WHOIS expiration date lookup"),
		kong.ConfigureHelp(kong.HelpOptions{
			FlagsLast:           true,
			NoExpandSubcommands: true,
			Compact:             true,
			Summary:             true,
		}),
		kong.UsageOnError(),
		kong.Vars{"version": strings.TrimSpace(versionMsg())},
	)

	if cli.Command.Debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	// Call the Run() method of the selected parsed command.
	err := ctx.Run(&cli.Command)

	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Unwrap(err))
	}
}
