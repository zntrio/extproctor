// SPDX-FileCopyrightText: 2025 Thibault NORMAND
// SPDX-License-Identifier: MIT

package runner

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	extproctorv1 "zntr.io/extproctor/gen/extproctor/v1"
	"zntr.io/extproctor/internal/comparator"
	"zntr.io/extproctor/internal/reporter"
)

func TestWithParallel(t *testing.T) {
	r := &Runner{}
	opt := WithParallel(4)
	opt(r)
	assert.Equal(t, 4, r.parallel)
}

func TestWithVerbose(t *testing.T) {
	r := &Runner{}
	opt := WithVerbose(true)
	opt(r)
	assert.True(t, r.verbose)

	opt = WithVerbose(false)
	opt(r)
	assert.False(t, r.verbose)
}

func TestWithFilter(t *testing.T) {
	r := &Runner{}
	opt := WithFilter("test-*")
	opt(r)
	assert.Equal(t, "test-*", r.filter)
}

func TestWithTags(t *testing.T) {
	r := &Runner{}
	opt := WithTags([]string{"smoke", "unit"})
	opt(r)
	assert.Equal(t, []string{"smoke", "unit"}, r.tags)
}

func TestWithUpdateGolden(t *testing.T) {
	r := &Runner{}
	opt := WithUpdateGolden(true)
	opt(r)
	assert.True(t, r.updateGolden)

	opt = WithUpdateGolden(false)
	opt(r)
	assert.False(t, r.updateGolden)
}

func TestWithReporter(t *testing.T) {
	r := &Runner{}
	mockReporter := &mockReporter{}
	opt := WithReporter(mockReporter)
	opt(r)
	assert.Equal(t, mockReporter, r.reporter)
}

func TestNew_DefaultValues(t *testing.T) {
	r := New(nil)
	assert.NotNil(t, r)
	assert.Equal(t, 1, r.parallel)
	assert.False(t, r.verbose)
	assert.Empty(t, r.filter)
	assert.Empty(t, r.tags)
	assert.False(t, r.updateGolden)
	assert.Nil(t, r.reporter)
	assert.NotNil(t, r.comparator)
}

func TestNew_WithOptions(t *testing.T) {
	mockReporter := &mockReporter{}
	r := New(nil,
		WithParallel(8),
		WithVerbose(true),
		WithFilter("test-*"),
		WithTags([]string{"smoke"}),
		WithUpdateGolden(true),
		WithReporter(mockReporter),
	)

	assert.Equal(t, 8, r.parallel)
	assert.True(t, r.verbose)
	assert.Equal(t, "test-*", r.filter)
	assert.Equal(t, []string{"smoke"}, r.tags)
	assert.True(t, r.updateGolden)
	assert.Equal(t, mockReporter, r.reporter)
}

func TestShouldRun_NoFilter(t *testing.T) {
	r := New(nil)

	tc := &extproctorv1.TestCase{
		Name: "test-case-1",
		Tags: []string{"smoke"},
	}

	assert.True(t, r.shouldRun(tc))
}

func TestShouldRun_MatchingFilter(t *testing.T) {
	r := New(nil, WithFilter("test-*"))

	tc := &extproctorv1.TestCase{
		Name: "test-case-1",
	}

	assert.True(t, r.shouldRun(tc))
}

func TestShouldRun_NonMatchingFilter(t *testing.T) {
	r := New(nil, WithFilter("other-*"))

	tc := &extproctorv1.TestCase{
		Name: "test-case-1",
	}

	assert.False(t, r.shouldRun(tc))
}

func TestShouldRun_ExactFilter(t *testing.T) {
	r := New(nil, WithFilter("test-case-1"))

	tc := &extproctorv1.TestCase{
		Name: "test-case-1",
	}

	assert.True(t, r.shouldRun(tc))

	tc.Name = "test-case-2"
	assert.False(t, r.shouldRun(tc))
}

