// SPDX-FileCopyrightText: 2025 Thibault NORMAND
// SPDX-License-Identifier: MIT

package comparator

import (
	"testing"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extprocv3 "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/stretchr/testify/assert"

	extproctorv1 "zntr.io/extproctor/gen/extproctor/v1"
	"zntr.io/extproctor/internal/client"
)

func TestComparator_Compare_ExactMatch(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
			Response: &extproctorv1.ExtProcExpectation_HeadersResponse{
				HeadersResponse: &extproctorv1.HeadersExpectation{
					SetHeaders: map[string]string{
						"x-custom-header": "custom-value",
					},
				},
			},
		},
	}

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_RequestHeaders{
						RequestHeaders: &extprocv3.HeadersResponse{
							Response: &extprocv3.CommonResponse{
								HeaderMutation: &extprocv3.HeaderMutation{
									SetHeaders: []*corev3.HeaderValueOption{
										{
											Header: &corev3.HeaderValue{
												Key:   "x-custom-header",
												Value: "custom-value",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.True(t, compResult.Passed)
	assert.Empty(t, compResult.Unmatched)
}

func TestComparator_Compare_Mismatch(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
			Response: &extproctorv1.ExtProcExpectation_HeadersResponse{
				HeadersResponse: &extproctorv1.HeadersExpectation{
					SetHeaders: map[string]string{
						"x-custom-header": "expected-value",
					},
				},
			},
		},
	}

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_RequestHeaders{
						RequestHeaders: &extprocv3.HeadersResponse{
							Response: &extprocv3.CommonResponse{
								HeaderMutation: &extprocv3.HeaderMutation{
									SetHeaders: []*corev3.HeaderValueOption{
										{
											Header: &corev3.HeaderValue{
												Key:   "x-custom-header",
												Value: "actual-value",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.False(t, compResult.Passed)
	assert.NotEmpty(t, compResult.Differences)
}

func TestComparator_Compare_UnmatchedExpectation(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
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

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_RequestHeaders{
						RequestHeaders: &extprocv3.HeadersResponse{},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.False(t, compResult.Passed)
	assert.Len(t, compResult.Unmatched, 1)
	assert.Equal(t, extproctorv1.ProcessingPhase_REQUEST_BODY, compResult.Unmatched[0].Phase)
}

func TestComparator_Compare_ImmediateResponse(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
			Response: &extproctorv1.ExtProcExpectation_ImmediateResponse{
				ImmediateResponse: &extproctorv1.ImmediateExpectation{
					StatusCode: 403,
					Body:       []byte("Forbidden"),
				},
			},
		},
	}

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_ImmediateResponse{
						ImmediateResponse: &extprocv3.ImmediateResponse{
							Status: &typev3.HttpStatus{
								Code: typev3.StatusCode_Forbidden,
							},
							Body: []byte("Forbidden"),
						},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.True(t, compResult.Passed)
}

func TestComparator_Compare_BodyResponse_Match(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_BODY,
			Response: &extproctorv1.ExtProcExpectation_BodyResponse{
				BodyResponse: &extproctorv1.BodyExpectation{
					Body: []byte("modified body"),
				},
			},
		},
	}

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_BODY,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_RequestBody{
						RequestBody: &extprocv3.BodyResponse{
							Response: &extprocv3.CommonResponse{
								BodyMutation: &extprocv3.BodyMutation{
									Mutation: &extprocv3.BodyMutation_Body{
										Body: []byte("modified body"),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.True(t, compResult.Passed)
}

func TestComparator_Compare_BodyResponse_ClearBody(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_BODY,
			Response: &extproctorv1.ExtProcExpectation_BodyResponse{
				BodyResponse: &extproctorv1.BodyExpectation{
					ClearBody: true,
				},
			},
		},
	}

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_BODY,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_RequestBody{
						RequestBody: &extprocv3.BodyResponse{
							Response: &extprocv3.CommonResponse{
								BodyMutation: &extprocv3.BodyMutation{
									Mutation: &extprocv3.BodyMutation_ClearBody{
										ClearBody: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.True(t, compResult.Passed)
}

func TestComparator_Compare_BodyResponse_Mismatch(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_BODY,
			Response: &extproctorv1.ExtProcExpectation_BodyResponse{
				BodyResponse: &extproctorv1.BodyExpectation{
					Body: []byte("expected body"),
				},
			},
		},
	}

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_BODY,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_RequestBody{
						RequestBody: &extprocv3.BodyResponse{
							Response: &extprocv3.CommonResponse{
								BodyMutation: &extprocv3.BodyMutation{
									Mutation: &extprocv3.BodyMutation_Body{
										Body: []byte("actual body"),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.False(t, compResult.Passed)
	assert.NotEmpty(t, compResult.Differences)
}

func TestComparator_Compare_BodyResponse_NilMutation(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_BODY,
			Response: &extproctorv1.ExtProcExpectation_BodyResponse{
				BodyResponse: &extproctorv1.BodyExpectation{
					Body: []byte("expected body"),
				},
			},
		},
	}

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_BODY,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_RequestBody{
						RequestBody: &extprocv3.BodyResponse{
							Response: &extprocv3.CommonResponse{},
						},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.False(t, compResult.Passed)
	assert.NotEmpty(t, compResult.Differences)
}

func TestComparator_Compare_TrailersResponse_Match(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_TRAILERS,
			Response: &extproctorv1.ExtProcExpectation_TrailersResponse{
				TrailersResponse: &extproctorv1.TrailersExpectation{
					SetTrailers: map[string]string{
						"x-trailer": "value",
					},
				},
			},
		},
	}

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_TRAILERS,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_RequestTrailers{
						RequestTrailers: &extprocv3.TrailersResponse{
							HeaderMutation: &extprocv3.HeaderMutation{
								SetHeaders: []*corev3.HeaderValueOption{
									{
										Header: &corev3.HeaderValue{
											Key:   "x-trailer",
											Value: "value",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.True(t, compResult.Passed)
}

func TestComparator_Compare_TrailersResponse_Mismatch(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_TRAILERS,
			Response: &extproctorv1.ExtProcExpectation_TrailersResponse{
				TrailersResponse: &extproctorv1.TrailersExpectation{
					SetTrailers: map[string]string{
						"x-trailer": "expected",
					},
				},
			},
		},
	}

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_TRAILERS,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_RequestTrailers{
						RequestTrailers: &extprocv3.TrailersResponse{
							HeaderMutation: &extprocv3.HeaderMutation{
								SetHeaders: []*corev3.HeaderValueOption{
									{
										Header: &corev3.HeaderValue{
											Key:   "x-trailer",
											Value: "actual",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.False(t, compResult.Passed)
	assert.NotEmpty(t, compResult.Differences)
}

func TestComparator_Compare_TrailersResponse_Missing(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_TRAILERS,
			Response: &extproctorv1.ExtProcExpectation_TrailersResponse{
				TrailersResponse: &extproctorv1.TrailersExpectation{
					SetTrailers: map[string]string{
						"x-trailer": "value",
					},
				},
			},
		},
	}

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_TRAILERS,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_RequestTrailers{
						RequestTrailers: &extprocv3.TrailersResponse{
							HeaderMutation: &extprocv3.HeaderMutation{},
						},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.False(t, compResult.Passed)
	assert.NotEmpty(t, compResult.Differences)
}

func TestComparator_Compare_ImmediateResponse_StatusMismatch(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
			Response: &extproctorv1.ExtProcExpectation_ImmediateResponse{
				ImmediateResponse: &extproctorv1.ImmediateExpectation{
					StatusCode: 401,
				},
			},
		},
	}

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_ImmediateResponse{
						ImmediateResponse: &extprocv3.ImmediateResponse{
							Status: &typev3.HttpStatus{
								Code: typev3.StatusCode_Forbidden,
							},
						},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.False(t, compResult.Passed)
	assert.NotEmpty(t, compResult.Differences)
}

func TestComparator_Compare_ImmediateResponse_BodyMismatch(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
			Response: &extproctorv1.ExtProcExpectation_ImmediateResponse{
				ImmediateResponse: &extproctorv1.ImmediateExpectation{
					Body: []byte("expected body"),
				},
			},
		},
	}

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_ImmediateResponse{
						ImmediateResponse: &extprocv3.ImmediateResponse{
							Body: []byte("actual body"),
						},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.False(t, compResult.Passed)
	assert.NotEmpty(t, compResult.Differences)
}

func TestComparator_Compare_ImmediateResponse_HeadersMismatch(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
			Response: &extproctorv1.ExtProcExpectation_ImmediateResponse{
				ImmediateResponse: &extproctorv1.ImmediateExpectation{
					Headers: map[string]string{
						"x-error": "expected",
					},
				},
			},
		},
	}

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_ImmediateResponse{
						ImmediateResponse: &extprocv3.ImmediateResponse{
							Headers: &extprocv3.HeaderMutation{
								SetHeaders: []*corev3.HeaderValueOption{
									{
										Header: &corev3.HeaderValue{
											Key:   "x-error",
											Value: "actual",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.False(t, compResult.Passed)
	assert.NotEmpty(t, compResult.Differences)
}

func TestComparator_Compare_ImmediateResponse_HeadersNotSet(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
			Response: &extproctorv1.ExtProcExpectation_ImmediateResponse{
				ImmediateResponse: &extproctorv1.ImmediateExpectation{
					Headers: map[string]string{
						"x-error": "expected",
					},
				},
			},
		},
	}

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_ImmediateResponse{
						ImmediateResponse: &extprocv3.ImmediateResponse{
							Headers: &extprocv3.HeaderMutation{
								SetHeaders: []*corev3.HeaderValueOption{},
							},
						},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.False(t, compResult.Passed)
	assert.NotEmpty(t, compResult.Differences)
}

func TestComparator_Compare_WrongResponseType(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
			Response: &extproctorv1.ExtProcExpectation_BodyResponse{
				BodyResponse: &extproctorv1.BodyExpectation{
					Body: []byte("body"),
				},
			},
		},
	}

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_RequestHeaders{
						RequestHeaders: &extprocv3.HeadersResponse{},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.False(t, compResult.Passed)
	assert.NotEmpty(t, compResult.Differences)
}

func TestComparator_Compare_ResponseHeaders(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_RESPONSE_HEADERS,
			Response: &extproctorv1.ExtProcExpectation_HeadersResponse{
				HeadersResponse: &extproctorv1.HeadersExpectation{
					SetHeaders: map[string]string{
						"x-response-header": "value",
					},
				},
			},
		},
	}

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_RESPONSE_HEADERS,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_ResponseHeaders{
						ResponseHeaders: &extprocv3.HeadersResponse{
							Response: &extprocv3.CommonResponse{
								HeaderMutation: &extprocv3.HeaderMutation{
									SetHeaders: []*corev3.HeaderValueOption{
										{
											Header: &corev3.HeaderValue{
												Key:   "x-response-header",
												Value: "value",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.True(t, compResult.Passed)
}

func TestComparator_Compare_ResponseBody(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_RESPONSE_BODY,
			Response: &extproctorv1.ExtProcExpectation_BodyResponse{
				BodyResponse: &extproctorv1.BodyExpectation{
					Body: []byte("response body"),
				},
			},
		},
	}

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_RESPONSE_BODY,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_ResponseBody{
						ResponseBody: &extprocv3.BodyResponse{
							Response: &extprocv3.CommonResponse{
								BodyMutation: &extprocv3.BodyMutation{
									Mutation: &extprocv3.BodyMutation_Body{
										Body: []byte("response body"),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.True(t, compResult.Passed)
}

func TestComparator_Compare_ResponseTrailers(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_RESPONSE_TRAILERS,
			Response: &extproctorv1.ExtProcExpectation_TrailersResponse{
				TrailersResponse: &extproctorv1.TrailersExpectation{
					SetTrailers: map[string]string{
						"x-trailer": "value",
					},
				},
			},
		},
	}

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_RESPONSE_TRAILERS,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_ResponseTrailers{
						ResponseTrailers: &extprocv3.TrailersResponse{
							HeaderMutation: &extprocv3.HeaderMutation{
								SetHeaders: []*corev3.HeaderValueOption{
									{
										Header: &corev3.HeaderValue{
											Key:   "x-trailer",
											Value: "value",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.True(t, compResult.Passed)
}

func TestComparator_CompareHeaderMutation_NilResponse(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
			Response: &extproctorv1.ExtProcExpectation_HeadersResponse{
				HeadersResponse: &extproctorv1.HeadersExpectation{
					CommonResponse: &extproctorv1.CommonResponse{
						HeaderMutation: &extproctorv1.HeaderMutation{
							SetHeaders: map[string]string{
								"x-header": "value",
							},
						},
					},
				},
			},
		},
	}

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_RequestHeaders{
						RequestHeaders: &extprocv3.HeadersResponse{
							Response: nil,
						},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.False(t, compResult.Passed)
	assert.NotEmpty(t, compResult.Differences)
}

func TestComparator_CompareRemoveHeaders_Match(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
			Response: &extproctorv1.ExtProcExpectation_HeadersResponse{
				HeadersResponse: &extproctorv1.HeadersExpectation{
					RemoveHeaders: []string{"x-remove-me"},
				},
			},
		},
	}

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_RequestHeaders{
						RequestHeaders: &extprocv3.HeadersResponse{
							Response: &extprocv3.CommonResponse{
								HeaderMutation: &extprocv3.HeaderMutation{
									RemoveHeaders: []string{"x-remove-me"},
								},
							},
						},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.True(t, compResult.Passed)
}

func TestComparator_CompareRemoveHeaders_Mismatch(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
			Response: &extproctorv1.ExtProcExpectation_HeadersResponse{
				HeadersResponse: &extproctorv1.HeadersExpectation{
					RemoveHeaders: []string{"x-remove-me"},
				},
			},
		},
	}

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_RequestHeaders{
						RequestHeaders: &extprocv3.HeadersResponse{
							Response: &extprocv3.CommonResponse{
								HeaderMutation: &extprocv3.HeaderMutation{
									RemoveHeaders: []string{"x-other"},
								},
							},
						},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.False(t, compResult.Passed)
	assert.NotEmpty(t, compResult.Differences)
}

func TestComparator_CompareRemoveHeaders_NilMutation(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
			Response: &extproctorv1.ExtProcExpectation_HeadersResponse{
				HeadersResponse: &extproctorv1.HeadersExpectation{
					RemoveHeaders: []string{"x-remove-me"},
				},
			},
		},
	}

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_RequestHeaders{
						RequestHeaders: &extprocv3.HeadersResponse{
							Response: nil,
						},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.False(t, compResult.Passed)
	assert.NotEmpty(t, compResult.Differences)
}

func TestComparator_CompareSetHeaders_NilMutation(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
			Response: &extproctorv1.ExtProcExpectation_HeadersResponse{
				HeadersResponse: &extproctorv1.HeadersExpectation{
					SetHeaders: map[string]string{
						"x-header": "value",
					},
				},
			},
		},
	}

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_RequestHeaders{
						RequestHeaders: &extprocv3.HeadersResponse{
							Response: nil,
						},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.False(t, compResult.Passed)
	assert.NotEmpty(t, compResult.Differences)
}

func TestComparator_CompareHeaderMutationRemove_Match(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
			Response: &extproctorv1.ExtProcExpectation_HeadersResponse{
				HeadersResponse: &extproctorv1.HeadersExpectation{
					CommonResponse: &extproctorv1.CommonResponse{
						HeaderMutation: &extproctorv1.HeaderMutation{
							RemoveHeaders: []string{"x-remove"},
						},
					},
				},
			},
		},
	}

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_RequestHeaders{
						RequestHeaders: &extprocv3.HeadersResponse{
							Response: &extprocv3.CommonResponse{
								HeaderMutation: &extprocv3.HeaderMutation{
									RemoveHeaders: []string{"x-remove"},
								},
							},
						},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.True(t, compResult.Passed)
}

func TestComparator_CompareHeaderMutationRemove_Mismatch(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
			Response: &extproctorv1.ExtProcExpectation_HeadersResponse{
				HeadersResponse: &extproctorv1.HeadersExpectation{
					CommonResponse: &extproctorv1.CommonResponse{
						HeaderMutation: &extproctorv1.HeaderMutation{
							RemoveHeaders: []string{"x-remove"},
						},
					},
				},
			},
		},
	}

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_RequestHeaders{
						RequestHeaders: &extprocv3.HeadersResponse{
							Response: &extprocv3.CommonResponse{
								HeaderMutation: &extprocv3.HeaderMutation{
									RemoveHeaders: []string{},
								},
							},
						},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.False(t, compResult.Passed)
	assert.NotEmpty(t, compResult.Differences)
}

func TestComparator_ImmediateResponse_WrongType(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
			Response: &extproctorv1.ExtProcExpectation_ImmediateResponse{
				ImmediateResponse: &extproctorv1.ImmediateExpectation{
					StatusCode: 403,
				},
			},
		},
	}

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_RequestHeaders{
						RequestHeaders: &extprocv3.HeadersResponse{},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.False(t, compResult.Passed)
	assert.NotEmpty(t, compResult.Differences)
}

func TestComparator_TrailersResponse_WrongType(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_TRAILERS,
			Response: &extproctorv1.ExtProcExpectation_TrailersResponse{
				TrailersResponse: &extproctorv1.TrailersExpectation{
					SetTrailers: map[string]string{
						"x-trailer": "value",
					},
				},
			},
		},
	}

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_TRAILERS,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_RequestHeaders{
						RequestHeaders: &extprocv3.HeadersResponse{},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.False(t, compResult.Passed)
	assert.NotEmpty(t, compResult.Differences)
}

func TestComparator_BodyResponse_WrongType(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_BODY,
			Response: &extproctorv1.ExtProcExpectation_BodyResponse{
				BodyResponse: &extproctorv1.BodyExpectation{
					Body: []byte("body"),
				},
			},
		},
	}

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_BODY,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_RequestHeaders{
						RequestHeaders: &extprocv3.HeadersResponse{},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.False(t, compResult.Passed)
	assert.NotEmpty(t, compResult.Differences)
}

func TestComparator_ClearBody_NotSet(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_BODY,
			Response: &extproctorv1.ExtProcExpectation_BodyResponse{
				BodyResponse: &extproctorv1.BodyExpectation{
					ClearBody: true,
				},
			},
		},
	}

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_BODY,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_RequestBody{
						RequestBody: &extprocv3.BodyResponse{
							Response: &extprocv3.CommonResponse{
								BodyMutation: &extprocv3.BodyMutation{
									Mutation: &extprocv3.BodyMutation_Body{
										Body: []byte("body"),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.False(t, compResult.Passed)
	assert.NotEmpty(t, compResult.Differences)
}

func TestComparator_HeaderMutation_SetHeaderMismatch(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
			Response: &extproctorv1.ExtProcExpectation_HeadersResponse{
				HeadersResponse: &extproctorv1.HeadersExpectation{
					CommonResponse: &extproctorv1.CommonResponse{
						HeaderMutation: &extproctorv1.HeaderMutation{
							SetHeaders: map[string]string{
								"x-header": "expected",
							},
						},
					},
				},
			},
		},
	}

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_RequestHeaders{
						RequestHeaders: &extprocv3.HeadersResponse{
							Response: &extprocv3.CommonResponse{
								HeaderMutation: &extprocv3.HeaderMutation{
									SetHeaders: []*corev3.HeaderValueOption{
										{
											Header: &corev3.HeaderValue{
												Key:   "x-header",
												Value: "actual",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.False(t, compResult.Passed)
	assert.NotEmpty(t, compResult.Differences)
}

func TestComparator_HeaderMutation_SetHeaderNotFound(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
			Response: &extproctorv1.ExtProcExpectation_HeadersResponse{
				HeadersResponse: &extproctorv1.HeadersExpectation{
					CommonResponse: &extproctorv1.CommonResponse{
						HeaderMutation: &extproctorv1.HeaderMutation{
							SetHeaders: map[string]string{
								"x-missing-header": "value",
							},
						},
					},
				},
			},
		},
	}

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_RequestHeaders{
						RequestHeaders: &extprocv3.HeadersResponse{
							Response: &extprocv3.CommonResponse{
								HeaderMutation: &extprocv3.HeaderMutation{
									SetHeaders: []*corev3.HeaderValueOption{},
								},
							},
						},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.False(t, compResult.Passed)
	assert.NotEmpty(t, compResult.Differences)
}

func TestComparator_ImmediateResponse_HeadersMatch(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
			Response: &extproctorv1.ExtProcExpectation_ImmediateResponse{
				ImmediateResponse: &extproctorv1.ImmediateExpectation{
					StatusCode: 403,
					Headers: map[string]string{
						"x-error": "forbidden",
					},
				},
			},
		},
	}

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_ImmediateResponse{
						ImmediateResponse: &extprocv3.ImmediateResponse{
							Status: &typev3.HttpStatus{
								Code: typev3.StatusCode_Forbidden,
							},
							Headers: &extprocv3.HeaderMutation{
								SetHeaders: []*corev3.HeaderValueOption{
									{
										Header: &corev3.HeaderValue{
											Key:   "x-error",
											Value: "forbidden",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.True(t, compResult.Passed)
}

func TestComparator_SetHeaders_HeaderNotFound(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
			Response: &extproctorv1.ExtProcExpectation_HeadersResponse{
				HeadersResponse: &extproctorv1.HeadersExpectation{
					SetHeaders: map[string]string{
						"x-missing": "value",
					},
				},
			},
		},
	}

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_RequestHeaders{
						RequestHeaders: &extprocv3.HeadersResponse{
							Response: &extprocv3.CommonResponse{
								HeaderMutation: &extprocv3.HeaderMutation{
									SetHeaders: []*corev3.HeaderValueOption{
										{
											Header: &corev3.HeaderValue{
												Key:   "x-other",
												Value: "value",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.False(t, compResult.Passed)
	assert.NotEmpty(t, compResult.Differences)
}

func TestComparator_ImmediateResponse_NilStatus(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
			Response: &extproctorv1.ExtProcExpectation_ImmediateResponse{
				ImmediateResponse: &extproctorv1.ImmediateExpectation{
					StatusCode: 403,
				},
			},
		},
	}

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_ImmediateResponse{
						ImmediateResponse: &extprocv3.ImmediateResponse{
							Status: nil,
						},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.False(t, compResult.Passed)
	assert.NotEmpty(t, compResult.Differences)
}

func TestComparator_HeadersResponse_NilActual(t *testing.T) {
	comp := New()

	expectations := []*extproctorv1.ExtProcExpectation{
		{
			Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
			Response: &extproctorv1.ExtProcExpectation_HeadersResponse{
				HeadersResponse: &extproctorv1.HeadersExpectation{
					SetHeaders: map[string]string{
						"x-header": "value",
					},
				},
			},
		},
	}

	// Response is not a headers response
	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_ImmediateResponse{
						ImmediateResponse: &extprocv3.ImmediateResponse{},
					},
				},
			},
		},
	}

	compResult := comp.Compare(expectations, result)
	assert.False(t, compResult.Passed)
	assert.NotEmpty(t, compResult.Differences)
}
