# Spiry

A (trivially) simple domain expiration date checker tool

## Usage

```text
$ spiry mckern.sh
(2020-09-25 19:30:27 +0000 UTC) 375 days, 10 hours
$ spiry github.com
(2020-10-09 18:20:50 +0000 UTC) 389 days, 9 hours
$ spiry example.com
(2020-08-13 04:00:00 +0000 UTC) 332 days, 19 hours
```

## Caveats

- There's no tests for any of the code
- There's basically no tests for domain validity

  ```text
  $ spiry example.local
  dial tcp: lookup local.whois-servers.net on 1.1.1.1:53: no such host
  ```

- There's no standardized output format (yet)

## License

`spiry` is licensed under the terms of The Parity Public License, version 6.0.0.
