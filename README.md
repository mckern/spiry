# Spiry

A (trivially) simple domain expiration date checker tool

## Usage

```text
$ spiry example.com
(2020-08-13 04:00:00 +0000 UTC) 330 days, 19 hours
$ spiry google.dev
(2020-06-13 22:30:20 +0000 UTC) 270 days, 13 hours
$ spiry example.sh
(2020-08-13 10:56:01 +0000 UTC) 330 days, 2 hours
$ spiry example.it
(2019-10-03 00:00:00 +0000 UTC) 15 days, 15 hours
$ spiry example.horse
domain "example.horse" is not registered or has expired
```

## Caveats

- There's no tests for any of the code
  - there **is** an initial effort at extracting functionality into a package,
    which should make it easier to write domain resolution test cases.
- There's basically no tests for domain validity and if they work, they work

  ```text
  $ spiry example.local
  domain 'example.local' is unmanaged and cannot be looked up
  ```

- There's no standardized output format (yet)

## License

`spiry` is licensed under the terms of The Parity Public License, version 6.0.0.
