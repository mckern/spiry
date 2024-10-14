# Spiry

A simple tool for checking domain & TLS certificate expiration dates

## Usage

### Top-Level Usage

These options are common to all `spiry` subcommands:

```text
$ spiry -h
Usage: spiry <command>

TLS & WHOIS expiration date lookup

Commands:
  domain         look up domain expiration date
  certificate    look up TLS certificate expiration date

Flags:
  -h, --help        Show context-sensitive help.
  -D, --debug       Enable debug mode
  -v, --version     display version information and exit
  -b, --bare        only display expiration date
  -j, --json        display output as JSON
  -u, --unix        display expiration date as UNIX timestamp
  -r, --rfc1123z    display expiration date as RFC1123Z timestamp
  -R, --rfc3339     display expiration date as RFC3339 timestamp

Run "spiry <command> --help" for more information on a command.
```

### Domain Lookup Usage

```text
$ spiry domain -h
Usage: spiry domain <domain>

look up domain expiration date

Arguments:
  <domain>    top-level domain name to look up

Flags:
  -h, --help             Show context-sensitive help.
  -D, --debug            Enable debug mode
  -v, --version          display version information and exit
  -b, --bare             only display expiration date
  -j, --json             display output as JSON
  -u, --unix             display expiration date as UNIX timestamp
  -r, --rfc1123z         display expiration date as RFC1123Z timestamp
  -R, --rfc3339          display expiration date as RFC3339 timestamp

  -s, --server=STRING    use <server> as specific whois server
```

### Certificate Lookup Usage

```text
$ spiry certificate -h
Usage: spiry certificate <address>

look up TLS certificate expiration date

Arguments:
  <address>    address to retrieve TLS certificate from

Flags:
  -h, --help           Show context-sensitive help.
  -D, --debug          Enable debug mode
  -v, --version        display version information and exit
  -b, --bare           only display expiration date
  -j, --json           display output as JSON
  -u, --unix           display expiration date as UNIX timestamp
  -r, --rfc1123z       display expiration date as RFC1123Z timestamp
  -R, --rfc3339        display expiration date as RFC3339 timestamp

  -n, --name=STRING    request TLS certificate for domain <name> instead of <address>
```

## Outputs & Examples

Command output is straightforward:

```text
$ spiry domain --bare --unix example.it
1696896000

$ spiry domain --bare example.it
2023-10-10T00:00:00+0000

$ spiry domain --json example.com
{
  "domainName": "example.com",
  "expiry": "2023-08-13T04:00:00+0000"
}

$ spiry domain --json --rfc3339 example.net
{
  "domainName": "example.net",
  "expiry": "2023-08-30T04:00:00Z"
}

$ spiry domain example.net
example.net	2023-08-30T04:00:00+0000

$ spiry domain "example.shh"
spiry: error: unable to find eTLD for domain example.shh: eTLD root "example.shh" is not publicly managed and cannot be looked up using `whois`

$ spiry domain "example.horse"
spiry: error: reserved domain record "example.horse" cannot be looked up
```

### Error handling

Error messages are emitted in plaintext format to standard error. If the errors are generated during flag parsing, the
help text will be emitted to standard error. WHOIS lookup errors will be emitted to standard error, without the help
text.

```text
$ ./build/spiry -j -b -u -r domain example.com 1>/dev/null
spiry: error: --bare and --json can't be used together

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
$ ./build/spiry no-such-example.com 1>/dev/null
ERROR: domain record "no-such-example.com" not found
```

## Caveats

`spiry` is provided as-is, with no assertions to correctness or completeness at this time.

Functionally, `spiry` is "fine" but it's a work in progress and like all moving targets, some of the interfaces and
options may change as it grows. Once a proper versioned release is cut, this warning will likely be removed.

## TO DO

- [ ] Move to GitHub Actions
- [ ] User-definable output, `printf` style?
  - Y'all like writing parsers? Because this will probably require a small parser.

## License

`spiry` is licensed under the terms of the [MIT License](LICENSE.txt).
