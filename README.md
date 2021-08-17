# Spiry

A (trivially) simple tool for checking domain expiration dates

[![Build Status](https://ci.home.mckern.sh/api/badges/mckern/spiry/status.svg)](https://ci.home.mckern.sh/mckern/spiry)

## Usage

```text
$ spiry -h
spiry: look up domain name expiration

usage: spiry [-b|-j] [-u|-r|-R] [-h|-v] [-s <server>] <domain>
  -s, --server string   use <server> as specific whois server
  -b, --bare            only display expiration date
  -j, --json            display output as JSON
  -u, --unix            display expiration date as UNIX time
  -r, --rfc1123z        display expiration date as RFC1123Z timestamp
  -R, --rfc3339         display expiration date as RFC3339 timestamp
  -v, --version         display version information and exit
  -h, --help            display this help and exit

environment variables:
  SPIRY_DEBUG:   print debug messages
```

And the output is straightforward:

```text
$ spiry --bare --unix example.it
1633824000
$ spiry --bare example.it
2021-10-10T00:00:00+0000
$ spiry --json example.com
{
  "domain": "example.com",
  "expiry": "0001-01-01T00:00:00+0000"
}
$ spiry --json --rfc3339 example.net
{
  "domain": "example.net",
  "expiry": "0001-01-01T00:00:00Z"
}
$ spiry example.net
example.net	0001-01-01T00:00:00+0000
$ spiry "example.shh"
ERROR: unable to find eTLD for domain example.shh: eTLD root "example.shh" is not publicly managed and cannot be looked up using `whois`
$ spiry "example.horse"
ERROR: whois reports domain "example.horse" as unregistered or expired
```

## Caveats

`spiry` is provided as-is, with no assertions to correctness or completeness at this time.

Functionally, `spiry` is "fine" but it's a work in progress and like all moving targets,
some of the interfaces and options may change as it grows. Once a proper versioned release is cut,
this warning will likely be removed.

## TO DO

- [ ] Move to Kong https://github.com/alecthomas/kong
  - [pflag](https://github.com/spf13/pflag) has been OK but it's showing its age in places
- [ ] SSL certificate expiry checking
  - Subcommand support may be easier post-Kong
- [ ] User-definable output, `printf` style.
  - Y'all like writing parsers? Because this will probably require a small parser.

## License

`spiry` is licensed under the terms of the [MIT License](LICENSE.txt).
