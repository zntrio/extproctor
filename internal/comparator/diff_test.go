// SPDX-FileCopyrightText: 2025 Thibault NORMAND
// SPDX-License-Identifier: MIT

package comparator

import (
	"testing"

	"github.com/stretchr/testify/assert"

	extproctorv1 "zntr.io/extproctor/gen/extproctor/v1"
)

func TestFormatDifferences_Empty(t *testing.T) {
	result := FormatDifferences(nil)
	assert.Empty(t, result)

	result = FormatDifferences([]Difference{})
	assert.Empty(t, result)
}

func TestFormatDifferences_Single(t *testing.T) {
	diffs := []Difference{
		{
			Phase:    extproctorv1.ProcessingPhase_REQUEST_HEADERS,
			Path:     "header_mutation.set_headers[x-custom]",
			Expected: "expected-value",
			Actual:   "actual-value",
		},
	}

	result := FormatDifferences(diffs)
	assert.Contains(t, result, "Differences:")
	assert.Contains(t, result, "REQUEST_HEADERS")
	assert.Contains(t, result, "header_mutation.set_headers[x-custom]")
	assert.Contains(t, result, "expected-value")
	assert.Contains(t, result, "actual-value")
}

func TestFormatDifferences_Multiple(t *testing.T) {
	diffs := []Difference{
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
	}

	result := FormatDifferences(diffs)
	assert.Contains(t, result, "REQUEST_HEADERS")
	assert.Contains(t, result, "REQUEST_BODY")
	assert.Contains(t, result, "header1")
	assert.Contains(t, result, "body")
}

func TestFormatUnmatched_Empty(t *testing.T) {
	result := FormatUnmatched(nil)
	assert.Empty(t, result)

	result = FormatUnmatched([]*extproctorv1.ExtProcExpectation{})
	assert.Empty(t, result)
}

func TestFormatUnmatched_Single(t *testing.T) {
	unmatched := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
			Response: &extproctorv1.ExtProcExpectation_HeadersResponse{
				HeadersResponse: &extproctorv1.HeadersExpectation{},
			},
		},
	}

	result := FormatUnmatched(unmatched)
	assert.Contains(t, result, "Unmatched expectations:")
	assert.Contains(t, result, "REQUEST_HEADERS")
	assert.Contains(t, result, "HeadersResponse")
}

func TestFormatUnmatched_Multiple(t *testing.T) {
	unmatched := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
			Response: &extproctorv1.ExtProcExpectation_HeadersResponse{
				HeadersResponse: &extproctorv1.HeadersExpectation{},
			},
		},
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_BODY,
			Response: &extproctorv1.ExtProcExpectation_BodyResponse{
				BodyResponse: &extproctorv1.BodyExpectation{},
			},
		},
	}

	result := FormatUnmatched(unmatched)
	assert.Contains(t, result, "REQUEST_HEADERS")
	assert.Contains(t, result, "REQUEST_BODY")
	assert.Contains(t, result, "HeadersResponse")
	assert.Contains(t, result, "BodyResponse")
}

func TestPhaseName_AllPhases(t *testing.T) {
	tests := []struct {
		phase    extproctorv1.ProcessingPhase
		expected string
	}{
		{extproctorv1.ProcessingPhase_REQUEST_HEADERS, "REQUEST_HEADERS"},
		{extproctorv1.ProcessingPhase_REQUEST_BODY, "REQUEST_BODY"},
		{extproctorv1.ProcessingPhase_REQUEST_TRAILERS, "REQUEST_TRAILERS"},
		{extproctorv1.ProcessingPhase_RESPONSE_HEADERS, "RESPONSE_HEADERS"},
		{extproctorv1.ProcessingPhase_RESPONSE_BODY, "RESPONSE_BODY"},
		{extproctorv1.ProcessingPhase_RESPONSE_TRAILERS, "RESPONSE_TRAILERS"},
		{extproctorv1.ProcessingPhase_PROCESSING_PHASE_UNSPECIFIED, "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, phaseName(tt.phase))
		})
	}
}
