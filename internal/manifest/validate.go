// SPDX-FileCopyrightText: 2025 Thibault NORMAND
// SPDX-License-Identifier: MIT

package manifest

import (
	"errors"
	"fmt"

	extproctorv1 "zntr.io/extproctor/gen/extproctor/v1"
)

// ValidationError represents a validation error with context.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidateTestCase validates a single test case.
func ValidateTestCase(tc *extproctorv1.TestCase) error {
	var errs []error

	if tc.Name == "" {
		errs = append(errs, &ValidationError{
			Field:   "name",
			Message: "test case name is required",
		})
	}

	if tc.Request == nil {
		errs = append(errs, &ValidationError{
			Field:   "request",
			Message: "request is required",
		})
	} else {
		if err := validateHttpRequest(tc.Request); err != nil {
			errs = append(errs, err)
		}
	}

	if len(tc.Expectations) == 0 && tc.GoldenFile == "" {
		errs = append(errs, &ValidationError{
			Field:   "expectations",
			Message: "at least one expectation or golden_file is required",
		})
	}

	for i, exp := range tc.Expectations {
		if err := validateExpectation(i, exp); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// validateHttpRequest validates an HTTP request definition.
func validateHttpRequest(req *extproctorv1.HttpRequest) error {
	var errs []error

	if req.Method == "" {
		errs = append(errs, &ValidationError{
			Field:   "request.method",
			Message: "HTTP method is required",
		})
	}

	if req.Path == "" {
		errs = append(errs, &ValidationError{
			Field:   "request.path",
			Message: "path is required",
		})
	}

	return errors.Join(errs...)
}

// validateExpectation validates a single expectation.
func validateExpectation(index int, exp *extproctorv1.ExtProcExpectation) error {
	var errs []error

	if exp.Phase == extproctorv1.ProcessingPhase_PROCESSING_PHASE_UNSPECIFIED {
		errs = append(errs, &ValidationError{
			Field:   fmt.Sprintf("expectations[%d].phase", index),
			Message: "processing phase is required",
		})
	}

	if exp.Response == nil {
		errs = append(errs, &ValidationError{
			Field:   fmt.Sprintf("expectations[%d].response", index),
			Message: "response is required",
		})
	}

	return errors.Join(errs...)
}

// ValidateManifest validates an entire test manifest.
func ValidateManifest(m *extproctorv1.TestManifest) error {
	var errs []error

	if len(m.TestCases) == 0 {
		errs = append(errs, &ValidationError{
			Field:   "test_cases",
			Message: "at least one test case is required",
		})
	}

	for _, tc := range m.TestCases {
		if err := ValidateTestCase(tc); err != nil {
			errs = append(errs, fmt.Errorf("test case %q: %w", tc.Name, err))
		}
	}

	return errors.Join(errs...)
}
