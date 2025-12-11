// SPDX-FileCopyrightText: 2025 Thibault NORMAND
// SPDX-License-Identifier: MIT

package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCmd_Basic(t *testing.T) {
	assert.NotNil(t, rootCmd)
	assert.Equal(t, "extproctor", rootCmd.Use)
}

func TestRootCmd_HasFlags(t *testing.T) {
	flags := rootCmd.PersistentFlags()
	assert.NotNil(t, flags)

	// Check target flag
	f := flags.Lookup("target")
	assert.NotNil(t, f)
	assert.Equal(t, "localhost:50051", f.DefValue)

	// Check unix-socket flag
	f = flags.Lookup("unix-socket")
	assert.NotNil(t, f)

	// Check TLS flags
	f = flags.Lookup("tls")
	assert.NotNil(t, f)
	f = flags.Lookup("tls-cert")
	assert.NotNil(t, f)
	f = flags.Lookup("tls-key")
	assert.NotNil(t, f)
	f = flags.Lookup("tls-ca")
	assert.NotNil(t, f)

	// Check execution flags
	f = flags.Lookup("parallel")
	assert.NotNil(t, f)
	assert.Equal(t, "1", f.DefValue)

	f = flags.Lookup("output")
	assert.NotNil(t, f)
	assert.Equal(t, "human", f.DefValue)

	f = flags.Lookup("verbose")
	assert.NotNil(t, f)

	// Check filtering flags
	f = flags.Lookup("filter")
	assert.NotNil(t, f)

	f = flags.Lookup("tags")
	assert.NotNil(t, f)
}

func TestRunCmd_Basic(t *testing.T) {
	assert.NotNil(t, runCmd)
	assert.Equal(t, "run [paths...]", runCmd.Use)
}

func TestRunCmd_HasUpdateGoldenFlag(t *testing.T) {
	f := runCmd.Flags().Lookup("update-golden")
	assert.NotNil(t, f)
	assert.Equal(t, "false", f.DefValue)
}

func TestValidateCmd_Basic(t *testing.T) {
	assert.NotNil(t, validateCmd)
	assert.Equal(t, "validate [paths...]", validateCmd.Use)
}

func TestValidateManifests_Success(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "test.textproto")

	content := `
name: "test-manifest"
test_cases: {
  name: "test-1"
  request: { method: "GET", path: "/" }
  expectations: { phase: REQUEST_HEADERS, headers_response: {} }
}
`
	err := os.WriteFile(manifestPath, []byte(content), 0o644)
	require.NoError(t, err)

	// Create a test command
	cmd := &cobra.Command{}

	err = validateManifests(cmd, []string{tmpDir})
	assert.NoError(t, err)
}

func TestValidateManifests_InvalidManifest(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "invalid.textproto")

	err := os.WriteFile(manifestPath, []byte("invalid { data"), 0o644)
	require.NoError(t, err)

	cmd := &cobra.Command{}

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	err = validateManifests(cmd, []string{tmpDir})

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stderr = oldStderr

	assert.Error(t, err)
	assert.Contains(t, buf.String(), "ERROR")
}

func TestValidateManifests_InvalidTestCase(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "test.textproto")

	// Create a manifest with invalid test case (missing name)
	content := `
name: "test-manifest"
test_cases: {
  request: { method: "GET", path: "/" }
  expectations: { phase: REQUEST_HEADERS, headers_response: {} }
}
`
	err := os.WriteFile(manifestPath, []byte(content), 0o644)
	require.NoError(t, err)

	cmd := &cobra.Command{}

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	err = validateManifests(cmd, []string{tmpDir})

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stderr = oldStderr

	assert.Error(t, err)
	assert.Contains(t, buf.String(), "ERROR")
}

func TestValidateManifests_NonExistentPath(t *testing.T) {
	cmd := &cobra.Command{}

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	err := validateManifests(cmd, []string{"/nonexistent/path"})

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stderr = oldStderr

	assert.Error(t, err)
	assert.Contains(t, buf.String(), "ERROR")
}

