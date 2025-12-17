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

func TestFmtCmd_Basic(t *testing.T) {
	assert.NotNil(t, fmtCmd)
	assert.Equal(t, "fmt [paths...]", fmtCmd.Use)
}

func TestFmtCmd_HasFlags(t *testing.T) {
	flags := fmtCmd.Flags()

	f := flags.Lookup("write")
	assert.NotNil(t, f)
	assert.Equal(t, "false", f.DefValue)

	f = flags.Lookup("diff")
	assert.NotNil(t, f)
	assert.Equal(t, "false", f.DefValue)
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "single line no newline",
			input:    "hello",
			expected: []string{"hello"},
		},
		{
			name:     "single line with newline",
			input:    "hello\n",
			expected: []string{"hello"},
		},
		{
			name:     "multiple lines",
			input:    "line1\nline2\nline3",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "multiple lines with trailing newline",
			input:    "line1\nline2\nline3\n",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "empty lines",
			input:    "line1\n\nline3\n",
			expected: []string{"line1", "", "line3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitLines(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPrintSimpleDiff(t *testing.T) {
	original := "line1\nline2\nline3"
	formatted := "line1\nline2-modified\nline3"

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printSimpleDiff(original, formatted)

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stdout = oldStdout

	output := buf.String()
	assert.Contains(t, output, "@@ changes @@")
	assert.Contains(t, output, "-line1")
	assert.Contains(t, output, "-line2")
	assert.Contains(t, output, "+line1")
	assert.Contains(t, output, "+line2-modified")
}

func TestCollectTextprotoFiles_SingleFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.textproto")
	err := os.WriteFile(testFile, []byte("content"), 0o644)
	require.NoError(t, err)

	files, err := collectTextprotoFiles(testFile)
	require.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, testFile, files[0])
}

func TestCollectTextprotoFiles_Directory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple textproto files
	file1 := filepath.Join(tmpDir, "test1.textproto")
	file2 := filepath.Join(tmpDir, "test2.textproto")
	file3 := filepath.Join(tmpDir, "other.json")

	err := os.WriteFile(file1, []byte("content1"), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(file2, []byte("content2"), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(file3, []byte("{}"), 0o644)
	require.NoError(t, err)

	files, err := collectTextprotoFiles(tmpDir)
	require.NoError(t, err)
	assert.Len(t, files, 2)
}

func TestCollectTextprotoFiles_Subdirectories(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	err := os.MkdirAll(subDir, 0o755)
	require.NoError(t, err)

	file1 := filepath.Join(tmpDir, "test1.textproto")
	file2 := filepath.Join(subDir, "test2.textproto")

	err = os.WriteFile(file1, []byte("content1"), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(file2, []byte("content2"), 0o644)
	require.NoError(t, err)

	files, err := collectTextprotoFiles(tmpDir)
	require.NoError(t, err)
	assert.Len(t, files, 2)
}

func TestCollectTextprotoFiles_NonExistent(t *testing.T) {
	_, err := collectTextprotoFiles("/nonexistent/path")
	assert.Error(t, err)
}

func TestFormatFile_NoChanges(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.textproto")

	// Write already formatted content
	content := `name: "test"
test_cases {
  name: "test-1"
  request {
    method: "GET"
    path: "/"
  }
}
`
	err := os.WriteFile(testFile, []byte(content), 0o644)
	require.NoError(t, err)

	changed, err := formatFile(testFile, false, false, false)
	require.NoError(t, err)
	assert.False(t, changed)
}

func TestFormatFile_WithChanges_SingleFileToStdout(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.textproto")

	// Write unformatted content
	content := `name:"test" test_cases{name:"test-1" request{method:"GET" path:"/"}}`
	err := os.WriteFile(testFile, []byte(content), 0o644)
	require.NoError(t, err)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	changed, err := formatFile(testFile, false, false, true)

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stdout = oldStdout

	require.NoError(t, err)
	assert.True(t, changed)
	assert.NotEmpty(t, buf.String())
}

func TestFormatFile_WithChanges_Write(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.textproto")

	// Write unformatted content
	content := `name:"test" test_cases{name:"test-1" request{method:"GET" path:"/"}}`
	err := os.WriteFile(testFile, []byte(content), 0o644)
	require.NoError(t, err)

	// Capture stdout to check the "formatted" message
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	changed, err := formatFile(testFile, true, false, false)

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stdout = oldStdout

	require.NoError(t, err)
	assert.True(t, changed)
	assert.Contains(t, buf.String(), "formatted")

	// Verify file was written
	formatted, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.NotEqual(t, content, string(formatted))
}

func TestFormatFile_WithChanges_Diff(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.textproto")

	// Write unformatted content
	content := `name:"test"`
	err := os.WriteFile(testFile, []byte(content), 0o644)
	require.NoError(t, err)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	changed, err := formatFile(testFile, false, true, false)

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stdout = oldStdout

	require.NoError(t, err)
	assert.True(t, changed)
	output := buf.String()
	assert.Contains(t, output, "---")
	assert.Contains(t, output, "+++")
}

func TestFormatFile_WithChanges_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.textproto")

	// Write unformatted content
	content := `name:"test"`
	err := os.WriteFile(testFile, []byte(content), 0o644)
	require.NoError(t, err)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	changed, err := formatFile(testFile, false, false, false)

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stdout = oldStdout

	require.NoError(t, err)
	assert.True(t, changed)
	assert.Contains(t, buf.String(), "needs formatting")
}

func TestFormatFile_NonExistent(t *testing.T) {
	_, err := formatFile("/nonexistent/file.textproto", false, false, false)
	assert.Error(t, err)
}

func TestRunFmt_SingleFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.textproto")

	content := `name: "test"
test_cases {
  name: "test-1"
  request {
    method: "GET"
    path: "/"
  }
}
`
	err := os.WriteFile(testFile, []byte(content), 0o644)
	require.NoError(t, err)

	cmd := &cobra.Command{}
	err = runFmt(cmd, []string{testFile})
	assert.NoError(t, err)
}

