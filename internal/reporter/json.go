// SPDX-FileCopyrightText: 2025 Thibault NORMAND
// SPDX-License-Identifier: MIT

package reporter

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"zntr.io/extproctor/internal/comparator"
)

// JSONReporter outputs test results in JSON format for CI integration.
type JSONReporter struct {
	out     io.Writer
	results *jsonResults
}

type jsonResults struct {
	StartTime time.Time    `json:"start_time"`
	Tests     []jsonTest   `json:"tests"`
	Summary   *jsonSummary `json:"summary,omitempty"`
}

type jsonTest struct {
	Name        string           `json:"name"`
	Status      string           `json:"status"`
	Duration    string           `json:"duration"`
	Error       string           `json:"error,omitempty"`
	Differences []jsonDifference `json:"differences,omitempty"`
	Unmatched   []jsonUnmatched  `json:"unmatched,omitempty"`
	Unexpected  []jsonUnexpected `json:"unexpected,omitempty"`
}

type jsonUnmatched struct {
	Phase        string `json:"phase"`
	ResponseType string `json:"response_type"`
}

type jsonUnexpected struct {
	Phase        string `json:"phase"`
	ResponseType string `json:"response_type"`
}

type jsonDifference struct {
	Phase    string `json:"phase"`
	Path     string `json:"path"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
}

type jsonSummary struct {
	Total    int    `json:"total"`
	Passed   int    `json:"passed"`
	Failed   int    `json:"failed"`
	Skipped  int    `json:"skipped"`
	Duration string `json:"duration"`
}

// NewJSONReporter creates a new JSON reporter.
func NewJSONReporter(out io.Writer) *JSONReporter {
	return &JSONReporter{
		out: out,
		results: &jsonResults{
			StartTime: time.Now(),
			Tests:     make([]jsonTest, 0),
		},
	}
}

// StartSuite implements Reporter.
func (r *JSONReporter) StartSuite(total int) {
	r.results.StartTime = time.Now()
}

// StartTest implements Reporter.
func (r *JSONReporter) StartTest(name string) {
	// No-op for JSON reporter
}

// EndTest implements Reporter.
func (r *JSONReporter) EndTest(result TestResult) {
	status := "passed"
	if result.Skipped {
		status = "skipped"
	} else if !result.Passed {
		status = "failed"
	}

	test := jsonTest{
		Name:     result.Name,
		Status:   status,
		Duration: result.Duration.String(),
	}

	if result.Error != nil {
		test.Error = result.Error.Error()
	}

	for _, d := range result.Differences {
		test.Differences = append(test.Differences, jsonDifference{
			Phase:    d.Phase.String(),
			Path:     d.Path,
			Expected: d.Expected,
			Actual:   d.Actual,
		})
	}

	for _, u := range result.Unmatched {
		test.Unmatched = append(test.Unmatched, jsonUnmatched{
			Phase:        u.Phase.String(),
			ResponseType: formatResponseType(u.Response),
		})
	}

	for _, u := range result.Unexpected {
		test.Unexpected = append(test.Unexpected, jsonUnexpected{
			Phase:        u.Phase.String(),
			ResponseType: formatResponseType(u.Response.Response),
		})
	}

	r.results.Tests = append(r.results.Tests, test)
}

// EndSuite implements Reporter.
func (r *JSONReporter) EndSuite(summary SuiteSummary) {
	r.results.Summary = &jsonSummary{
		Total:    summary.Total,
		Passed:   summary.Passed,
		Failed:   summary.Failed,
		Skipped:  summary.Skipped,
		Duration: summary.Duration.String(),
	}

	encoder := json.NewEncoder(r.out)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(r.results)
}

// FormatDifference formats a difference for JSON output.
func FormatDifference(d comparator.Difference) jsonDifference {
	return jsonDifference{
		Phase:    d.Phase.String(),
		Path:     d.Path,
		Expected: d.Expected,
		Actual:   d.Actual,
	}
}

// formatResponseType returns a human-readable type name for a response.
func formatResponseType(resp any) string {
	return fmt.Sprintf("%T", resp)
}
