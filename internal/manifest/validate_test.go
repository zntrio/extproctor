// SPDX-FileCopyrightText: 2025 Thibault NORMAND
// SPDX-License-Identifier: MIT

package manifest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	extproctorv1 "zntr.io/extproctor/gen/extproctor/v1"
)

func TestValidateTestCase_Valid(t *testing.T) {
	tc := &extproctorv1.TestCase{
		Name: "valid-test",
		Request: &extproctorv1.HttpRequest{
			Method: "GET",
			Path:   "/api/test",
		},
		Expectations: []*extproctorv1.ExtProcExpectation{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Response: &extproctorv1.ExtProcExpectation_HeadersResponse{
					HeadersResponse: &extproctorv1.HeadersExpectation{},
				},
			},
		},
	}

	err := ValidateTestCase(tc)
	assert.NoError(t, err)
}

func TestValidateTestCase_MissingName(t *testing.T) {
	tc := &extproctorv1.TestCase{
		Request: &extproctorv1.HttpRequest{
			Method: "GET",
			Path:   "/api/test",
		},
		Expectations: []*extproctorv1.ExtProcExpectation{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Response: &extproctorv1.ExtProcExpectation_HeadersResponse{
					HeadersResponse: &extproctorv1.HeadersExpectation{},
				},
			},
		},
	}

	err := ValidateTestCase(tc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

func TestValidateTestCase_MissingRequest(t *testing.T) {
	tc := &extproctorv1.TestCase{
		Name: "test-without-request",
		Expectations: []*extproctorv1.ExtProcExpectation{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Response: &extproctorv1.ExtProcExpectation_HeadersResponse{
					HeadersResponse: &extproctorv1.HeadersExpectation{},
				},
			},
		},
	}

	err := ValidateTestCase(tc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "request")
}

func TestValidateTestCase_MissingExpectations(t *testing.T) {
	tc := &extproctorv1.TestCase{
		Name: "test-without-expectations",
		Request: &extproctorv1.HttpRequest{
			Method: "GET",
			Path:   "/api/test",
		},
	}

	err := ValidateTestCase(tc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expectation")
}

func TestValidateTestCase_WithGoldenFile(t *testing.T) {
	tc := &extproctorv1.TestCase{
		Name: "test-with-golden",
		Request: &extproctorv1.HttpRequest{
			Method: "GET",
			Path:   "/api/test",
		},
		GoldenFile: "golden/test.textproto",
	}

	err := ValidateTestCase(tc)
	assert.NoError(t, err)
}

func TestValidateTestCase_MissingMethod(t *testing.T) {
	tc := &extproctorv1.TestCase{
		Name: "test-missing-method",
		Request: &extproctorv1.HttpRequest{
			Path: "/api/test",
		},
		Expectations: []*extproctorv1.ExtProcExpectation{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Response: &extproctorv1.ExtProcExpectation_HeadersResponse{
					HeadersResponse: &extproctorv1.HeadersExpectation{},
				},
			},
		},
	}

	err := ValidateTestCase(tc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "method")
}

func TestValidateTestCase_MissingPhase(t *testing.T) {
	tc := &extproctorv1.TestCase{
		Name: "test-missing-phase",
		Request: &extproctorv1.HttpRequest{
			Method: "GET",
			Path:   "/api/test",
		},
		Expectations: []*extproctorv1.ExtProcExpectation{
			{
				Response: &extproctorv1.ExtProcExpectation_HeadersResponse{
					HeadersResponse: &extproctorv1.HeadersExpectation{},
				},
			},
		},
	}

	err := ValidateTestCase(tc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "phase")
}

func TestValidateTestCase_MissingPath(t *testing.T) {
	tc := &extproctorv1.TestCase{
		Name: "test-missing-path",
		Request: &extproctorv1.HttpRequest{
			Method: "GET",
		},
		Expectations: []*extproctorv1.ExtProcExpectation{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Response: &extproctorv1.ExtProcExpectation_HeadersResponse{
					HeadersResponse: &extproctorv1.HeadersExpectation{},
				},
			},
		},
	}

	err := ValidateTestCase(tc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path")
}

func TestValidateTestCase_MissingResponse(t *testing.T) {
	tc := &extproctorv1.TestCase{
		Name: "test-missing-response",
		Request: &extproctorv1.HttpRequest{
			Method: "GET",
			Path:   "/api/test",
		},
		Expectations: []*extproctorv1.ExtProcExpectation{
			{
				Phase:    extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Response: nil,
			},
		},
	}

	err := ValidateTestCase(tc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "response")
}

func TestValidateTestCase_MultipleErrors(t *testing.T) {
	tc := &extproctorv1.TestCase{
		// Missing name
		Request: &extproctorv1.HttpRequest{
			// Missing method
			Path: "/api/test",
		},
		Expectations: []*extproctorv1.ExtProcExpectation{
			{
				// Missing phase
				Response: &extproctorv1.ExtProcExpectation_HeadersResponse{
					HeadersResponse: &extproctorv1.HeadersExpectation{},
				},
			},
		},
	}

	err := ValidateTestCase(tc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name")
	assert.Contains(t, err.Error(), "method")
	assert.Contains(t, err.Error(), "phase")
}

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{
		Field:   "test_field",
		Message: "test message",
	}

	assert.Equal(t, "test_field: test message", err.Error())
}

