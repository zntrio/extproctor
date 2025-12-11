// SPDX-FileCopyrightText: 2025 Thibault NORMAND
// SPDX-License-Identifier: MIT

package runner

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"time"

	extproctorv1 "zntr.io/extproctor/gen/extproctor/v1"
	"zntr.io/extproctor/internal/client"
	"zntr.io/extproctor/internal/comparator"
	"zntr.io/extproctor/internal/golden"
	"zntr.io/extproctor/internal/manifest"
	"zntr.io/extproctor/internal/reporter"
)

// Runner executes test cases against an ExtProc service.
type Runner struct {
	client       *client.Client
	comparator   *comparator.Comparator
	reporter     reporter.Reporter
	parallel     int
	verbose      bool
	filter       string
	tags         []string
	updateGolden bool
}

// Option configures the runner.
type Option func(*Runner)

// WithParallel sets the parallelism level.
func WithParallel(n int) Option {
	return func(r *Runner) {
		r.parallel = n
	}
}

// WithReporter sets the reporter.
func WithReporter(rep reporter.Reporter) Option {
	return func(r *Runner) {
		r.reporter = rep
	}
}

// WithVerbose enables verbose output.
func WithVerbose(v bool) Option {
	return func(r *Runner) {
		r.verbose = v
	}
}

// WithFilter sets the test name filter pattern.
func WithFilter(pattern string) Option {
	return func(r *Runner) {
		r.filter = pattern
	}
}

// WithTags sets the tag filter.
func WithTags(tags []string) Option {
	return func(r *Runner) {
		r.tags = tags
	}
}

// WithUpdateGolden enables golden file updates.
func WithUpdateGolden(update bool) Option {
	return func(r *Runner) {
		r.updateGolden = update
	}
}

