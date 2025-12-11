# Go-based Prototext Test Runner for Envoy ExtProc Implementations

## Context

This project is a Go-based test runner designed for validating Envoy External Processing (ExtProc) filter implementations. The main goal is to provide a repeatable, automated way to verify that a given ExtProc service behaves as expected when handling HTTP requests encoded using Envoy’s ExtProc protobuf API.

The tool will:
- Read test manifests defined using protobuf messages encoded in Prototext (via protobuf-go’s prototext encoding).
- Use the Envoy ExtProc protobuf definitions (`envoy/extensions/filters/http/ext_proc/v3/ext_proc.proto`) as the schema for the messages.
- Construct and send requests (headers, body, and any relevant ExtProc messages) to a target Envoy ExtProc service.
- Receive the ExtProc service’s responses and compare them against expectations defined in the manifest.

This matters because implementing and evolving ExtProc services is error-prone: behaviors depend on correct protobuf message structures and a sequence of callbacks. A dedicated test runner allows:
- Automated regression testing for ExtProc implementations.
- Easier local development and debugging of filter behavior.
- Integration into CI pipelines for consistent validation of behavior across changes.

Assumptions (to be validated/clarified):
- The “prototest manifest” is itself a protobuf-defined message (or set of messages) that describes:
  - Input HTTP request data (headers, body, possibly method, path, scheme, etc.).
  - Expected ExtProc interactions and responses (e.g., `HeadersResponse`, `BodyResponse`, `TrailersResponse`, status, mutations, etc.).
  - Any test metadata (name, description, tags).
- The ExtProc service is exposed over gRPC and follows the Envoy ExtProc API semantics.
- The tool will be primarily CLI-based, suitable for local runs and CI integration.

## Constraints

### Functional Requirements (FR)

- Implement the test runner in Go.
- Use the Envoy ExtProc protobuf definitions from `envoy/extensions/filters/http/ext_proc/v3/ext_proc.proto` to define request/response message types.
- Use protobuf-go’s Prototext encoding/decoding (`encoding/prototext`) for reading and writing protobuf messages.
- Define and support a “prototest manifest” format (protobuf-based) that includes at minimum:
  - Test case metadata (name, description, optional tags).
  - HTTP request specification: method, URL/path, headers, body, and any relevant attributes.
  - Expected ExtProc responses and/or mutations (e.g., expected headers modifications, body modifications, status changes, immediate responses, etc.).
- Provide a mechanism to load one or more test manifests from files or directories (e.g., `.textproto` or `.prototext` files).
- For each test case in the manifest:
  - Parse the manifest using Prototext into the corresponding protobuf structures.
  - Construct the appropriate ExtProc request messages that Envoy would send (e.g., `ProcessingRequest` with `request_headers`, `request_body`, etc.) based on the manifest’s HTTP request definition.
  - Establish a connection (likely gRPC) to the target ExtProc service endpoint (configured via CLI flags or configuration file).
  - Send the sequence of ExtProc requests (headers, body, trailers as needed) that correspond to the test input.
  - Receive and record all ExtProc responses from the service.
  - Compare actual ExtProc responses to expected responses defined in the manifest (field-by-field comparison, with support for exact match; possibly configurable tolerance or partial match rules).
- Produce a test result per test case including:
  - Pass/Fail status.
  - Summary of mismatches for failed tests (e.g., which field/path differs, expected vs actual values).
- Provide a command-line interface (CLI) with options such as:
  - Path(s) to manifest file(s) or directories.
  - ExtProc service address and port.
  - Output format (human-readable, and optionally machine-readable such as JSON for CI integration).
  - Verbosity level for logging.
- Return a non-zero process exit code if any test fails, to support CI integration.
- Optionally support:
  - Multiple test cases in a single manifest file.
  - Grouping or filtering tests to run by name or tag.
  - Snapshot or golden-file style expected response definitions (stored as prototext) and automatic diffing when responses change.


### Non-functional Requirements (NFR)

- Performance
  - Handle a reasonable number of test cases in a single run without significant slowdown (e.g., hundreds to low thousands of tests).
  - Minimize overhead in establishing connections to the ExtProc service (e.g., connection reuse where appropriate).
- Scalability
  - Support running tests in parallel (e.g., multiple test cases concurrently) when talking to ExtProc services that can handle concurrency.
  - Be able to extend to more complex scenarios (multiple sequences of ExtProc callbacks per test) without significant rework.
- Reliability
  - Provide clear error messages when manifests are invalid, the ExtProc service is unreachable, or protobuf decoding fails.
  - Fail gracefully and continue running remaining tests when individual tests encounter errors, where possible.
- Maintainability
  - Structure the Go code with clear separation between:
    - Manifest parsing.
    - ExtProc client interaction.
    - Comparison/assertion logic.
    - CLI and reporting.
  - Use generated Go code from the official Envoy ExtProc protobuf definition and automate regeneration via tooling.
  - Include tests for core components (e.g., manifest parsing, comparison logic).
- Usability
  - Provide a simple, well-documented CLI interface.
  - Offer readable test output summaries, including concise per-test results and an overall summary.
  - Optionally support a verbose mode that logs the exact protobuf messages sent and received (in Prototext or JSON) for debugging.
- Compatibility
  - Be compatible with modern Go versions (e.g., Go 1.20+ or current LTS at time of development).
  - Track and remain compatible with updates to Envoy’s ExtProc protobuf definition where feasible.
- Security
  - Avoid executing arbitrary code from manifests (treat them as data only).
  - Use secure defaults for network connections (e.g., options for TLS if needed in target environments).


## Target Users

- Backend/Platform Engineers implementing Envoy ExtProc services
  - Need a reliable, automated way to validate that their ExtProc logic behaves correctly for a variety of HTTP request scenarios.
  - Want to run fast feedback loops locally during development.

- SRE/DevOps/Platform Teams maintaining Envoy-based infrastructure
  - Need a tool that can be integrated into CI/CD pipelines to ensure ExtProc changes do not introduce regressions.
  - Require machine-readable test outputs and clear pass/fail signals.

- QA/Testing Engineers working on traffic routing and policy enforcement
  - Need to express complex test cases in a structured, version-controlled manifest format.
  - Want detailed diffs when behavior deviates from expected ExtProc responses.

- Open-source contributors to Envoy-related ExtProc libraries or services
  - Need a standardized test runner aligned with Envoy’s ExtProc protobuf API for cross-project consistency.