func TestValidateManifests_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()

	content1 := `
name: "manifest-1"
test_cases: {
  name: "test-1"
  request: { method: "GET", path: "/" }
  expectations: { phase: REQUEST_HEADERS, headers_response: {} }
}
`
	content2 := `
name: "manifest-2"
test_cases: {
  name: "test-2"
  request: { method: "POST", path: "/api" }
  expectations: { phase: REQUEST_HEADERS, headers_response: {} }
}
`

	err := os.WriteFile(filepath.Join(tmpDir, "test1.textproto"), []byte(content1), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "test2.textproto"), []byte(content2), 0o644)
	require.NoError(t, err)

	cmd := &cobra.Command{}

	err = validateManifests(cmd, []string{tmpDir})
	assert.NoError(t, err)
}

func TestExecute_Basic(t *testing.T) {
	// Test that Execute returns without error when rootCmd is set up
	// We don't actually run it as it would try to parse command line args
	assert.NotNil(t, rootCmd)
}

func TestRunTests_NoManifests(t *testing.T) {
	tmpDir := t.TempDir()

	// Override global flags for this test
	oldTarget := target
	target = "localhost:50051"
	defer func() { target = oldTarget }()

	cmd := &cobra.Command{}

	err := runTests(cmd, []string{tmpDir})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no test manifests found")
}

func TestRunTests_InvalidPath(t *testing.T) {
	// Override global flags for this test
	oldTarget := target
	target = "localhost:50051"
	defer func() { target = oldTarget }()

	cmd := &cobra.Command{}

	err := runTests(cmd, []string{"/nonexistent/path"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load manifests")
}

func TestGlobalFlags_Defaults(t *testing.T) {
	// Test default values are set correctly in rootCmd
	flags := rootCmd.PersistentFlags()

	// Get the default values
	targetFlag := flags.Lookup("target")
	require.NotNil(t, targetFlag)
	assert.Equal(t, "localhost:50051", targetFlag.DefValue)

	parallelFlag := flags.Lookup("parallel")
	require.NotNil(t, parallelFlag)
	assert.Equal(t, "1", parallelFlag.DefValue)

	outputFlag := flags.Lookup("output")
	require.NotNil(t, outputFlag)
	assert.Equal(t, "human", outputFlag.DefValue)
}

func TestValidateCmd_HasSubcommand(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "validate" {
			found = true
			break
		}
	}
	assert.True(t, found, "validate command should be registered")
}

func TestRunCmd_HasSubcommand(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "run" {
			found = true
			break
		}
	}
	assert.True(t, found, "run command should be registered")
}

func TestValidateManifests_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := &cobra.Command{}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := validateManifests(cmd, []string{tmpDir})

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stdout = oldStdout

	// Empty directory should succeed with 0 manifests
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Validated 0 manifest(s)")
}

func TestValidateManifests_MultipleTestCases(t *testing.T) {
	tmpDir := t.TempDir()

	content := `
name: "test-manifest"
test_cases: {
  name: "test-1"
  request: { method: "GET", path: "/" }
  expectations: { phase: REQUEST_HEADERS, headers_response: {} }
}
test_cases: {
  name: "test-2"
  request: { method: "POST", path: "/api" }
  expectations: { phase: REQUEST_HEADERS, headers_response: {} }
}
`
	err := os.WriteFile(filepath.Join(tmpDir, "test.textproto"), []byte(content), 0o644)
	require.NoError(t, err)

	cmd := &cobra.Command{}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = validateManifests(cmd, []string{tmpDir})

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stdout = oldStdout

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "2 test case(s)")
}

func TestValidateManifests_PartialInvalid(t *testing.T) {
	tmpDir := t.TempDir()

	// Create one valid and one invalid manifest
	validContent := `
name: "valid-manifest"
test_cases: {
  name: "test-1"
  request: { method: "GET", path: "/" }
  expectations: { phase: REQUEST_HEADERS, headers_response: {} }
}
`
	invalidContent := `
name: "invalid-manifest"
test_cases: {
  request: { method: "GET", path: "/" }
  expectations: { phase: REQUEST_HEADERS, headers_response: {} }
}
`

	err := os.WriteFile(filepath.Join(tmpDir, "valid.textproto"), []byte(validContent), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "invalid.textproto"), []byte(invalidContent), 0o644)
	require.NoError(t, err)

	cmd := &cobra.Command{}

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	err = validateManifests(cmd, []string{tmpDir})

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stderr = oldStderr

	// Should fail because of invalid manifest
	assert.Error(t, err)
	assert.Contains(t, buf.String(), "ERROR")
}

