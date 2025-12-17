// SPDX-FileCopyrightText: 2025 Thibault NORMAND
// SPDX-License-Identifier: MIT

package manifest

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoader_LoadFile(t *testing.T) {
	// Create a temporary manifest file
	content := `
name: "test-manifest"
description: "A test manifest"
test_cases: {
  name: "test-case-1"
  description: "First test case"
  tags: ["smoke", "unit"]
  request: {
    method: "GET"
    path: "/api/v1/test"
    scheme: "https"
    authority: "example.com"
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
`
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "test.textproto")
	err := os.WriteFile(manifestPath, []byte(content), 0o644)
	require.NoError(t, err)

	loader := NewLoader()
	manifest, err := loader.LoadFile(manifestPath)
	require.NoError(t, err)

	assert.Equal(t, "test-manifest", manifest.Name)
	assert.Equal(t, "A test manifest", manifest.Description)
	assert.Len(t, manifest.TestCases, 1)
	assert.Equal(t, "test-case-1", manifest.TestCases[0].Name)
	assert.Equal(t, "GET", manifest.TestCases[0].Request.Method)
	assert.Equal(t, "/api/v1/test", manifest.TestCases[0].Request.Path)
}

func TestLoader_LoadFile_InvalidPrototext(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "invalid.textproto")
	err := os.WriteFile(manifestPath, []byte("invalid { prototext"), 0o644)
	require.NoError(t, err)

	loader := NewLoader()
	_, err = loader.LoadFile(manifestPath)
	assert.Error(t, err)
}

func TestLoader_LoadFile_DefaultName(t *testing.T) {
	// Manifest without a name should use filename as default
	content := `
test_cases: {
  name: "test"
  request: { method: "GET", path: "/" }
  expectations: { phase: REQUEST_HEADERS, headers_response: {} }
}
`
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "my-test.textproto")
	err := os.WriteFile(manifestPath, []byte(content), 0o644)
	require.NoError(t, err)

	loader := NewLoader()
	manifest, err := loader.LoadFile(manifestPath)
	require.NoError(t, err)

	assert.Equal(t, "my-test.textproto", manifest.Name)
}

func TestLoader_LoadFile_NonExistent(t *testing.T) {
	loader := NewLoader()
	_, err := loader.LoadFile("/nonexistent/path/test.textproto")
	assert.Error(t, err)
}

func TestLoader_LoadDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple manifest files
	manifest1 := `
name: "manifest-1"
test_cases: {
  name: "test-1"
  request: { method: "GET", path: "/" }
  expectations: { phase: REQUEST_HEADERS, headers_response: {} }
}
`
	manifest2 := `
name: "manifest-2"
test_cases: {
  name: "test-2"
  request: { method: "POST", path: "/api" }
  expectations: { phase: REQUEST_HEADERS, headers_response: {} }
}
`

	err := os.WriteFile(filepath.Join(tmpDir, "test1.textproto"), []byte(manifest1), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "test2.prototext"), []byte(manifest2), 0o644)
	require.NoError(t, err)

	loader := NewLoader()
	manifests, err := loader.LoadPath(tmpDir)
	require.NoError(t, err)

	assert.Len(t, manifests, 2)
}

func TestLoader_LoadDirectory_WithSubdirectories(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	err := os.MkdirAll(subDir, 0o755)
	require.NoError(t, err)

	manifest1 := `
name: "manifest-1"
test_cases: {
  name: "test-1"
  request: { method: "GET", path: "/" }
  expectations: { phase: REQUEST_HEADERS, headers_response: {} }
}
`
	manifest2 := `
name: "manifest-2"
test_cases: {
  name: "test-2"
  request: { method: "POST", path: "/api" }
  expectations: { phase: REQUEST_HEADERS, headers_response: {} }
}
`

	err = os.WriteFile(filepath.Join(tmpDir, "test1.textproto"), []byte(manifest1), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(subDir, "test2.textproto"), []byte(manifest2), 0o644)
	require.NoError(t, err)

	loader := NewLoader()
	manifests, err := loader.LoadPath(tmpDir)
	require.NoError(t, err)

	assert.Len(t, manifests, 2)
}

func TestLoader_LoadDirectory_IgnoresNonManifestFiles(t *testing.T) {
	tmpDir := t.TempDir()

	manifest := `
name: "manifest-1"
test_cases: {
  name: "test-1"
  request: { method: "GET", path: "/" }
  expectations: { phase: REQUEST_HEADERS, headers_response: {} }
}
`

	err := os.WriteFile(filepath.Join(tmpDir, "test.textproto"), []byte(manifest), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# README"), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "config.json"), []byte("{}"), 0o644)
	require.NoError(t, err)

	loader := NewLoader()
	manifests, err := loader.LoadPath(tmpDir)
	require.NoError(t, err)

	assert.Len(t, manifests, 1)
}

func TestLoader_LoadDirectory_InvalidManifestInDir(t *testing.T) {
	tmpDir := t.TempDir()

	err := os.WriteFile(filepath.Join(tmpDir, "invalid.textproto"), []byte("invalid { data"), 0o644)
	require.NoError(t, err)

	loader := NewLoader()
	_, err = loader.LoadPath(tmpDir)
	assert.Error(t, err)
}

func TestLoader_isManifestFile(t *testing.T) {
	loader := NewLoader()

	tests := []struct {
		path     string
		expected bool
	}{
		{"test.textproto", true},
		{"test.prototext", true},
		{"test.txtpb", true},
		{"test.proto", false},
		{"test.json", false},
		{"test.yaml", false},
		{"test.TEXTPROTO", true},
		{"test.PROTOTEXT", true},
		{"test.TxTpB", true},
		{"/some/path/to/test.textproto", true},
		{"/some/path/to/test.json", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.Equal(t, tt.expected, loader.isManifestFile(tt.path))
		})
	}
}

