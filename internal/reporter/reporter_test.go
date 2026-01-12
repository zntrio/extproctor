// SPDX-FileCopyrightText: 2025 Thibault NORMAND
// SPDX-License-Identifier: MIT

package reporter

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	extproctorv1 "zntr.io/extproctor/gen/extproctor/v1"
	"zntr.io/extproctor/internal/comparator"
)

func TestHumanReporter_StartSuite(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := NewHumanReporter(buf, false)

	reporter.StartSuite(10)

	assert.Contains(t, buf.String(), "Running 10 test(s)")
}

func TestHumanReporter_StartTest_Verbose(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := NewHumanReporter(buf, true)

	reporter.StartTest("test-case-1")

	assert.Contains(t, buf.String(), "test-case-1")
}

func TestHumanReporter_StartTest_NotVerbose(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := NewHumanReporter(buf, false)

	reporter.StartTest("test-case-1")

	// Should not output anything in non-verbose mode
	assert.Empty(t, buf.String())
}

func TestHumanReporter_EndTest_Passed(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := NewHumanReporter(buf, false)

	reporter.EndTest(TestResult{
		Name:     "test-case-1",
		Passed:   true,
		Duration: 100 * time.Millisecond,
	})

	output := buf.String()
	assert.Contains(t, output, "PASS")
	assert.Contains(t, output, "test-case-1")
}

func TestHumanReporter_EndTest_Passed_Verbose(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := NewHumanReporter(buf, true)

	reporter.StartTest("test-case-1")
	reporter.EndTest(TestResult{
		Name:     "test-case-1",
		Passed:   true,
		Duration: 100 * time.Millisecond,
	})

	output := buf.String()
	assert.Contains(t, output, "PASS")
	assert.Contains(t, output, "test-case-1")
}

func TestHumanReporter_EndTest_Failed(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := NewHumanReporter(buf, false)

	reporter.EndTest(TestResult{
		Name:     "test-case-1",
		Passed:   false,
		Duration: 100 * time.Millisecond,
		Differences: []comparator.Difference{
			{
				Phase:    extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Path:     "header_mutation.set_headers[x-custom]",
				Expected: "expected-value",
				Actual:   "actual-value",
			},
		},
	})

	output := buf.String()
	assert.Contains(t, output, "FAIL")
	assert.Contains(t, output, "test-case-1")
	assert.Contains(t, output, "Differences:")
	assert.Contains(t, output, "expected-value")
	assert.Contains(t, output, "actual-value")
}

func TestHumanReporter_EndTest_Skipped(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := NewHumanReporter(buf, false)

	reporter.EndTest(TestResult{
		Name:     "test-case-1",
		Skipped:  true,
		Duration: 100 * time.Millisecond,
	})

	output := buf.String()
	assert.Contains(t, output, "SKIP")
	assert.Contains(t, output, "test-case-1")
}

func TestHumanReporter_EndTest_WithError(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := NewHumanReporter(buf, false)

	reporter.EndTest(TestResult{
		Name:     "test-case-1",
		Passed:   false,
		Duration: 100 * time.Millisecond,
		Error:    assert.AnError,
	})

	output := buf.String()
	assert.Contains(t, output, "FAIL")
	assert.Contains(t, output, "Error:")
}

func TestHumanReporter_EndTest_WithUnmatched(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := NewHumanReporter(buf, false)

	reporter.EndTest(TestResult{
		Name:     "test-case-1",
		Passed:   false,
		Duration: 100 * time.Millisecond,
		Unmatched: []*extproctorv1.ExtProcExpectation{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_BODY,
				Response: &extproctorv1.ExtProcExpectation_BodyResponse{
					BodyResponse: &extproctorv1.BodyExpectation{},
				},
			},
		},
	})

	output := buf.String()
	assert.Contains(t, output, "FAIL")
	assert.Contains(t, output, "Unmatched expectations:")
	assert.Contains(t, output, "REQUEST_BODY")
}

func TestHumanReporter_EndSuite_AllPassed(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := NewHumanReporter(buf, false)

	reporter.EndSuite(SuiteSummary{
		Total:    5,
		Passed:   5,
		Failed:   0,
		Skipped:  0,
		Duration: 1 * time.Second,
	})

	output := buf.String()
	assert.Contains(t, output, "5 passed")
	assert.Contains(t, output, "0 failed")
	assert.Contains(t, output, "PASSED")
}

func TestHumanReporter_EndSuite_SomeFailed(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := NewHumanReporter(buf, false)

	reporter.EndSuite(SuiteSummary{
		Total:    5,
		Passed:   3,
		Failed:   2,
		Skipped:  0,
		Duration: 1 * time.Second,
	})

	output := buf.String()
	assert.Contains(t, output, "3 passed")
	assert.Contains(t, output, "2 failed")
	assert.Contains(t, output, "FAILED")
}

