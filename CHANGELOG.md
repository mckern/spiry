# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.3.0](https://github.com/mckern/spiry/compare/v0.2.1...v0.3.0) - released 2024-10-14

v0.3.0 is a breaking change from v0.2.x. It introduces subcommands, and involved
considerable refactoring under the hood. `spiry` now supports getting TLS certificate
expiration dates from HTTPS URLs.

### Added

- Support for checking & displaying TLS expiration dates from HTTPS addresses
  using the `spiry certificate` subcommand
- This [Changelog](./CHANGELOG.md)
- `spiry` now ships precompiled binaries for major platforms
  - macOS builds are signed & notarized [fat binaries](https://developer.apple.com/documentation/apple-silicon/building-a-universal-macos-binary)

### Changed

- Checking & displaying domain registration expiration dates has been moved into
  the `spiry domain` subcommand
- Upgraded [`likexian/whois-parser`][whois-parser] to v1.24.20
- Upgraded [`likexian/whois`][whois] to v1.15.5
- Upgrade Go API from 1.16 to 1.22
- Build pre-compiled binaries with Go 1.23
- Migrate CI from Drone to GitHub Actions
- Migrate compilation tooling from [GNU make](https://www.gnu.org/software/make/) to [Goreleaser](https://goreleaser.com)

### Removed

- Dependencies are no longer vendored into this repo
- Coinciding with the shift to Goreleaser, `Makefile` has been removed
- Internal `console` package has been replaced with [`slog`](https://pkg.go.dev/log/slog)

## [0.2.1](https://github.com/mckern/spiry/compare/v0.2.0...v0.2.1) - 2021-08-17

### Changed

- Quick Go module cleanup

## [0.2.0](https://github.com/mckern/spiry/compare/v0.1.1...v0.2.0) - 2021-08-17

v0.2.0 relicenses the `spiry` codebase under the terms of the MIT license.

### Added

- Initial (mediocre) tests

### Changed

- Upgraded [`likexian/whois-parser`][whois-parser] to v1.20.4
- Upgraded [`likexian/whois`][whois] to v1.12.1
- Upgrade Go API from 1.13 to 1.16
- Revamped command line flags to support more flexible output
- Relicensed under the [MIT License](./LICENSE.txt)

[whois-parser]: https://github.com/likexian/whois-parser/
[whois]: https://github.com/likexian/whois/
