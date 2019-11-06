# Spiry

A (trivially) simple tool for checking domain expiration dates

## Usage

```text
$ spiry --help
spiry: print number of days until a domain name expires

usage: spiry [-h|-v|-b|-H] <domain>
  -b, --bare             display expiration date as ISO8601 timestamp
  -H, --human-readable   display a human-readable number of days until expiration
  -v, --version          display version information and exit
  -h, --help             display this help and exit

environment variables:
  SPIRY_DEBUG:   print debug messages
```

And the output is straightforward:

```text
$ spiry --human-readable "example.it"
example.it expires in 331 days, 0 hours
$ spiry --human-readable "example.com"
example.com expires in 280 days, 4 hours
$ spiry --human-readable "example.dev"
ERROR: canonical whois server "whois.nic.google" reports domain "example.dev" as unregistered
$ spiry --human-readable "example.co.uk"
example.co.uk expires in -106751 days, -23 hours
$ spiry "example.com"
example.com	2020-08-13T04:00:00+0000
$ spiry "example.dev"
ERROR: canonical whois server "whois.nic.google" reports domain "example.dev" as unregistered
$ spiry "example.sh"
ERROR: canonical whois server "whois.nic.sh" reports domain "example.sh" as unregistered
$ spiry "example.it"
example.it	2020-10-03T00:00:00+0000
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

`spiry` is licensed under the terms of The Parity Public License, version 6.0.0.