func TestRunFmt_Directory(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "test1.textproto")
	file2 := filepath.Join(tmpDir, "test2.textproto")

	// Write already formatted content
	content := `name: "test"
`
	err := os.WriteFile(file1, []byte(content), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(file2, []byte(content), 0o644)
	require.NoError(t, err)

	cmd := &cobra.Command{}
	err = runFmt(cmd, []string{tmpDir})
	assert.NoError(t, err)
}

func TestRunFmt_NoFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a directory with no textproto files
	cmd := &cobra.Command{}
	err := runFmt(cmd, []string{tmpDir})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no .textproto files found")
}

func TestRunFmt_NonExistentPath(t *testing.T) {
	cmd := &cobra.Command{}
	err := runFmt(cmd, []string{"/nonexistent/path"})
	assert.Error(t, err)
}

func TestRunFmt_MultipleFilesNeedFormatting(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "test1.textproto")
	file2 := filepath.Join(tmpDir, "test2.textproto")

	// Write unformatted content
	unformatted := `name:"test" test_cases{name:"test-1"}`
	err := os.WriteFile(file1, []byte(unformatted), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(file2, []byte(unformatted), 0o644)
	require.NoError(t, err)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	err = runFmt(cmd, []string{tmpDir})

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stdout = oldStdout

	// Should return error when files need formatting and --write is not set
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "some files need formatting")
}

func TestRunFmt_WriteMode(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "test1.textproto")

	// Write unformatted content
	unformatted := `name:"test"`
	err := os.WriteFile(file1, []byte(unformatted), 0o644)
	require.NoError(t, err)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Enable write mode
	fmtWrite = true
	defer func() { fmtWrite = false }()

	cmd := &cobra.Command{}
	err = runFmt(cmd, []string{tmpDir})

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stdout = oldStdout

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "formatted")

	// Verify file was written
	formatted, err := os.ReadFile(file1)
	require.NoError(t, err)
	assert.NotEqual(t, unformatted, string(formatted))
}

func TestRunFmt_DiffMode(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "test1.textproto")

	// Write unformatted content
	unformatted := `name:"test"`
	err := os.WriteFile(file1, []byte(unformatted), 0o644)
	require.NoError(t, err)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Enable diff mode
	fmtDiff = true
	defer func() { fmtDiff = false }()

	cmd := &cobra.Command{}
	err = runFmt(cmd, []string{file1})

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stdout = oldStdout

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "---")
	assert.Contains(t, output, "+++")
}

func TestCollectTextprotoFiles_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a directory with no .textproto files
	files, err := collectTextprotoFiles(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, files)
}

func TestFormatFile_SingleFile_AlreadyFormatted(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.textproto")

	// Write already formatted content
	content := `name: "test"
`
	err := os.WriteFile(testFile, []byte(content), 0o644)
	require.NoError(t, err)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	changed, err := formatFile(testFile, false, false, true)

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stdout = oldStdout

	require.NoError(t, err)
	assert.False(t, changed)
	assert.NotEmpty(t, buf.String()) // Should print to stdout even if unchanged
}