func TestHumanReporter_EndSuite_WithSkipped(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := NewHumanReporter(buf, false)

	reporter.EndSuite(SuiteSummary{
		Total:    5,
		Passed:   3,
		Failed:   0,
		Skipped:  2,
		Duration: 1 * time.Second,
	})

	output := buf.String()
	assert.Contains(t, output, "3 passed")
	assert.Contains(t, output, "2 skipped")
	assert.Contains(t, output, "PASSED")
}

func TestJSONReporter_StartSuite(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := NewJSONReporter(buf)

	reporter.StartSuite(10)

	// JSON reporter doesn't output anything until EndSuite
	assert.Empty(t, buf.String())
}

func TestJSONReporter_StartTest(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := NewJSONReporter(buf)

	reporter.StartTest("test-case-1")

	// JSON reporter doesn't output anything until EndSuite
	assert.Empty(t, buf.String())
}

func TestJSONReporter_EndTest(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := NewJSONReporter(buf)

	reporter.EndTest(TestResult{
		Name:     "test-case-1",
		Passed:   true,
		Duration: 100 * time.Millisecond,
	})

	// JSON reporter doesn't output anything until EndSuite
	assert.Empty(t, buf.String())
}

func TestJSONReporter_EndSuite(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := NewJSONReporter(buf)

	reporter.StartSuite(2)
	reporter.StartTest("test-1")
	reporter.EndTest(TestResult{
		Name:     "test-1",
		Passed:   true,
		Duration: 100 * time.Millisecond,
	})
	reporter.StartTest("test-2")
	reporter.EndTest(TestResult{
		Name:     "test-2",
		Passed:   false,
		Duration: 200 * time.Millisecond,
		Differences: []comparator.Difference{
			{
				Phase:    extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Path:     "header",
				Expected: "expected",
				Actual:   "actual",
			},
		},
	})

	reporter.EndSuite(SuiteSummary{
		Total:    2,
		Passed:   1,
		Failed:   1,
		Skipped:  0,
		Duration: 300 * time.Millisecond,
	})

	// Parse the JSON output
	var result jsonResults
	err := json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	assert.Len(t, result.Tests, 2)
	assert.Equal(t, "test-1", result.Tests[0].Name)
	assert.Equal(t, "passed", result.Tests[0].Status)
	assert.Equal(t, "test-2", result.Tests[1].Name)
	assert.Equal(t, "failed", result.Tests[1].Status)
	assert.Len(t, result.Tests[1].Differences, 1)

	assert.NotNil(t, result.Summary)
	assert.Equal(t, 2, result.Summary.Total)
	assert.Equal(t, 1, result.Summary.Passed)
	assert.Equal(t, 1, result.Summary.Failed)
}

func TestJSONReporter_EndTest_Skipped(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := NewJSONReporter(buf)

	reporter.StartSuite(1)
	reporter.EndTest(TestResult{
		Name:     "test-1",
		Skipped:  true,
		Duration: 100 * time.Millisecond,
	})
	reporter.EndSuite(SuiteSummary{
		Total:   1,
		Skipped: 1,
	})

	var result jsonResults
	err := json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	assert.Len(t, result.Tests, 1)
	assert.Equal(t, "skipped", result.Tests[0].Status)
}

func TestJSONReporter_EndTest_WithError(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := NewJSONReporter(buf)

	reporter.StartSuite(1)
	reporter.EndTest(TestResult{
		Name:     "test-1",
		Passed:   false,
		Duration: 100 * time.Millisecond,
		Error:    assert.AnError,
	})
	reporter.EndSuite(SuiteSummary{
		Total:  1,
		Failed: 1,
	})

	var result jsonResults
	err := json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	assert.Len(t, result.Tests, 1)
	assert.NotEmpty(t, result.Tests[0].Error)
}

func TestFormatDifference(t *testing.T) {
	diff := comparator.Difference{
		Phase:    extproctorv1.ProcessingPhase_REQUEST_HEADERS,
		Path:     "header_mutation.set_headers[x-custom]",
		Expected: "expected-value",
		Actual:   "actual-value",
	}

	formatted := FormatDifference(diff)

	assert.Equal(t, "REQUEST_HEADERS", formatted.Phase)
	assert.Equal(t, "header_mutation.set_headers[x-custom]", formatted.Path)
	assert.Equal(t, "expected-value", formatted.Expected)
	assert.Equal(t, "actual-value", formatted.Actual)
}