// New creates a new test runner.
func New(client *client.Client, opts ...Option) *Runner {
	r := &Runner{
		client:     client,
		comparator: comparator.New(),
		parallel:   1,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// Results contains the overall test run results.
type Results struct {
	Total    int
	Passed   int
	Failed   int
	Skipped  int
	Duration time.Duration
	Tests    []*TestResult
}

// TestResult contains the result of a single test.
type TestResult struct {
	Name        string
	Passed      bool
	Skipped     bool
	Duration    time.Duration
	Error       error
	Differences []comparator.Difference
	Unmatched   []*extproctorv1.ExtProcExpectation
}

// Run executes all test cases from the loaded manifests.
func (r *Runner) Run(ctx context.Context, manifests []*manifest.LoadedManifest) (*Results, error) {
	// Collect all test cases
	var testCases []*testCaseWithManifest
	for _, m := range manifests {
		for _, tc := range m.TestCases {
			if r.shouldRun(tc) {
				testCases = append(testCases, &testCaseWithManifest{
					testCase:   tc,
					manifest:   m,
					sourcePath: m.SourcePath,
				})
			}
		}
	}

	results := &Results{
		Total: len(testCases),
		Tests: make([]*TestResult, 0, len(testCases)),
	}

	if r.reporter != nil {
		r.reporter.StartSuite(len(testCases))
	}

	startTime := time.Now()

	if r.parallel > 1 {
		r.runParallel(ctx, testCases, results)
	} else {
		r.runSequential(ctx, testCases, results)
	}

	results.Duration = time.Since(startTime)

	if r.reporter != nil {
		r.reporter.EndSuite(reporter.SuiteSummary{
			Total:    results.Total,
			Passed:   results.Passed,
			Failed:   results.Failed,
			Skipped:  results.Skipped,
			Duration: results.Duration,
		})
	}

	return results, nil
}

type testCaseWithManifest struct {
	testCase   *extproctorv1.TestCase
	manifest   *manifest.LoadedManifest
	sourcePath string
}

// runSequential runs tests one at a time.
func (r *Runner) runSequential(ctx context.Context, testCases []*testCaseWithManifest, results *Results) {
	for _, tc := range testCases {
		select {
		case <-ctx.Done():
			return
		default:
		}

		result := r.runTest(ctx, tc)
		r.recordResult(results, result)
	}
}

// runParallel runs tests concurrently.
func (r *Runner) runParallel(ctx context.Context, testCases []*testCaseWithManifest, results *Results) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := make(chan struct{}, r.parallel)

	for _, tc := range testCases {
		select {
		case <-ctx.Done():
			return
		default:
		}

		wg.Add(1)
		sem <- struct{}{}

		go func(tc *testCaseWithManifest) {
			defer wg.Done()
			defer func() { <-sem }()

			result := r.runTest(ctx, tc)

			mu.Lock()
			r.recordResult(results, result)
			mu.Unlock()
		}(tc)
	}

	wg.Wait()
}

// runTest executes a single test case.
func (r *Runner) runTest(ctx context.Context, tc *testCaseWithManifest) *TestResult {
	if r.reporter != nil {
		r.reporter.StartTest(tc.testCase.Name)
	}

	startTime := time.Now()
	result := &TestResult{
		Name: tc.testCase.Name,
	}

	// Process the request
	procResult, err := r.client.Process(ctx, tc.testCase.Request)
	if err != nil {
		result.Error = err
		result.Duration = time.Since(startTime)
		r.reportResult(result)
		return result
	}

	// Get expectations (from inline or golden file)
	expectations, err := r.getExpectations(tc)
	if err != nil {
		result.Error = err
		result.Duration = time.Since(startTime)
		r.reportResult(result)
		return result
	}

	// Update golden file if requested
	if r.updateGolden && tc.testCase.GoldenFile != "" {
		goldenPath := r.resolveGoldenPath(tc)
		if err := golden.Write(goldenPath, procResult); err != nil {
			result.Error = err
			result.Duration = time.Since(startTime)
			r.reportResult(result)
			return result
		}
		result.Passed = true
		result.Duration = time.Since(startTime)
		r.reportResult(result)
		return result
	}

	// Compare expectations against actual responses
	compResult := r.comparator.Compare(expectations, procResult)

	result.Passed = compResult.Passed
	result.Differences = compResult.Differences
	result.Unmatched = compResult.Unmatched
	result.Duration = time.Since(startTime)

	r.reportResult(result)
	return result
}

// getExpectations returns expectations from inline definitions or golden files.
func (r *Runner) getExpectations(tc *testCaseWithManifest) ([]*extproctorv1.ExtProcExpectation, error) {
	if len(tc.testCase.Expectations) > 0 {
		return tc.testCase.Expectations, nil
	}

	if tc.testCase.GoldenFile != "" {
		goldenPath := r.resolveGoldenPath(tc)
		return golden.Read(goldenPath)
	}

	return nil, nil
}

// resolveGoldenPath resolves the golden file path relative to the manifest.
func (r *Runner) resolveGoldenPath(tc *testCaseWithManifest) string {
	if filepath.IsAbs(tc.testCase.GoldenFile) {
		return tc.testCase.GoldenFile
	}
	return filepath.Join(filepath.Dir(tc.sourcePath), tc.testCase.GoldenFile)
}

// reportResult reports a test result to the reporter.
func (r *Runner) reportResult(result *TestResult) {
	if r.reporter != nil {
		r.reporter.EndTest(reporter.TestResult{
			Name:        result.Name,
			Passed:      result.Passed,
			Skipped:     result.Skipped,
			Duration:    result.Duration,
			Error:       result.Error,
			Differences: result.Differences,
			Unmatched:   result.Unmatched,
		})
	}
}

// recordResult records a test result in the overall results.
func (r *Runner) recordResult(results *Results, result *TestResult) {
	results.Tests = append(results.Tests, result)

	if result.Skipped {
		results.Skipped++
	} else if result.Passed {
		results.Passed++
	} else {
		results.Failed++
	}
}

// shouldRun checks if a test case should be run based on filters.
func (r *Runner) shouldRun(tc *extproctorv1.TestCase) bool {
	// Check name filter
	if r.filter != "" {
		matched, err := filepath.Match(r.filter, tc.Name)
		if err != nil || !matched {
			return false
		}
	}

	// Check tag filter
	if len(r.tags) > 0 {
		hasMatchingTag := false
		for _, filterTag := range r.tags {
			for _, tcTag := range tc.Tags {
				if strings.EqualFold(filterTag, tcTag) {
					hasMatchingTag = true
					break
				}
			}
			if hasMatchingTag {
				break
			}
		}
		if !hasMatchingTag {
			return false
		}
	}

	return true
}
