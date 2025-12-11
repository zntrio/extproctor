// SPDX-FileCopyrightText: 2025 Thibault NORMAND
// SPDX-License-Identifier: MIT

package golden

import (
	"os"
	"path/filepath"
	"testing"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extprocv3 "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	extproctorv1 "zntr.io/extproctor/gen/extproctor/v1"
	"zntr.io/extproctor/internal/client"
)

func TestWrite_RequestHeaders(t *testing.T) {
	tmpDir := t.TempDir()
	goldenPath := filepath.Join(tmpDir, "golden.textproto")

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
									RemoveHeaders: []string{"x-remove-me"},
								},
							},
						},
					},
				},
			},
		},
	}

	err := Write(goldenPath, result)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(goldenPath)
	require.NoError(t, err)

	// Read back and verify
	expectations, err := Read(goldenPath)
	require.NoError(t, err)
	assert.Len(t, expectations, 1)
	assert.Equal(t, extproctorv1.ProcessingPhase_REQUEST_HEADERS, expectations[0].Phase)
}

func TestWrite_ResponseHeaders(t *testing.T) {
	tmpDir := t.TempDir()
	goldenPath := filepath.Join(tmpDir, "golden.textproto")

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
												Value: "response-value",
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

	err := Write(goldenPath, result)
	require.NoError(t, err)

	expectations, err := Read(goldenPath)
	require.NoError(t, err)
	assert.Len(t, expectations, 1)
}

func TestWrite_RequestBody(t *testing.T) {
	tmpDir := t.TempDir()
	goldenPath := filepath.Join(tmpDir, "golden.textproto")

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

	err := Write(goldenPath, result)
	require.NoError(t, err)

	expectations, err := Read(goldenPath)
	require.NoError(t, err)
	assert.Len(t, expectations, 1)
}

func TestWrite_ResponseBody(t *testing.T) {
	tmpDir := t.TempDir()
	goldenPath := filepath.Join(tmpDir, "golden.textproto")

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_RESPONSE_BODY,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_ResponseBody{
						ResponseBody: &extprocv3.BodyResponse{
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

	err := Write(goldenPath, result)
	require.NoError(t, err)

	expectations, err := Read(goldenPath)
	require.NoError(t, err)
	assert.Len(t, expectations, 1)
}

func TestWrite_RequestTrailers(t *testing.T) {
	tmpDir := t.TempDir()
	goldenPath := filepath.Join(tmpDir, "golden.textproto")

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
											Value: "trailer-value",
										},
									},
								},
								RemoveHeaders: []string{"x-remove-trailer"},
							},
						},
					},
				},
			},
		},
	}

	err := Write(goldenPath, result)
	require.NoError(t, err)

	expectations, err := Read(goldenPath)
	require.NoError(t, err)
	assert.Len(t, expectations, 1)
}

func TestWrite_ResponseTrailers(t *testing.T) {
	tmpDir := t.TempDir()
	goldenPath := filepath.Join(tmpDir, "golden.textproto")

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
											Key:   "x-response-trailer",
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

	err := Write(goldenPath, result)
	require.NoError(t, err)

	expectations, err := Read(goldenPath)
	require.NoError(t, err)
	assert.Len(t, expectations, 1)
}

func TestWrite_ImmediateResponse(t *testing.T) {
	tmpDir := t.TempDir()
	goldenPath := filepath.Join(tmpDir, "golden.textproto")

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
							Body:    []byte("Forbidden"),
							Details: "Access denied",
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
							GrpcStatus: &extprocv3.GrpcStatus{
								Status: 7,
							},
						},
					},
				},
			},
		},
	}

	err := Write(goldenPath, result)
	require.NoError(t, err)

	expectations, err := Read(goldenPath)
	require.NoError(t, err)
	assert.Len(t, expectations, 1)
}

func TestWrite_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	goldenPath := filepath.Join(tmpDir, "subdir", "nested", "golden.textproto")

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

	err := Write(goldenPath, result)
	require.NoError(t, err)

	// Verify directory was created
	_, err = os.Stat(filepath.Join(tmpDir, "subdir", "nested"))
	require.NoError(t, err)
}

