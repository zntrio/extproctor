# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Calendar Versioning](https://calver.org/).

## [v2025.12-2](https://github.com/zntrio/extproctor/releases/tag/v2025.12-2) - 2025-12-01

### Added

- **CLI Commands**
  - `run` - Execute ExtProc tests against a target service
    - Support for TCP (`--target`) and Unix domain socket (`--unix-socket`) connections
    - TLS support with client certificates (`--tls`, `--tls-cert`, `--tls-key`, `--tls-ca`)
    - Parallel test execution (`--parallel`)
    - Test filtering by name pattern (`--filter`) and tags (`--tags`)
    - Golden file update mode (`--update-golden`)
    - Multiple output formats: human-readable and JSON (`--output`)
    - Verbose mode for debugging (`--verbose`)
  - `validate` - Validate manifest files without running tests
  - `fmt` - Format textproto manifest files using txtpbfmt
    - In-place formatting (`--write`)
    - Diff preview mode (`--diff`)

- **Test Manifest Format**
  - Prototext-based test manifests for human-readable test definitions
  - Support for all ExtProc processing phases:
    - `REQUEST_HEADERS`
    - `REQUEST_BODY`
    - `REQUEST_TRAILERS`
    - `RESPONSE_HEADERS`
    - `RESPONSE_BODY`
    - `RESPONSE_TRAILERS`
  - Test case metadata: name, description, and tags
  - HTTP request specification: method, path, scheme, authority, headers, body, trailers

- **Expectation Types**
  - Headers response expectations (set, remove, append headers)
  - Body response expectations (body replacement, clear body)
  - Trailers response expectations (set, remove trailers)
  - Immediate response expectations (status code, headers, body, gRPC status)

- **Comparison Engine**
  - Unordered expectation matching - all expectations must be satisfied
  - Detailed diff reporting showing expected vs actual values
  - Support for partial matching (only specified fields are compared)

- **Golden File Support**
  - Snapshot testing with golden files
  - Automatic golden file generation with `--update-golden`
  - Golden files stored in prototext format for human review

- **Reporting**
  - Human-readable reporter with colored output
  - JSON reporter for CI/CD pipeline integration
  - Detailed failure messages with diff information
  - Test duration tracking and summary statistics

- **Sample ExtProc Server**
  - Reference implementation demonstrating ExtProc patterns
  - Useful for testing and learning the ExtProc API

- **Development**
  - Comprehensive test suite for core components
  - Protobuf code generation with Buf
  - Makefile with common development tasks