func TestShouldRun_MatchingTag(t *testing.T) {
	r := New(nil, WithTags([]string{"smoke"}))

	tc := &extproctorv1.TestCase{
		Name: "test-case-1",
		Tags: []string{"smoke", "unit"},
	}

	assert.True(t, r.shouldRun(tc))
}

func TestShouldRun_NonMatchingTag(t *testing.T) {
	r := New(nil, WithTags([]string{"integration"}))

	tc := &extproctorv1.TestCase{
		Name: "test-case-1",
		Tags: []string{"smoke", "unit"},
	}

	assert.False(t, r.shouldRun(tc))
}

func TestShouldRun_MultipleTags(t *testing.T) {
	r := New(nil, WithTags([]string{"integration", "smoke"}))

	tc := &extproctorv1.TestCase{
		Name: "test-case-1",
		Tags: []string{"smoke"},
	}

	assert.True(t, r.shouldRun(tc))
}

func TestShouldRun_TagCaseInsensitive(t *testing.T) {
	r := New(nil, WithTags([]string{"SMOKE"}))

	tc := &extproctorv1.TestCase{
		Name: "test-case-1",
		Tags: []string{"smoke"},
	}

	assert.True(t, r.shouldRun(tc))
}

func TestShouldRun_FilterAndTags(t *testing.T) {
	r := New(nil, WithFilter("test-*"), WithTags([]string{"smoke"}))

	// Matching filter and tag
	tc := &extproctorv1.TestCase{
		Name: "test-case-1",
		Tags: []string{"smoke"},
	}
	assert.True(t, r.shouldRun(tc))

	// Matching filter but not tag
	tc = &extproctorv1.TestCase{
		Name: "test-case-2",
		Tags: []string{"unit"},
	}
	assert.False(t, r.shouldRun(tc))

	// Not matching filter
	tc = &extproctorv1.TestCase{
		Name: "other-case",
		Tags: []string{"smoke"},
	}
	assert.False(t, r.shouldRun(tc))
}

func TestShouldRun_InvalidFilterPattern(t *testing.T) {
	r := New(nil, WithFilter("[invalid"))

	tc := &extproctorv1.TestCase{
		Name: "test-case-1",
	}

	// Invalid patterns return false
	assert.False(t, r.shouldRun(tc))
}

func TestRecordResult_Passed(t *testing.T) {
	r := New(nil)
	results := &Results{
		Tests: make([]*TestResult, 0),
	}

	result := &TestResult{
		Name:   "test-1",
		Passed: true,
	}

	r.recordResult(results, result)

	assert.Len(t, results.Tests, 1)
	assert.Equal(t, 1, results.Passed)
	assert.Equal(t, 0, results.Failed)
	assert.Equal(t, 0, results.Skipped)
}

func TestRecordResult_Failed(t *testing.T) {
	r := New(nil)
	results := &Results{
		Tests: make([]*TestResult, 0),
	}

	result := &TestResult{
		Name:   "test-1",
		Passed: false,
	}

	r.recordResult(results, result)

	assert.Len(t, results.Tests, 1)
	assert.Equal(t, 0, results.Passed)
	assert.Equal(t, 1, results.Failed)
	assert.Equal(t, 0, results.Skipped)
}

func TestRecordResult_Skipped(t *testing.T) {
	r := New(nil)
	results := &Results{
		Tests: make([]*TestResult, 0),
	}

	result := &TestResult{
		Name:    "test-1",
		Skipped: true,
	}

	r.recordResult(results, result)

	assert.Len(t, results.Tests, 1)
	assert.Equal(t, 0, results.Passed)
	assert.Equal(t, 0, results.Failed)
	assert.Equal(t, 1, results.Skipped)
}

func TestRecordResult_Multiple(t *testing.T) {
	r := New(nil)
	results := &Results{
		Tests: make([]*TestResult, 0),
	}

	r.recordResult(results, &TestResult{Name: "test-1", Passed: true})
	r.recordResult(results, &TestResult{Name: "test-2", Passed: false})
	r.recordResult(results, &TestResult{Name: "test-3", Skipped: true})
	r.recordResult(results, &TestResult{Name: "test-4", Passed: true})

	assert.Len(t, results.Tests, 4)
	assert.Equal(t, 2, results.Passed)
	assert.Equal(t, 1, results.Failed)
	assert.Equal(t, 1, results.Skipped)
}