func TestRootCmd_LongDescription(t *testing.T) {
	assert.NotEmpty(t, rootCmd.Long)
	assert.Contains(t, rootCmd.Long, "ExtProc")
}

func TestRunCmd_LongDescription(t *testing.T) {
	assert.NotEmpty(t, runCmd.Long)
	assert.Contains(t, runCmd.Long, "ExtProc")
}

func TestValidateCmd_LongDescription(t *testing.T) {
	assert.NotEmpty(t, validateCmd.Long)
	assert.Contains(t, validateCmd.Long, "prototext")
}

func TestRunTests_WithValidManifest_NoServer(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "test.textproto")

	content := `
name: "test-manifest"
test_cases: {
  name: "test-1"
  request: { method: "GET", path: "/" }
  expectations: { phase: REQUEST_HEADERS, headers_response: {} }
}
`
	err := os.WriteFile(manifestPath, []byte(content), 0o644)
	require.NoError(t, err)

	// Override global flags
	oldTarget := target
	oldOutput := output
	oldVerbose := verbose
	oldParallel := parallel

	target = "localhost:50051"
	output = "human"
	verbose = false
	parallel = 1

	defer func() {
		target = oldTarget
		output = oldOutput
		verbose = oldVerbose
		parallel = oldParallel
	}()

	cmd := &cobra.Command{}

	// This should fail because no server is running
	err = runTests(cmd, []string{tmpDir})

	// The exact error depends on whether connection is blocked or refused
	assert.Error(t, err)
}

func TestRunTests_JSONOutput(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "test.textproto")

	content := `
name: "test-manifest"
test_cases: {
  name: "test-1"
  request: { method: "GET", path: "/" }
  expectations: { phase: REQUEST_HEADERS, headers_response: {} }
}
`
	err := os.WriteFile(manifestPath, []byte(content), 0o644)
	require.NoError(t, err)

	// Override global flags
	oldTarget := target
	oldOutput := output

	target = "localhost:50051"
	output = "json"

	defer func() {
		target = oldTarget
		output = oldOutput
	}()

	cmd := &cobra.Command{}

	// Will fail but tests the json reporter path
	err = runTests(cmd, []string{tmpDir})
	assert.Error(t, err)
}

func TestRunTests_WithFilter(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "test.textproto")

	content := `
name: "test-manifest"
test_cases: {
  name: "test-1"
  request: { method: "GET", path: "/" }
  expectations: { phase: REQUEST_HEADERS, headers_response: {} }
}
`
	err := os.WriteFile(manifestPath, []byte(content), 0o644)
	require.NoError(t, err)

	// Override global flags
	oldTarget := target
	oldFilter := filter

	target = "localhost:50051"
	filter = "test-*"

	defer func() {
		target = oldTarget
		filter = oldFilter
	}()

	cmd := &cobra.Command{}

	err = runTests(cmd, []string{tmpDir})
	assert.Error(t, err)
}

func TestRunTests_WithTags(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "test.textproto")

	content := `
name: "test-manifest"
test_cases: {
  name: "test-1"
  tags: ["smoke"]
  request: { method: "GET", path: "/" }
  expectations: { phase: REQUEST_HEADERS, headers_response: {} }
}
`
	err := os.WriteFile(manifestPath, []byte(content), 0o644)
	require.NoError(t, err)

	// Override global flags
	oldTarget := target
	oldTags := tags

	target = "localhost:50051"
	tags = []string{"smoke"}

	defer func() {
		target = oldTarget
		tags = oldTags
	}()

	cmd := &cobra.Command{}

	err = runTests(cmd, []string{tmpDir})
	assert.Error(t, err)
}