func TestHumanReporter_EndTest_Failed_MultipleDifferences(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := NewHumanReporter(buf, false)

	reporter.EndTest(TestResult{
		Name:     "test-case-1",
		Passed:   false,
		Duration: 100 * time.Millisecond,
		Differences: []comparator.Difference{
			{
				Phase:    extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Path:     "header1",
				Expected: "expected1",
				Actual:   "actual1",
			},
			{
				Phase:    extproctorv1.ProcessingPhase_REQUEST_BODY,
				Path:     "body",
				Expected: "expected2",
				Actual:   "actual2",
			},
		},
	})

	output := buf.String()
	assert.Contains(t, output, "FAIL")
	assert.Contains(t, output, "Differences:")
	assert.Contains(t, output, "header1")
	assert.Contains(t, output, "body")
}

func TestNewHumanReporter(t *testing.T) {
	buf := &bytes.Buffer{}

	// Test non-verbose
	reporter := NewHumanReporter(buf, false)
	assert.NotNil(t, reporter)
	assert.Equal(t, buf, reporter.out)
	assert.False(t, reporter.verbose)

	// Test verbose
	reporter = NewHumanReporter(buf, true)
	assert.NotNil(t, reporter)
	assert.True(t, reporter.verbose)
}

func TestNewJSONReporter(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := NewJSONReporter(buf)

	assert.NotNil(t, reporter)
	assert.Equal(t, buf, reporter.out)
	assert.NotNil(t, reporter.results)
	assert.NotNil(t, reporter.results.Tests)
}

func TestJSONReporter_FullFlow(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := NewJSONReporter(buf)

	reporter.StartSuite(3)

	reporter.StartTest("test-1")
	reporter.EndTest(TestResult{
		Name:     "test-1",
		Passed:   true,
		Duration: 50 * time.Millisecond,
	})

	reporter.StartTest("test-2")
	reporter.EndTest(TestResult{
		Name:     "test-2",
		Passed:   false,
		Duration: 100 * time.Millisecond,
		Error:    assert.AnError,
	})

	reporter.StartTest("test-3")
	reporter.EndTest(TestResult{
		Name:     "test-3",
		Skipped:  true,
		Duration: 10 * time.Millisecond,
	})

	reporter.EndSuite(SuiteSummary{
		Total:    3,
		Passed:   1,
		Failed:   1,
		Skipped:  1,
		Duration: 160 * time.Millisecond,
	})

	var result jsonResults
	err := json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	assert.Len(t, result.Tests, 3)
	assert.Equal(t, "passed", result.Tests[0].Status)
	assert.Equal(t, "failed", result.Tests[1].Status)
	assert.NotEmpty(t, result.Tests[1].Error)
	assert.Equal(t, "skipped", result.Tests[2].Status)

	assert.NotNil(t, result.Summary)
	assert.Equal(t, 3, result.Summary.Total)
	assert.Equal(t, 1, result.Summary.Passed)
	assert.Equal(t, 1, result.Summary.Failed)
	assert.Equal(t, 1, result.Summary.Skipped)
}

func TestHumanReporter_FullFlow(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := NewHumanReporter(buf, true)

	reporter.StartSuite(3)
	output := buf.String()
	assert.Contains(t, output, "Running 3 test(s)")

	reporter.StartTest("test-1")
	reporter.EndTest(TestResult{
		Name:     "test-1",
		Passed:   true,
		Duration: 50 * time.Millisecond,
	})

	reporter.StartTest("test-2")
	reporter.EndTest(TestResult{
		Name:     "test-2",
		Passed:   false,
		Duration: 100 * time.Millisecond,
		Differences: []comparator.Difference{
			{
				Phase:    extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Path:     "header",
				Expected: "exp",
				Actual:   "act",
			},
		},
	})

	reporter.StartTest("test-3")
	reporter.EndTest(TestResult{
		Name:     "test-3",
		Skipped:  true,
		Duration: 10 * time.Millisecond,
	})

	reporter.EndSuite(SuiteSummary{
		Total:    3,
		Passed:   1,
		Failed:   1,
		Skipped:  1,
		Duration: 160 * time.Millisecond,
	})

	output = buf.String()
	assert.Contains(t, output, "PASS")
	assert.Contains(t, output, "FAIL")
	assert.Contains(t, output, "SKIP")
	assert.Contains(t, output, "1 passed")
	assert.Contains(t, output, "1 failed")
	assert.Contains(t, output, "1 skipped")
}

func TestJSONReporter_StartTest_NoOp(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := NewJSONReporter(buf)

	// StartTest should be a no-op for JSON reporter
	reporter.StartTest("test-1")

	// Verify no output was written
	assert.Empty(t, buf.String())
}