func TestResultsStruct(t *testing.T) {
	results := &Results{
		Total:    10,
		Passed:   7,
		Failed:   2,
		Skipped:  1,
		Duration: 5 * time.Second,
		Tests: []*TestResult{
			{
				Name:   "test-1",
				Passed: true,
			},
		},
	}

	assert.Equal(t, 10, results.Total)
	assert.Equal(t, 7, results.Passed)
	assert.Equal(t, 2, results.Failed)
	assert.Equal(t, 1, results.Skipped)
	assert.Equal(t, 5*time.Second, results.Duration)
	assert.Len(t, results.Tests, 1)
}

func TestTestResultStruct(t *testing.T) {
	result := &TestResult{
		Name:     "test-1",
		Passed:   false,
		Skipped:  false,
		Duration: 100 * time.Millisecond,
		Error:    assert.AnError,
		Differences: []comparator.Difference{
			{
				Phase:    extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Path:     "header",
				Expected: "exp",
				Actual:   "act",
			},
		},
		Unmatched: []*extproctorv1.ExtProcExpectation{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_BODY,
			},
		},
	}

	assert.Equal(t, "test-1", result.Name)
	assert.False(t, result.Passed)
	assert.False(t, result.Skipped)
	assert.Equal(t, 100*time.Millisecond, result.Duration)
	assert.NotNil(t, result.Error)
	assert.Len(t, result.Differences, 1)
	assert.Len(t, result.Unmatched, 1)
}

// mockReporter is a test double for the Reporter interface
type mockReporter struct {
	startSuiteCalled int
	startTestCalled  int
	endTestCalled    int
	endSuiteCalled   int
	lastTotal        int
	lastTestName     string
	lastResult       reporter.TestResult
	lastSummary      reporter.SuiteSummary
}

func (m *mockReporter) StartSuite(total int) {
	m.startSuiteCalled++
	m.lastTotal = total
}

func (m *mockReporter) StartTest(name string) {
	m.startTestCalled++
	m.lastTestName = name
}

func (m *mockReporter) EndTest(result reporter.TestResult) {
	m.endTestCalled++
	m.lastResult = result
}

func (m *mockReporter) EndSuite(summary reporter.SuiteSummary) {
	m.endSuiteCalled++
	m.lastSummary = summary
}

func TestReportResult_CallsReporter(t *testing.T) {
	mock := &mockReporter{}
	r := New(nil, WithReporter(mock))

	result := &TestResult{
		Name:     "test-1",
		Passed:   true,
		Duration: 100 * time.Millisecond,
	}

	r.reportResult(result)

	assert.Equal(t, 1, mock.endTestCalled)
	assert.Equal(t, "test-1", mock.lastResult.Name)
	assert.True(t, mock.lastResult.Passed)
}

func TestReportResult_NoReporter(t *testing.T) {
	r := New(nil)

	result := &TestResult{
		Name:   "test-1",
		Passed: true,
	}

	// Should not panic
	r.reportResult(result)
}

func TestReportResult_WithDifferences(t *testing.T) {
	mock := &mockReporter{}
	r := New(nil, WithReporter(mock))

	result := &TestResult{
		Name:   "test-1",
		Passed: false,
		Differences: []comparator.Difference{
			{
				Phase:    extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Path:     "header",
				Expected: "exp",
				Actual:   "act",
			},
		},
		Unmatched: []*extproctorv1.ExtProcExpectation{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_BODY,
			},
		},
	}

	r.reportResult(result)

	assert.Len(t, mock.lastResult.Differences, 1)
	assert.Len(t, mock.lastResult.Unmatched, 1)
}

func TestShouldRun_EmptyTags(t *testing.T) {
	r := New(nil, WithTags([]string{}))

	tc := &extproctorv1.TestCase{
		Name: "test-case-1",
		Tags: []string{"smoke"},
	}

	// Empty tags filter should not filter anything
	assert.True(t, r.shouldRun(tc))
}

