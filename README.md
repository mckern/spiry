# Spiry

A (trivially) simple tool for checking domain expiration dates

## Usage

```text
$ spiry -h
spiry: look up domain name expiration

usage: spiry [-h|-v] [-b|-j] [-u|-r|-R] <domain>
  -b, --bare       only display expiration date
  -j, --json       display output as JSON
  -u, --unix       display expiration date as UNIX time
  -r, --rfc1123z   display expiration date as RFC1123Z timestamp
  -R, --rfc3339    display expiration date as RFC3339 timestamp
  -v, --version    display version information and exit
  -h, --help       display this help and exit

environment variables:
  SPIRY_DEBUG:   print debug messages
```

And the output is straightforward:

```text
$ spiry --bare --unix example.it
1601683200
$ spiry --bare example.it
2020-10-03T00:00:00+0000
$ spiry --json mckern.sh
{
  "domain": "mckern.sh",
  "expiry": "2020-09-25T19:30:27+0000"
}
$ spiry --json --rfc3339 mckern.dev
{
  "domain": "mckern.dev",
  "expiry": "2021-02-28T16:00:11Z"
}
$ spiry github.com
github.com	2020-10-09T18:20:50+0000
$ spiry "example.sh"
ERROR: canonical whois server "whois.nic.sh" reports domain "example.sh" as unregistered
$ spiry "example.horse"
ERROR: canonical whois server "whois.nic.horse" reports domain "example.horse" as unregistered
$ spiry "example.co.uk"
example.co.uk	0001-01-01T00:00:00+0000
$ spiry "google.dev"
google.dev	2020-06-13T22:30:20+0000
```

## Caveats

- There's no tests for any of the code **yet**
  - there **is** an initial effort at extracting functionality into a package,
    which should make it easier to write domain resolution test cases.

## Some ideas for future features

- SSL certificate expiry checking
- user-definable output

## License

`spiry` is licensed under the terms of [The Parity Public License](https://github.com/mckern/spiry/blob/af49ce3c641796d700d2269d46ada24bcfb7c33b/LICENSE.md), version 7.0.0.
