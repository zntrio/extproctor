// SPDX-FileCopyrightText: 2025 Thibault NORMAND
// SPDX-License-Identifier: MIT

package cli

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/protocolbuffers/txtpbfmt/parser"
	"github.com/spf13/cobra"
)

var (
	fmtWrite bool
	fmtDiff  bool
)

var fmtCmd = &cobra.Command{
	Use:   "fmt [paths...]",
	Short: "Format textproto manifest files",
	Long: `Format textproto manifest files using txtpbfmt.

By default, fmt prints the formatted output to stdout for a single file,
or reports which files would be changed for multiple files/directories.

Examples:
  # Format a single file to stdout
  extproctor fmt test.textproto

  # Format files in-place
  extproctor fmt --write ./tests/

  # Show diff of what would change
  extproctor fmt --diff ./tests/

  # Format specific files in-place
  extproctor fmt -w test1.textproto test2.textproto`,
	Args: cobra.MinimumNArgs(1),
	RunE: runFmt,
}

func init() {
	fmtCmd.Flags().BoolVarP(&fmtWrite, "write", "w", false, "Write formatted output back to files (in-place)")
	fmtCmd.Flags().BoolVarP(&fmtDiff, "diff", "d", false, "Show diff of what would change")
	rootCmd.AddCommand(fmtCmd)
}

func runFmt(cmd *cobra.Command, args []string) error {
	// Collect all textproto files from paths
	var files []string
	for _, path := range args {
		collected, err := collectTextprotoFiles(path)
		if err != nil {
			return fmt.Errorf("failed to collect files from %s: %w", path, err)
		}
		files = append(files, collected...)
	}

	if len(files) == 0 {
		return fmt.Errorf("no .textproto files found in specified paths")
	}

	var hasChanges bool
	var hasErrors bool

	for _, file := range files {
		changed, err := formatFile(file, fmtWrite, fmtDiff, len(files) == 1)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s: %v\n", file, err)
			hasErrors = true
			continue
		}
		if changed {
			hasChanges = true
		}
	}

	if hasErrors {
		return fmt.Errorf("formatting failed for one or more files")
	}

	// If checking mode (no --write) and there are changes, return error for CI usage
	if !fmtWrite && hasChanges && len(files) > 1 {
		return fmt.Errorf("some files need formatting (use --write to fix)")
	}

	return nil
}

// collectTextprotoFiles walks paths and collects all .textproto files
func collectTextprotoFiles(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		// Single file
		return []string{path}, nil
	}

	// Walk directory
	var files []string
	err = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(p) == ".textproto" {
			files = append(files, p)
		}
		return nil
	})

	return files, err
}

// formatFile formats a single file and returns whether it was changed
func formatFile(path string, write, showDiff, singleFile bool) (bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	// Format using txtpbfmt
	formatted, err := parser.Format(content)
	if err != nil {
		return false, fmt.Errorf("parse error: %w", err)
	}

	// Check if content changed
	if bytes.Equal(content, formatted) {
		if singleFile && !write && !showDiff {
			// Single file to stdout - print even if unchanged
			fmt.Print(string(formatted))
		}
		return false, nil
	}

	// Content changed
	if write {
		// Write back to file
		if err := os.WriteFile(path, formatted, 0644); err != nil {
			return true, fmt.Errorf("write error: %w", err)
		}
		fmt.Printf("formatted %s\n", path)
	} else if showDiff {
		// Show diff
		fmt.Printf("--- %s (original)\n+++ %s (formatted)\n", path, path)
		printSimpleDiff(string(content), string(formatted))
	} else if singleFile {
		// Single file to stdout
		fmt.Print(string(formatted))
	} else {
		// Multiple files - just report
		fmt.Printf("%s needs formatting\n", path)
	}

	return true, nil
}

// printSimpleDiff prints a simple line-by-line diff
func printSimpleDiff(original, formatted string) {
	origLines := splitLines(original)
	fmtLines := splitLines(formatted)

	// Simple diff: show all original lines with -, then all formatted with +
	// This is a basic implementation; could use a proper diff algorithm
	fmt.Println("@@ changes @@")
	for _, line := range origLines {
		fmt.Printf("-%s\n", line)
	}
	for _, line := range fmtLines {
		fmt.Printf("+%s\n", line)
	}
}

// splitLines splits content into lines without trailing newlines
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