func TestShouldRun_TestCaseWithNoTags(t *testing.T) {
	r := New(nil, WithTags([]string{"smoke"}))

	tc := &extproctorv1.TestCase{
		Name: "test-case-1",
		Tags: []string{},
	}

	// Test case has no tags, should not match
	assert.False(t, r.shouldRun(tc))
}

func TestShouldRun_MultipleTagsInTestCase(t *testing.T) {
	r := New(nil, WithTags([]string{"e2e"}))

	tc := &extproctorv1.TestCase{
		Name: "test-case-1",
		Tags: []string{"smoke", "unit", "e2e"},
	}

	assert.True(t, r.shouldRun(tc))
}

func TestNew_ComparatorInitialized(t *testing.T) {
	r := New(nil)
	assert.NotNil(t, r.comparator)
}

func TestWithFilter_EmptyString(t *testing.T) {
	r := New(nil, WithFilter(""))

	tc := &extproctorv1.TestCase{
		Name: "any-test-name",
	}

	// Empty filter should allow all tests
	assert.True(t, r.shouldRun(tc))
}

func TestResolveGoldenPath_Absolute(t *testing.T) {
	r := New(nil)

	tc := &testCaseWithManifest{
		testCase: &extproctorv1.TestCase{
			GoldenFile: "/absolute/path/golden.textproto",
		},
		sourcePath: "/some/path/manifest.textproto",
	}

	path := r.resolveGoldenPath(tc)
	assert.Equal(t, "/absolute/path/golden.textproto", path)
}

func TestResolveGoldenPath_Relative(t *testing.T) {
	r := New(nil)

	tc := &testCaseWithManifest{
		testCase: &extproctorv1.TestCase{
			GoldenFile: "golden/test.textproto",
		},
		sourcePath: "/some/path/manifest.textproto",
	}

	path := r.resolveGoldenPath(tc)
	assert.Equal(t, "/some/path/golden/test.textproto", path)
}

func TestGetExpectations_Inline(t *testing.T) {
	r := New(nil)

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
			Response: &extproctorv1.ExtProcExpectation_HeadersResponse{
				HeadersResponse: &extproctorv1.HeadersExpectation{},
			},
		},
	}

	tc := &testCaseWithManifest{
		testCase: &extproctorv1.TestCase{
			Expectations: expectations,
		},
	}

	result, err := r.getExpectations(tc)
	assert.NoError(t, err)
	assert.Equal(t, expectations, result)
}

func TestGetExpectations_NoExpectationsOrGolden(t *testing.T) {
	r := New(nil)

	tc := &testCaseWithManifest{
		testCase: &extproctorv1.TestCase{},
	}

	result, err := r.getExpectations(tc)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestGetExpectations_GoldenFile(t *testing.T) {
	tmpDir := t.TempDir()
	goldenPath := filepath.Join(tmpDir, "golden.textproto")

	// Create a valid golden file
	content := `
name: "golden"
expectations: {
  phase: REQUEST_HEADERS
  headers_response: {
    set_headers: {
      key: "x-test"
      value: "value"
    }
  }
}
`
	err := os.WriteFile(goldenPath, []byte(content), 0o644)
	require.NoError(t, err)

	r := New(nil)

	tc := &testCaseWithManifest{
		testCase: &extproctorv1.TestCase{
			GoldenFile: "golden.textproto",
		},
		sourcePath: filepath.Join(tmpDir, "manifest.textproto"),
	}

	expectations, err := r.getExpectations(tc)
	require.NoError(t, err)
	assert.NotNil(t, expectations)
	assert.Len(t, expectations, 1)
}

func TestGetExpectations_GoldenFileNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	r := New(nil)

	tc := &testCaseWithManifest{
		testCase: &extproctorv1.TestCase{
			GoldenFile: "nonexistent.textproto",
		},
		sourcePath: filepath.Join(tmpDir, "manifest.textproto"),
	}

	_, err := r.getExpectations(tc)
	assert.Error(t, err)
}
