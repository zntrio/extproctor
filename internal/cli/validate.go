// SPDX-FileCopyrightText: 2025 Thibault NORMAND
// SPDX-License-Identifier: MIT

package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"zntr.io/extproctor/internal/manifest"
)

var validateCmd = &cobra.Command{
	Use:   "validate [paths...]",
	Short: "Validate manifest files without running tests",
	Long: `Validate checks that prototext manifest files are syntactically correct
and contain all required fields without actually running the tests.

Examples:
  # Validate all manifests in a directory
  extproctor validate ./tests/

  # Validate specific files
  extproctor validate test1.textproto test2.textproto`,
	Args: cobra.MinimumNArgs(1),
	RunE: validateManifests,
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func validateManifests(cmd *cobra.Command, args []string) error {
	loader := manifest.NewLoader()

	var hasErrors bool
	var totalManifests, totalTestCases int

	for _, path := range args {
		manifests, err := loader.LoadPath(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s: %v\n", path, err)
			hasErrors = true
			continue
		}

		for _, m := range manifests {
			totalManifests++
			totalTestCases += len(m.TestCases)

			// Validate each test case
			for _, tc := range m.TestCases {
				if err := manifest.ValidateTestCase(tc); err != nil {
					fmt.Fprintf(os.Stderr, "ERROR: %s: test case %q: %v\n", m.SourcePath, tc.Name, err)
					hasErrors = true
				}
			}
		}
	}

	if hasErrors {
		return fmt.Errorf("validation failed")
	}

	fmt.Printf("Validated %d manifest(s) with %d test case(s)\n", totalManifests, totalTestCases)
	return nil
}