func TestValidateManifest_Valid(t *testing.T) {
	m := &extproctorv1.TestManifest{
		Name: "test-manifest",
		TestCases: []*extproctorv1.TestCase{
			{
				Name: "test-1",
				Request: &extproctorv1.HttpRequest{
					Method: "GET",
					Path:   "/api/test",
				},
				Expectations: []*extproctorv1.ExtProcExpectation{
					{
						Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
						Response: &extproctorv1.ExtProcExpectation_HeadersResponse{
							HeadersResponse: &extproctorv1.HeadersExpectation{},
						},
					},
				},
			},
		},
	}

	err := ValidateManifest(m)
	assert.NoError(t, err)
}

func TestValidateManifest_NoTestCases(t *testing.T) {
	m := &extproctorv1.TestManifest{
		Name: "empty-manifest",
	}

	err := ValidateManifest(m)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "test_cases")
}

func TestValidateManifest_InvalidTestCase(t *testing.T) {
	m := &extproctorv1.TestManifest{
		Name: "manifest",
		TestCases: []*extproctorv1.TestCase{
			{
				// Missing name
				Request: &extproctorv1.HttpRequest{
					Method: "GET",
					Path:   "/api/test",
				},
				Expectations: []*extproctorv1.ExtProcExpectation{
					{
						Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
						Response: &extproctorv1.ExtProcExpectation_HeadersResponse{
							HeadersResponse: &extproctorv1.HeadersExpectation{},
						},
					},
				},
			},
		},
	}

	err := ValidateManifest(m)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

func TestValidateManifest_MultipleTestCases(t *testing.T) {
	m := &extproctorv1.TestManifest{
		Name: "manifest",
		TestCases: []*extproctorv1.TestCase{
			{
				Name: "test-1",
				Request: &extproctorv1.HttpRequest{
					Method: "GET",
					Path:   "/api/test1",
				},
				Expectations: []*extproctorv1.ExtProcExpectation{
					{
						Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
						Response: &extproctorv1.ExtProcExpectation_HeadersResponse{
							HeadersResponse: &extproctorv1.HeadersExpectation{},
						},
					},
				},
			},
			{
				Name: "test-2",
				Request: &extproctorv1.HttpRequest{
					Method: "POST",
					Path:   "/api/test2",
				},
				Expectations: []*extproctorv1.ExtProcExpectation{
					{
						Phase: extproctorv1.ProcessingPhase_REQUEST_BODY,
						Response: &extproctorv1.ExtProcExpectation_BodyResponse{
							BodyResponse: &extproctorv1.BodyExpectation{},
						},
					},
				},
			},
		},
	}

	err := ValidateManifest(m)
	assert.NoError(t, err)
}

func TestValidateManifest_MultipleInvalidTestCases(t *testing.T) {
	m := &extproctorv1.TestManifest{
		Name: "manifest",
		TestCases: []*extproctorv1.TestCase{
			{
				// Missing name
				Request: &extproctorv1.HttpRequest{
					Method: "GET",
					Path:   "/api/test1",
				},
				Expectations: []*extproctorv1.ExtProcExpectation{
					{
						Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
						Response: &extproctorv1.ExtProcExpectation_HeadersResponse{
							HeadersResponse: &extproctorv1.HeadersExpectation{},
						},
					},
				},
			},
			{
				Name: "test-2",
				// Missing request
				Expectations: []*extproctorv1.ExtProcExpectation{
					{
						Phase: extproctorv1.ProcessingPhase_REQUEST_BODY,
						Response: &extproctorv1.ExtProcExpectation_BodyResponse{
							BodyResponse: &extproctorv1.BodyExpectation{},
						},
					},
				},
			},
		},
	}

	err := ValidateManifest(m)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name")
	assert.Contains(t, err.Error(), "request")
}

func TestValidateTestCase_MultipleExpectations(t *testing.T) {
	tc := &extproctorv1.TestCase{
		Name: "test-multiple-expectations",
		Request: &extproctorv1.HttpRequest{
			Method: "POST",
			Path:   "/api/test",
		},
		Expectations: []*extproctorv1.ExtProcExpectation{
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
		},
	}

	err := ValidateTestCase(tc)
	assert.NoError(t, err)
}

func TestValidateTestCase_MultipleInvalidExpectations(t *testing.T) {
	tc := &extproctorv1.TestCase{
		Name: "test-invalid-expectations",
		Request: &extproctorv1.HttpRequest{
			Method: "POST",
			Path:   "/api/test",
		},
		Expectations: []*extproctorv1.ExtProcExpectation{
			{
				// Missing phase
				Response: &extproctorv1.ExtProcExpectation_HeadersResponse{
					HeadersResponse: &extproctorv1.HeadersExpectation{},
				},
			},
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_BODY,
				// Missing response
				Response: nil,
			},
		},
	}

	err := ValidateTestCase(tc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "phase")
	assert.Contains(t, err.Error(), "response")
}
