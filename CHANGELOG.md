# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-09-08

### Changed

- Moved the package from `pkg/configloader` to the module root so it imports as `github.com/go-odies/configloader`.

### Added

- Initial release of the Go config loader library.
- Added a generic config loading flow that reads values from a config file, applies environment overrides, and then applies CLI argument overrides in precedence order.
- Added support for loading JSON and YAML configuration files with automatic type detection based on the file extension.
- Added environment variable mapping for nested configuration keys using prefixes and separators, including support for tag-based and camelCase field matching.
- Added command-line flag parsing for nested config keys such as `--database.host` and `--database.port`.
- Added typed value conversion for strings, booleans, integers, unsigned integers, and floating-point values.
- Added integration and unit tests covering file loading, nested environment overrides, tag handling, and CLI precedence behavior.
- Added initial documentation describing the library’s precedence model and example usage.

## [0.1.0] - 2026-09-07

### Added

- Initial project scaffolding for a Go library.
- Containerized development workflow with Docker and VS Code Dev Containers.
- Build, test, formatting, and vet automation via Make targets.