func TestRunFmt_MultiplePaths(t *testing.T) {
	tmpDir1 := t.TempDir()
	tmpDir2 := t.TempDir()

	file1 := filepath.Join(tmpDir1, "test1.textproto")
	file2 := filepath.Join(tmpDir2, "test2.textproto")

	content := `name: "test"
`
	err := os.WriteFile(file1, []byte(content), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(file2, []byte(content), 0o644)
	require.NoError(t, err)

	cmd := &cobra.Command{}
	err = runFmt(cmd, []string{tmpDir1, tmpDir2})
	assert.NoError(t, err)
}

func TestFormatFile_WriteError(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.textproto")

	// Write unformatted content
	unformatted := `name:"test"`
	err := os.WriteFile(testFile, []byte(unformatted), 0o644)
	require.NoError(t, err)

	// Make file read-only to cause write error
	err = os.Chmod(testFile, 0o444)
	require.NoError(t, err)
	defer func() { _ = os.Chmod(testFile, 0o644) }()

	// Try to format with write mode
	_, err = formatFile(testFile, true, false, false)
	assert.Error(t, err)
}

func TestPrintSimpleDiff_EmptyStrings(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printSimpleDiff("", "")

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stdout = oldStdout

	output := buf.String()
	assert.Contains(t, output, "@@ changes @@")
}

func TestSplitLines_OnlyNewline(t *testing.T) {
	result := splitLines("\n")
	assert.Equal(t, []string{""}, result)
}

func TestRunFmt_SingleFileStdoutMode(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.textproto")

	content := `name:"test"`
	err := os.WriteFile(testFile, []byte(content), 0o644)
	require.NoError(t, err)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	err = runFmt(cmd, []string{testFile})

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stdout = oldStdout

	require.NoError(t, err)
	assert.NotEmpty(t, buf.String())
}

func TestCollectTextprotoFiles_WalkDirError(t *testing.T) {
	// Try to collect files from a file (not a directory)
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("content"), 0o644)
	require.NoError(t, err)

	// Collecting from a file should work (returns the file)
	files, err := collectTextprotoFiles(testFile)
	require.NoError(t, err)
	assert.Len(t, files, 1)
}

func TestFormatFile_ReadError(t *testing.T) {
	// Try to format a non-existent file
	_, err := formatFile("/nonexistent/file.textproto", false, false, false)
	assert.Error(t, err)
}

func TestRunFmt_MixedFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// One properly formatted, one not
	file1 := filepath.Join(tmpDir, "formatted.textproto")
	file2 := filepath.Join(tmpDir, "unformatted.textproto")

	formatted := `name: "test"
`
	unformatted := `name:"test"`

	err := os.WriteFile(file1, []byte(formatted), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(file2, []byte(unformatted), 0o644)
	require.NoError(t, err)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	err = runFmt(cmd, []string{tmpDir})

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stdout = oldStdout

	// Should error because one file needs formatting
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "some files need formatting")
}

func TestCollectTextprotoFiles_DirectoryWithSubdirError(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	err := os.MkdirAll(subDir, 0o755)
	require.NoError(t, err)

	// Create some textproto files
	file1 := filepath.Join(tmpDir, "test1.textproto")
	file2 := filepath.Join(subDir, "test2.textproto")

	err = os.WriteFile(file1, []byte("name: \"test1\""), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(file2, []byte("name: \"test2\""), 0o644)
	require.NoError(t, err)

	files, err := collectTextprotoFiles(tmpDir)
	require.NoError(t, err)
	assert.Len(t, files, 2)
}

func TestRunFmt_SingleFileNoChanges(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.textproto")

	// Write already formatted content
	content := `name: "test"
`
	err := os.WriteFile(testFile, []byte(content), 0o644)
	require.NoError(t, err)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	err = runFmt(cmd, []string{testFile})

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stdout = oldStdout

	require.NoError(t, err)
	// Single file should print to stdout even if no changes
	assert.NotEmpty(t, buf.String())
}

func TestRunFmt_MultipleFilesWithDiff(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "test1.textproto")
	file2 := filepath.Join(tmpDir, "test2.textproto")

	unformatted := `name:"test"`
	err := os.WriteFile(file1, []byte(unformatted), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(file2, []byte(unformatted), 0o644)
	require.NoError(t, err)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Enable diff mode
	fmtDiff = true
	defer func() { fmtDiff = false }()

	cmd := &cobra.Command{}
	err = runFmt(cmd, []string{tmpDir})

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stdout = oldStdout

	// With multiple files needing formatting and no --write, should error
	assert.Error(t, err)
	output := buf.String()
	assert.Contains(t, output, "---")
}
