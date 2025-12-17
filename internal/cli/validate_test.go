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

func TestValidateCmd_Basic(t *testing.T) {
	assert.NotNil(t, validateCmd)
	assert.Equal(t, "validate [paths...]", validateCmd.Use)
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

func TestValidateCmd_LongDescription(t *testing.T) {
	assert.NotEmpty(t, validateCmd.Long)
	assert.Contains(t, validateCmd.Long, "prototext")
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