func TestRunTests_WithUpdateGolden(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "test.textproto")

	content := `
name: "test-manifest"
test_cases: {
  name: "test-1"
  request: { method: "GET", path: "/" }
  expectations: { phase: REQUEST_HEADERS, headers_response: {} }
}
`
	err := os.WriteFile(manifestPath, []byte(content), 0o644)
	require.NoError(t, err)

	// Override global flags
	oldTarget := target
	oldUpdateGolden := updateGolden

	target = "localhost:50051"
	updateGolden = true

	defer func() {
		target = oldTarget
		updateGolden = oldUpdateGolden
	}()

	cmd := &cobra.Command{}

	err = runTests(cmd, []string{tmpDir})
	assert.Error(t, err)
}

func TestRunTests_WithUnixSocket(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "test.textproto")

	content := `
name: "test-manifest"
test_cases: {
  name: "test-1"
  request: { method: "GET", path: "/" }
  expectations: { phase: REQUEST_HEADERS, headers_response: {} }
}
`
	err := os.WriteFile(manifestPath, []byte(content), 0o644)
	require.NoError(t, err)

	// Override global flags
	oldUnixSocket := unixSocket

	unixSocket = "/tmp/test.sock"

	defer func() {
		unixSocket = oldUnixSocket
	}()

	cmd := &cobra.Command{}

	err = runTests(cmd, []string{tmpDir})
	assert.Error(t, err)
}

func TestRunTests_WithTLS(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "test.textproto")

	content := `
name: "test-manifest"
test_cases: {
  name: "test-1"
  request: { method: "GET", path: "/" }
  expectations: { phase: REQUEST_HEADERS, headers_response: {} }
}
`
	err := os.WriteFile(manifestPath, []byte(content), 0o644)
	require.NoError(t, err)

	// Override global flags
	oldTarget := target
	oldTLSEnable := tlsEnable
	oldTLSCert := tlsCert
	oldTLSKey := tlsKey
	oldTLSCA := tlsCA

	target = "localhost:50051"
	tlsEnable = true
	tlsCert = ""
	tlsKey = ""
	tlsCA = ""

	defer func() {
		target = oldTarget
		tlsEnable = oldTLSEnable
		tlsCert = oldTLSCert
		tlsKey = oldTLSKey
		tlsCA = oldTLSCA
	}()

	cmd := &cobra.Command{}

	err = runTests(cmd, []string{tmpDir})
	assert.Error(t, err)
}

func TestRunTests_WithParallel(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "test.textproto")

	content := `
name: "test-manifest"
test_cases: {
  name: "test-1"
  request: { method: "GET", path: "/" }
  expectations: { phase: REQUEST_HEADERS, headers_response: {} }
}
`
	err := os.WriteFile(manifestPath, []byte(content), 0o644)
	require.NoError(t, err)

	// Override global flags
	oldTarget := target
	oldParallel := parallel

	target = "localhost:50051"
	parallel = 4

	defer func() {
		target = oldTarget
		parallel = oldParallel
	}()

	cmd := &cobra.Command{}

	err = runTests(cmd, []string{tmpDir})
	assert.Error(t, err)
}

func TestRunTests_VerboseOutput(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "test.textproto")

	content := `
name: "test-manifest"
test_cases: {
  name: "test-1"
  request: { method: "GET", path: "/" }
  expectations: { phase: REQUEST_HEADERS, headers_response: {} }
}
`
	err := os.WriteFile(manifestPath, []byte(content), 0o644)
	require.NoError(t, err)

	// Override global flags
	oldTarget := target
	oldVerbose := verbose

	target = "localhost:50051"
	verbose = true

	defer func() {
		target = oldTarget
		verbose = oldVerbose
	}()

	cmd := &cobra.Command{}

	err = runTests(cmd, []string{tmpDir})
	assert.Error(t, err)
}

func TestExecute_NoArgs(t *testing.T) {
	// Set args to just the binary name (no subcommand)
	oldArgs := os.Args
	os.Args = []string{"extproctor", "--help"}
	defer func() { os.Args = oldArgs }()

	// Execute should not error on help
	err := Execute()
	assert.NoError(t, err)
}
