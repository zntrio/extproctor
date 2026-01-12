// SPDX-FileCopyrightText: 2025 Thibault NORMAND
// SPDX-License-Identifier: MIT

package reporter

import (
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
	"zntr.io/extproctor/internal/comparator"
)

// HumanReporter outputs test results in a human-readable format.
type HumanReporter struct {
	out     io.Writer
	verbose bool

	passColor *color.Color
	failColor *color.Color
	skipColor *color.Color
	dimColor  *color.Color
}

// NewHumanReporter creates a new human-readable reporter.
func NewHumanReporter(out io.Writer, verbose bool) *HumanReporter {
	return &HumanReporter{
		out:       out,
		verbose:   verbose,
		passColor: color.New(color.FgGreen),
		failColor: color.New(color.FgRed),
		skipColor: color.New(color.FgYellow),
		dimColor:  color.New(color.Faint),
	}
}

// StartSuite implements Reporter.
func (r *HumanReporter) StartSuite(total int) {
	_, _ = fmt.Fprintf(r.out, "Running %d test(s)...\n\n", total)
}

// StartTest implements Reporter.
func (r *HumanReporter) StartTest(name string) {
	if r.verbose {
		_, _ = fmt.Fprintf(r.out, "  %s ", name)
	}
}

// EndTest implements Reporter.
func (r *HumanReporter) EndTest(result TestResult) {
	var status string
	var statusColor *color.Color

	switch {
	case result.Skipped:
		status = "SKIP"
		statusColor = r.skipColor
	case result.Passed:
		status = "PASS"
		statusColor = r.passColor
	default:
		status = "FAIL"
		statusColor = r.failColor
	}

	if r.verbose {
		_, _ = statusColor.Fprintf(r.out, "[%s]", status)
		_, _ = r.dimColor.Fprintf(r.out, " (%s)\n", result.Duration)
	} else {
		// Compact output
		_, _ = statusColor.Fprintf(r.out, "  [%s] %s", status, result.Name)
		_, _ = r.dimColor.Fprintf(r.out, " (%s)\n", result.Duration)
	}

	// Show error if present
	if result.Error != nil {
		_, _ = r.failColor.Fprintf(r.out, "    Error: %v\n", result.Error)
	}

	// Show differences for failed tests
	if !result.Passed && !result.Skipped {
		if len(result.Differences) > 0 {
			_, _ = fmt.Fprintln(r.out, "    Differences:")
			for _, d := range result.Differences {
				_, _ = fmt.Fprintf(r.out, "      [%s] %s:\n", comparator.FormatDifferences([]comparator.Difference{d}), d.Path)
				_, _ = r.failColor.Fprintf(r.out, "        expected: %s\n", d.Expected)
				_, _ = r.passColor.Fprintf(r.out, "        actual:   %s\n", d.Actual)
			}
		}

		if len(result.Unmatched) > 0 {
			_, _ = fmt.Fprintln(r.out, "    Unmatched expectations:")
			for _, exp := range result.Unmatched {
				_, _ = fmt.Fprintf(r.out, "      - Phase: %s, Type: %T\n", exp.Phase, exp.Response)
			}
		}

		if len(result.Unexpected) > 0 {
			_, _ = fmt.Fprintln(r.out, "    Unexpected responses (not matched by any expectation):")
			for _, resp := range result.Unexpected {
				_, _ = fmt.Fprintf(r.out, "      - Phase: %s, Type: %T\n", resp.Phase, resp.Response.Response)
			}
		}
	}
}

// EndSuite implements Reporter.
func (r *HumanReporter) EndSuite(summary SuiteSummary) {
	_, _ = fmt.Fprintln(r.out, strings.Repeat("-", 60))

	// Summary line
	_, _ = fmt.Fprintf(r.out, "Results: ")
	_, _ = r.passColor.Fprintf(r.out, "%d passed", summary.Passed)
	_, _ = fmt.Fprintf(r.out, ", ")
	if summary.Failed > 0 {
		_, _ = r.failColor.Fprintf(r.out, "%d failed", summary.Failed)
	} else {
		_, _ = fmt.Fprintf(r.out, "%d failed", summary.Failed)
	}
	if summary.Skipped > 0 {
		_, _ = fmt.Fprintf(r.out, ", ")
		_, _ = r.skipColor.Fprintf(r.out, "%d skipped", summary.Skipped)
	}
	_, _ = fmt.Fprintf(r.out, " of %d total\n", summary.Total)

	// Duration
	_, _ = r.dimColor.Fprintf(r.out, "Duration: %s\n", summary.Duration)

	// Final status
	_, _ = fmt.Fprintln(r.out)
	if summary.Failed > 0 {
		_, _ = r.failColor.Fprintln(r.out, "FAILED")
	} else {
		_, _ = r.passColor.Fprintln(r.out, "PASSED")
	}
}
