// SPDX-FileCopyrightText: 2025 Thibault NORMAND
// SPDX-License-Identifier: MIT

package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCmd_Basic(t *testing.T) {
	assert.NotNil(t, runCmd)
	assert.Equal(t, "run [paths...]", runCmd.Use)
}

func TestRunCmd_HasUpdateGoldenFlag(t *testing.T) {
	f := runCmd.Flags().Lookup("update-golden")
	assert.NotNil(t, f)
	assert.Equal(t, "false", f.DefValue)
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

func TestRunCmd_LongDescription(t *testing.T) {
	assert.NotEmpty(t, runCmd.Long)
	assert.Contains(t, runCmd.Long, "ExtProc")
}

func TestRunTests_NoManifests(t *testing.T) {
	tmpDir := t.TempDir()

	// Override global flags for this test
	oldTarget := target
	target = "localhost:59999"
	defer func() { target = oldTarget }()

	cmd := &cobra.Command{}

	err := runTests(cmd, []string{tmpDir})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no test manifests found")
}

func TestRunTests_InvalidPath(t *testing.T) {
	// Override global flags for this test
	oldTarget := target
	target = "localhost:59999"
	defer func() { target = oldTarget }()

	cmd := &cobra.Command{}

	err := runTests(cmd, []string{"/nonexistent/path"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load manifests")
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

	target = "localhost:59999"
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

	target = "localhost:59999"
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

	target = "localhost:59999"
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

	target = "localhost:59999"
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

	target = "localhost:59999"
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

	target = "localhost:59999"
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

	target = "localhost:59999"
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

	target = "localhost:59999"
	verbose = true

	defer func() {
		target = oldTarget
		verbose = oldVerbose
	}()

	cmd := &cobra.Command{}

	err = runTests(cmd, []string{tmpDir})
	assert.Error(t, err)
}
