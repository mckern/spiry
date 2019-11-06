# Spiry

A (trivially) simple tool for checking domain expiration dates

## Usage

```text
$ spiry --help
spiry: print number of days until a domain name expires

usage: spiry [-h|-v|-b|-H] <domain>
  -b, --bare             display the bare expiration date in some mish-mash unix format that might be RFCish?
  -H, --human-readable   print the human-readable number of days until expiration
  -v, --version          display version information and exit
  -h, --help             display this help and exit

environment variables:
  SPIRY_DEBUG:   print debug messages
```

And the output is straightforward:

```text
$ spiry --human-readable "example.com"
example.com expires in 280 days, 19 hours
$ spiry "example.dev"
ERROR: canonical whois server "whois.nic.google" reports domain "example.dev" as unregistered
$ spiry "example.sh"
ERROR: canonical whois server "whois.nic.sh" reports domain "example.sh" as unregistered
$ spiry "example.it"
example.it	Sat Oct  3 00:00:00 UTC 2020
$ spiry --bare "example.it"
Sat Oct  3 00:00:00 UTC 2020
$ spiry "example.horse"
ERROR: canonical whois server "whois.nic.horse" reports domain "example.horse" as unregistered
$ spiry "google.co.uk"
google.co.uk	Fri Feb 14 00:00:00 UTC 2020
$ spiry "google.dev"
google.dev	Sat Jun 13 22:30:20 UTC 2020
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