func TestWrite_EmptyResult(t *testing.T) {
	tmpDir := t.TempDir()
	goldenPath := filepath.Join(tmpDir, "golden.textproto")

	result := &client.ProcessingResult{
		Responses: []*client.PhaseResponse{},
	}

	err := Write(goldenPath, result)
	require.NoError(t, err)

	expectations, err := Read(goldenPath)
	require.NoError(t, err)
	assert.Empty(t, expectations)
}

func TestRead_NonExistent(t *testing.T) {
	_, err := Read("/nonexistent/path/golden.textproto")
	assert.Error(t, err)
}

func TestRead_InvalidPrototext(t *testing.T) {
	tmpDir := t.TempDir()
	goldenPath := filepath.Join(tmpDir, "invalid.textproto")

	err := os.WriteFile(goldenPath, []byte("invalid { prototext"), 0o644)
	require.NoError(t, err)

	_, err = Read(goldenPath)
	assert.Error(t, err)
}

func TestWrite_NilHeaderInResponse(t *testing.T) {
	tmpDir := t.TempDir()
	goldenPath := filepath.Join(tmpDir, "golden.textproto")

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
										{Header: nil}, // nil header
										{
											Header: &corev3.HeaderValue{
												Key:   "x-valid",
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

	err := Write(goldenPath, result)
	require.NoError(t, err)

	expectations, err := Read(goldenPath)
	require.NoError(t, err)
	assert.Len(t, expectations, 1)
}

func TestConvertEnvoyHeadersResponse_NilResponse(t *testing.T) {
	result := convertEnvoyHeadersResponse(nil)
	assert.NotNil(t, result)
	assert.NotNil(t, result.HeadersResponse)
}

func TestConvertEnvoyBodyResponse_NilResponse(t *testing.T) {
	result := convertEnvoyBodyResponse(nil)
	assert.NotNil(t, result)
	assert.NotNil(t, result.BodyResponse)
}

func TestConvertEnvoyTrailersResponse_NilResponse(t *testing.T) {
	result := convertEnvoyTrailersResponse(nil)
	assert.NotNil(t, result)
	assert.NotNil(t, result.TrailersResponse)
}

func TestConvertEnvoyImmediateResponse_NilResponse(t *testing.T) {
	result := convertEnvoyImmediateResponse(nil)
	assert.NotNil(t, result)
	assert.NotNil(t, result.ImmediateResponse)
}

func TestWrite_MultipleResponses(t *testing.T) {
	tmpDir := t.TempDir()
	goldenPath := filepath.Join(tmpDir, "golden.textproto")

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
												Key:   "x-header-1",
												Value: "value-1",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_BODY,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_RequestBody{
						RequestBody: &extprocv3.BodyResponse{
							Response: &extprocv3.CommonResponse{
								BodyMutation: &extprocv3.BodyMutation{
									Mutation: &extprocv3.BodyMutation_Body{
										Body: []byte("body content"),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	err := Write(goldenPath, result)
	require.NoError(t, err)

	expectations, err := Read(goldenPath)
	require.NoError(t, err)
	assert.Len(t, expectations, 2)
}

func TestConvertEnvoyBodyResponse_ClearBody(t *testing.T) {
	resp := &extprocv3.CommonResponse{
		BodyMutation: &extprocv3.BodyMutation{
			Mutation: &extprocv3.BodyMutation_ClearBody{
				ClearBody: true,
			},
		},
	}

	result := convertEnvoyBodyResponse(resp)
	assert.True(t, result.BodyResponse.ClearBody)
}

func TestConvertEnvoyTrailersResponse_NilMutation(t *testing.T) {
	resp := &extprocv3.TrailersResponse{
		HeaderMutation: nil,
	}

	result := convertEnvoyTrailersResponse(resp)
	assert.NotNil(t, result)
	assert.Nil(t, result.TrailersResponse.SetTrailers)
}

func TestConvertEnvoyImmediateResponse_AllFields(t *testing.T) {
	resp := &extprocv3.ImmediateResponse{
		Status: &typev3.HttpStatus{
			Code: typev3.StatusCode_Unauthorized,
		},
		Body:    []byte("unauthorized"),
		Details: "auth failed",
		Headers: &extprocv3.HeaderMutation{
			SetHeaders: []*corev3.HeaderValueOption{
				{
					Header: &corev3.HeaderValue{
						Key:   "www-authenticate",
						Value: "Bearer",
					},
				},
			},
		},
		GrpcStatus: &extprocv3.GrpcStatus{
			Status: 16,
		},
	}

	result := convertEnvoyImmediateResponse(resp)
	assert.Equal(t, int32(401), result.ImmediateResponse.StatusCode)
	assert.Equal(t, []byte("unauthorized"), result.ImmediateResponse.Body)
	assert.Equal(t, "auth failed", result.ImmediateResponse.Details)
	assert.Contains(t, result.ImmediateResponse.Headers, "www-authenticate")
	assert.Equal(t, int32(16), result.ImmediateResponse.GrpcStatus.Status)
}

func TestConvertEnvoyImmediateResponse_NilHeaderInMutation(t *testing.T) {
	resp := &extprocv3.ImmediateResponse{
		Headers: &extprocv3.HeaderMutation{
			SetHeaders: []*corev3.HeaderValueOption{
				{Header: nil},
				{
					Header: &corev3.HeaderValue{
						Key:   "x-valid",
						Value: "value",
					},
				},
			},
		},
	}

	result := convertEnvoyImmediateResponse(resp)
	assert.NotNil(t, result.ImmediateResponse.Headers)
	assert.Contains(t, result.ImmediateResponse.Headers, "x-valid")
}

func TestWrite_InvalidPath(t *testing.T) {
	// Try to write to a path that can't be created
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

	// Use /dev/null as parent which can't have subdirectories
	err := Write("/dev/null/subdir/golden.textproto", result)
	assert.Error(t, err)
}

func TestWrite_ReadOnlyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	err := os.MkdirAll(readOnlyDir, 0o755)
	require.NoError(t, err)

	// Make directory read-only
	err = os.Chmod(readOnlyDir, 0o555)
	require.NoError(t, err)
	defer func() { _ = os.Chmod(readOnlyDir, 0o755) }() // Restore for cleanup

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

	err = Write(filepath.Join(readOnlyDir, "golden.textproto"), result)
	assert.Error(t, err)
}

func TestConvertEnvoyTrailersResponse_WithNilHeader(t *testing.T) {
	resp := &extprocv3.TrailersResponse{
		HeaderMutation: &extprocv3.HeaderMutation{
			SetHeaders: []*corev3.HeaderValueOption{
				{Header: nil},
				{
					Header: &corev3.HeaderValue{
						Key:   "x-trailer",
						Value: "value",
					},
				},
			},
			RemoveHeaders: []string{"x-remove"},
		},
	}

	result := convertEnvoyTrailersResponse(resp)
	assert.Contains(t, result.TrailersResponse.SetTrailers, "x-trailer")
	assert.Contains(t, result.TrailersResponse.RemoveTrailers, "x-remove")
}

func TestConvertEnvoyHeadersResponse_WithNilHeader(t *testing.T) {
	resp := &extprocv3.CommonResponse{
		HeaderMutation: &extprocv3.HeaderMutation{
			SetHeaders: []*corev3.HeaderValueOption{
				{Header: nil},
				{
					Header: &corev3.HeaderValue{
						Key:   "x-valid",
						Value: "value",
					},
				},
			},
			RemoveHeaders: []string{"x-remove"},
		},
	}

	result := convertEnvoyHeadersResponse(resp)
	assert.Contains(t, result.HeadersResponse.SetHeaders, "x-valid")
	assert.Contains(t, result.HeadersResponse.RemoveHeaders, "x-remove")
}