func TestLoader_LoadPaths_MultiplePaths(t *testing.T) {
	tmpDir1 := t.TempDir()
	tmpDir2 := t.TempDir()

	manifest1 := `
name: "manifest-1"
test_cases: {
  name: "test-1"
  request: { method: "GET", path: "/" }
  expectations: { phase: REQUEST_HEADERS, headers_response: {} }
}
`
	manifest2 := `
name: "manifest-2"
test_cases: {
  name: "test-2"
  request: { method: "POST", path: "/api" }
  expectations: { phase: REQUEST_HEADERS, headers_response: {} }
}
`

	err := os.WriteFile(filepath.Join(tmpDir1, "test1.textproto"), []byte(manifest1), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir2, "test2.textproto"), []byte(manifest2), 0o644)
	require.NoError(t, err)

	loader := NewLoader()
	manifests, err := loader.LoadPaths([]string{tmpDir1, tmpDir2})
	require.NoError(t, err)

	assert.Len(t, manifests, 2)
}

func TestLoader_LoadPaths_SingleFile(t *testing.T) {
	tmpDir := t.TempDir()

	manifest := `
name: "manifest-1"
test_cases: {
  name: "test-1"
  request: { method: "GET", path: "/" }
  expectations: { phase: REQUEST_HEADERS, headers_response: {} }
}
`

	manifestPath := filepath.Join(tmpDir, "test.textproto")
	err := os.WriteFile(manifestPath, []byte(manifest), 0o644)
	require.NoError(t, err)

	loader := NewLoader()
	manifests, err := loader.LoadPaths([]string{manifestPath})
	require.NoError(t, err)

	assert.Len(t, manifests, 1)
}

func TestLoader_LoadPaths_ErrorOnInvalid(t *testing.T) {
	loader := NewLoader()
	_, err := loader.LoadPaths([]string{"/nonexistent/path"})
	assert.Error(t, err)
}

func TestLoader_LoadPath_NonExistent(t *testing.T) {
	loader := NewLoader()
	_, err := loader.LoadPath("/nonexistent/path")
	assert.Error(t, err)
}

func TestLoader_LoadPath_File(t *testing.T) {
	tmpDir := t.TempDir()

	manifest := `
name: "manifest"
test_cases: {
  name: "test"
  request: { method: "GET", path: "/" }
  expectations: { phase: REQUEST_HEADERS, headers_response: {} }
}
`

	manifestPath := filepath.Join(tmpDir, "test.textproto")
	err := os.WriteFile(manifestPath, []byte(manifest), 0o644)
	require.NoError(t, err)

	loader := NewLoader()
	manifests, err := loader.LoadPath(manifestPath)
	require.NoError(t, err)

	assert.Len(t, manifests, 1)
	assert.Equal(t, "manifest", manifests[0].Name)
	assert.Equal(t, manifestPath, manifests[0].SourcePath)
}

func TestLoader_LoadFile_AllExtensions(t *testing.T) {
	tmpDir := t.TempDir()

	manifest := `
name: "manifest"
test_cases: {
  name: "test"
  request: { method: "GET", path: "/" }
  expectations: { phase: REQUEST_HEADERS, headers_response: {} }
}
`

	extensions := []string{".textproto", ".prototext", ".txtpb"}
	for _, ext := range extensions {
		t.Run(ext, func(t *testing.T) {
			manifestPath := filepath.Join(tmpDir, "test"+ext)
			err := os.WriteFile(manifestPath, []byte(manifest), 0o644)
			require.NoError(t, err)
			t.Cleanup(func() { _ = os.Remove(manifestPath) })

			loader := NewLoader()
			m, err := loader.LoadFile(manifestPath)
			require.NoError(t, err)
			assert.Equal(t, "manifest", m.Name)
		})
	}
}

func TestLoader_LoadPath_SingleManifest(t *testing.T) {
	tmpDir := t.TempDir()

	manifest := `
name: "manifest"
test_cases: {
  name: "test"
  request: { method: "GET", path: "/" }
  expectations: { phase: REQUEST_HEADERS, headers_response: {} }
}
`

	manifestPath := filepath.Join(tmpDir, "test.textproto")
	err := os.WriteFile(manifestPath, []byte(manifest), 0o644)
	require.NoError(t, err)

	loader := NewLoader()
	manifests, err := loader.LoadPath(manifestPath)
	require.NoError(t, err)
	assert.Len(t, manifests, 1)
	assert.Equal(t, manifestPath, manifests[0].SourcePath)
}

func TestLoader_LoadFile_LargeFile(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "large.textproto")

	// Create a file with many test cases but still under the 1MB limit
	content := `name: "large-manifest"
`
	for i := 0; i < 100; i++ {
		content += fmt.Sprintf(`
test_cases: {
  name: "test-%d"
  request: { method: "GET", path: "/test-%d" }
  expectations: { phase: REQUEST_HEADERS, headers_response: {} }
}
`, i, i)
	}

	err := os.WriteFile(manifestPath, []byte(content), 0o644)
	require.NoError(t, err)

	loader := NewLoader()
	manifest, err := loader.LoadFile(manifestPath)
	require.NoError(t, err)
	assert.Len(t, manifest.TestCases, 100)
}

func TestLoader_LoadDirectory_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	loader := NewLoader()
	manifests, err := loader.LoadPath(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, manifests)
}
