package spiry

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/alecthomas/kong"
)

const ISO8601 = "2006-01-02T15:04:05-0700"

type Command struct {
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

func (g *Command) Render(res ExpiringResource) (output string, err error) {
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
