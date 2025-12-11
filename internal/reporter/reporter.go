// SPDX-FileCopyrightText: 2025 Thibault NORMAND
// SPDX-License-Identifier: MIT

package reporter

import (
	"time"

	extproctorv1 "zntr.io/extproctor/gen/extproctor/v1"
	"zntr.io/extproctor/internal/comparator"
)

// Reporter defines the interface for test result reporting.
type Reporter interface {
	// StartSuite is called when the test suite starts.
	StartSuite(total int)

	// StartTest is called when a test starts.
	StartTest(name string)

	// EndTest is called when a test completes.
	EndTest(result TestResult)

	// EndSuite is called when the test suite completes.
	EndSuite(summary SuiteSummary)
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

// SuiteSummary contains the summary of the entire test suite.
type SuiteSummary struct {
	Total    int
	Passed   int
	Failed   int
	Skipped  int
	Duration time.Duration
}
