<div align="center">

# 🧪 ExtProctor

**A test runner for Envoy ExtProc implementations**

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/zntr.io/extproctor)](https://goreportcard.com/report/zntr.io/extproctor)
[![Documentation](https://img.shields.io/badge/docs-reference-blue)](https://pkg.go.dev/zntr.io/extproctor)

[Features](#features) •
[Installation](#installation) •
[Quick Start](#quick-start) •
[Documentation](#documentation) •
[Examples](#examples) •
[Contributing](#contributing)

</div>

---

## Why ExtProctor?

Implementing and evolving [Envoy External Processing (ExtProc)](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/ext_proc_filter) services is error-prone. Behaviours depend on correct protobuf message structures and a sequence of callbacks that are difficult to test manually.

> [!NOTE]
> This project is heavily inspired by [Google Service Extensions](https://github.com/GoogleCloudPlatform/service-extensions) where they use the same approach to test their service extensions.

**ExtProctor** provides a dedicated test runner that enables:

- ✅ **Automated regression testing** for ExtProc implementations
- ✅ **Fast feedback loops** during local development
- ✅ **CI/CD integration** with machine-readable outputs
- ✅ **Version-controlled test cases** using human-readable prototext manifests

## Features

| Feature | Description |
|---------|-------------|
| 📝 **Prototext Manifests** | Define test cases using human-readable prototext format |
| 🔄 **Full ExtProc Support** | Test all processing phases: headers, body, and trailers |
| 📸 **Golden Files** | Capture and compare responses using the golden file pattern |
| ⚡ **Parallel Execution** | Run tests concurrently for faster feedback |
| 🏷️ **Flexible Filtering** | Filter tests by name pattern or tags |
| 📊 **Multiple Output Formats** | Human-readable or JSON output for CI integration |
| 🔌 **Unix Socket Support** | Connect to ExtProc services via Unix domain sockets |
| 🔒 **TLS Support** | Secure gRPC connections with client certificates |

## Installation

### Using Go

```bash
go install zntr.io/extproctor/cmd/extproctor@latest
```

### From Source

```bash
git clone https://github.com/zntrio/extproctor.git
cd extproctor
go build -o extproctor ./cmd/extproctor
```

### Verify Installation

```bash
extproctor --help
```

## Quick Start

### 1. Create a Test Manifest

Create a file `tests/basic.textproto`:

```prototext
name: "basic-test"
description: "Basic ExtProc test"

test_cases: {
  name: "add-header"
  description: "Verify ExtProc adds a custom header"
  tags: ["smoke"]

  request: {
    method: "GET"
    path: "/api/v1/users"
    scheme: "https"
    authority: "api.example.com"
    headers: {
      key: "content-type"
      value: "application/json"
    }
  }

  expectations: {
    phase: REQUEST_HEADERS
    headers_response: {
      set_headers: {
        key: "x-custom-header"
        value: "custom-value"
      }
    }
  }
}
```

### 2. Run the Tests

```bash
extproctor run ./tests/ --target localhost:50051
```

### 3. View Results

```
Running tests from 1 manifest(s)...

✓ basic-test/add-header (12ms)

Results: 1 passed, 0 failed, 0 skipped
```

## Documentation

### CLI Commands

#### `extproctor run`

Execute tests against an ExtProc service.

```bash
# Run all tests in a directory
extproctor run ./tests/ --target localhost:50051

# Run with Unix domain socket
extproctor run ./tests/ --unix-socket /var/run/extproc.sock

# Run with parallel execution
extproctor run ./tests/ --target localhost:50051 --parallel 4

# Filter by test name pattern
extproctor run ./tests/ --target localhost:50051 --filter "auth*"

# Filter by tags
extproctor run ./tests/ --target localhost:50051 --tags "smoke,regression"

# JSON output for CI pipelines
extproctor run ./tests/ --target localhost:50051 --output json

# Verbose mode for debugging
extproctor run ./tests/ --target localhost:50051 -v

# Update golden files
extproctor run ./tests/ --target localhost:50051 --update-golden
```

#### `extproctor validate`

Validate manifest syntax without running tests.

```bash
# Validate all manifests in a directory
extproctor validate ./tests/

# Validate specific files
extproctor validate test1.textproto test2.textproto
```

#### `extproctor fmt`

Format textproto manifest files using [txtpbfmt](https://github.com/protocolbuffers/txtpbfmt).

```bash
# Format a single file to stdout
extproctor fmt test.textproto

# Format files in-place
extproctor fmt --write ./tests/

# Show diff of what would change
extproctor fmt --diff ./tests/

# Format specific files in-place
extproctor fmt -w test1.textproto test2.textproto

# CI check - returns error if files need formatting
extproctor fmt ./tests/
```

### Command-Line Options

#### Run Command Options

| Flag | Description | Default |
|------|-------------|---------|
| `--target` | ExtProc service address (host:port) | `localhost:50051` |
| `--unix-socket` | Unix domain socket path | — |
| `--tls` | Enable TLS for gRPC connection | `false` |
| `--tls-cert` | TLS client certificate file | — |
| `--tls-key` | TLS client key file | — |
| `--tls-ca` | TLS CA certificate file | — |
| `-p, --parallel` | Number of parallel test executions | `1` |
| `-o, --output` | Output format (`human`, `json`) | `human` |
| `-v, --verbose` | Enable verbose output | `false` |
| `--filter` | Filter tests by name pattern | — |
| `--tags` | Filter tests by tags (comma-separated) | — |
| `--update-golden` | Update golden files with actual responses | `false` |

> **Note:** `--target` and `--unix-socket` are mutually exclusive.

#### Fmt Command Options

| Flag | Description | Default |
|------|-------------|---------|
| `-w, --write` | Write formatted output back to files (in-place) | `false` |
| `-d, --diff` | Show diff of what would change | `false` |

### Manifest Format

Test manifests are written in [Prototext](https://protobuf.dev/reference/protobuf/textformat-spec/) format.

#### Structure

```prototext
name: "manifest-name"
description: "Description of the test suite"

test_cases: {
  name: "test-case-name"
  description: "What this test validates"
  tags: ["tag1", "tag2"]

  request: {
    method: "POST"
    path: "/api/endpoint"
    scheme: "https"
    authority: "api.example.com"
    headers: {
      key: "content-type"
      value: "application/json"
    }
    body: '{"key": "value"}'
    process_request_body: true
    process_response_headers: true
  }

  expectations: {
    phase: REQUEST_HEADERS
    headers_response: {
      set_headers: {
        key: "x-custom"
        value: "value"
      }
    }
  }
}
```

#### Processing Phases

| Phase | Description |
|-------|-------------|
| `REQUEST_HEADERS` | Processing request headers |
| `REQUEST_BODY` | Processing request body |
| `REQUEST_TRAILERS` | Processing request trailers |
| `RESPONSE_HEADERS` | Processing response headers |
| `RESPONSE_BODY` | Processing response body |
| `RESPONSE_TRAILERS` | Processing response trailers |

#### Expectation Types

<details>
<summary><strong>Headers Response</strong></summary>

```prototext
expectations: {
  phase: REQUEST_HEADERS
  headers_response: {
    set_headers: {
      key: "x-custom"
      value: "value"
    }
    remove_headers: "x-internal"
    append_headers: {
      key: "x-multi"
      value: "value"
    }
  }
}
```

</details>

<details>
<summary><strong>Body Response</strong></summary>

```prototext
expectations: {
  phase: REQUEST_BODY
  body_response: {
    body: '{"modified": true}'
    common_response: {
      status: CONTINUE_AND_REPLACE
    }
  }
}
```

</details>

<details>
<summary><strong>Trailers Response</strong></summary>

```prototext
expectations: {
  phase: REQUEST_TRAILERS
  trailers_response: {
    set_trailers: {
      key: "x-checksum-validated"
      value: "true"
    }
  }
}
```

</details>

<details>
<summary><strong>Immediate Response (Short-circuit)</strong></summary>

```prototext
expectations: {
  phase: REQUEST_HEADERS
  immediate_response: {
    status_code: 403
    headers: {
      key: "content-type"
      value: "application/json"
    }
    body: '{"error": "forbidden"}'
  }
}
```

</details>

#### Golden Files

Use golden files for snapshot testing:

```prototext
test_cases: {
  name: "golden-test"
  request: { ... }
  golden_file: "golden/test-response.textproto"
}
```

Update golden files when behavior changes intentionally:

```bash
extproctor run ./tests/ --target localhost:50051 --update-golden
```

## Examples

The [`testdata/examples/`](testdata/examples) directory contains complete example manifests:

| File | Description |
|------|-------------|
| [`basic_headers.textproto`](testdata/examples/basic_headers.textproto) | Basic header processing (add/remove headers) |
| [`auth_flow.textproto`](testdata/examples/auth_flow.textproto) | Authentication flow with immediate response rejection |
| [`body_processing.textproto`](testdata/examples/body_processing.textproto) | Request body inspection and transformation |
| [`multi_phase_flow.textproto`](testdata/examples/multi_phase_flow.textproto) | Multi-phase processing across request/response lifecycle |

### Sample ExtProc Server

A sample ExtProc server is included for testing and reference:

```bash
# Start the sample server
go run ./sample/extproc/ --addr :50051

# Run tests against it
extproctor run ./sample/extproc/test/ --target localhost:50051
```

The sample server demonstrates:
- Request headers processing with custom header injection
- Request/response body handling
- Response headers modification
- gRPC health check endpoint

## Development

### Prerequisites

- [Go 1.24+](https://go.dev/dl/)
- [Buf CLI](https://buf.build/docs/installation) (for protobuf generation)

### Building

```bash
go build -o extproctor ./cmd/extproctor
```

### Running Tests

```bash
go test ./...
```

### Regenerating Protobuf Code

```bash
buf generate
```

### Project Structure

```
extproctor/
├── cmd/extproctor/          # CLI entry point
├── internal/
│   ├── cli/              # Command-line interface
│   ├── client/           # ExtProc gRPC client
│   ├── comparator/       # Response comparison logic
│   ├── golden/           # Golden file handling
│   ├── manifest/         # Manifest loading and validation
│   ├── reporter/         # Test result reporting
│   └── runner/           # Test execution engine
├── proto/                # Protobuf definitions
├── sample/extproc/       # Sample ExtProc server
└── testdata/examples/    # Example test manifests
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please make sure to:
- Update tests as appropriate
- Follow the existing code style
- Update documentation for any new features

## Related Resources

- [Envoy ExtProc Documentation](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/ext_proc_filter)
- [Envoy ExtProc Proto Definition](https://github.com/envoyproxy/envoy/blob/main/api/envoy/service/ext_proc/v3/external_processor.proto)
- [Prototext Format Specification](https://protobuf.dev/reference/protobuf/textformat-spec/)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

<div align="center">

Made with ❤️ for the Envoy community

</div>
