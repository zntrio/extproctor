// SPDX-FileCopyrightText: 2025 Thibault NORMAND
// SPDX-License-Identifier: MIT

package golden

import (
	"fmt"
	"os"
	"path/filepath"

	extprocv3 "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	"google.golang.org/protobuf/encoding/prototext"

	extproctorv1 "zntr.io/extproctor/gen/extproctor/v1"
	"zntr.io/extproctor/internal/client"
)

// Write writes the processing result as a golden file.
func Write(path string, result *client.ProcessingResult) error {
	expectations := convertToExpectations(result)

	// Create wrapper message for serialization
	wrapper := &extproctorv1.TestCase{
		Name:         "golden",
		Expectations: expectations,
	}

	data, err := prototext.MarshalOptions{
		Multiline: true,
		Indent:    "  ",
	}.Marshal(wrapper)
	if err != nil {
		return fmt.Errorf("failed to marshal golden file: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write golden file: %w", err)
	}

	return nil
}

// Read reads expectations from a golden file.
func Read(path string) ([]*extproctorv1.ExtProcExpectation, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read golden file: %w", err)
	}

	wrapper := &extproctorv1.TestCase{}
	if err := prototext.Unmarshal(data, wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse golden file: %w", err)
	}

	return wrapper.Expectations, nil
}

// convertToExpectations converts processing results to expectations.
func convertToExpectations(result *client.ProcessingResult) []*extproctorv1.ExtProcExpectation {
	expectations := make([]*extproctorv1.ExtProcExpectation, 0, len(result.Responses))

	for _, resp := range result.Responses {
		exp := &extproctorv1.ExtProcExpectation{
			Phase: resp.Phase,
		}

		// Convert the response based on type
		switch {
		case resp.Response.GetRequestHeaders() != nil:
			exp.Response = convertEnvoyHeadersResponse(resp.Response.GetRequestHeaders().Response)
		case resp.Response.GetResponseHeaders() != nil:
			exp.Response = convertEnvoyHeadersResponse(resp.Response.GetResponseHeaders().Response)
		case resp.Response.GetRequestBody() != nil:
			exp.Response = convertEnvoyBodyResponse(resp.Response.GetRequestBody().Response)
		case resp.Response.GetResponseBody() != nil:
			exp.Response = convertEnvoyBodyResponse(resp.Response.GetResponseBody().Response)
		case resp.Response.GetRequestTrailers() != nil:
			exp.Response = convertEnvoyTrailersResponse(resp.Response.GetRequestTrailers())
		case resp.Response.GetResponseTrailers() != nil:
			exp.Response = convertEnvoyTrailersResponse(resp.Response.GetResponseTrailers())
		case resp.Response.GetImmediateResponse() != nil:
			exp.Response = convertEnvoyImmediateResponse(resp.Response.GetImmediateResponse())
		}

		expectations = append(expectations, exp)
	}

	return expectations
}

// convertEnvoyHeadersResponse converts an ExtProc headers response to our expectation format.
func convertEnvoyHeadersResponse(resp *extprocv3.CommonResponse) *extproctorv1.ExtProcExpectation_HeadersResponse {
	if resp == nil {
		return &extproctorv1.ExtProcExpectation_HeadersResponse{
			HeadersResponse: &extproctorv1.HeadersExpectation{},
		}
	}

	headersExp := &extproctorv1.HeadersExpectation{}

	// Convert header mutations
	if resp.HeaderMutation != nil {
		headersExp.SetHeaders = make(map[string]string)
		for _, h := range resp.HeaderMutation.SetHeaders {
			if h.Header != nil {
				headersExp.SetHeaders[h.Header.Key] = h.Header.Value
			}
		}
		headersExp.RemoveHeaders = resp.HeaderMutation.RemoveHeaders
	}

	return &extproctorv1.ExtProcExpectation_HeadersResponse{
		HeadersResponse: headersExp,
	}
}

// convertEnvoyBodyResponse converts an ExtProc body response to our expectation format.
func convertEnvoyBodyResponse(resp *extprocv3.CommonResponse) *extproctorv1.ExtProcExpectation_BodyResponse {
	if resp == nil {
		return &extproctorv1.ExtProcExpectation_BodyResponse{
			BodyResponse: &extproctorv1.BodyExpectation{},
		}
	}

	bodyExp := &extproctorv1.BodyExpectation{}

	if resp.BodyMutation != nil {
		bodyExp.Body = resp.BodyMutation.GetBody()
		bodyExp.ClearBody = resp.BodyMutation.GetClearBody()
	}

	return &extproctorv1.ExtProcExpectation_BodyResponse{
		BodyResponse: bodyExp,
	}
}

// convertEnvoyTrailersResponse converts an ExtProc trailers response to our expectation format.
func convertEnvoyTrailersResponse(resp *extprocv3.TrailersResponse) *extproctorv1.ExtProcExpectation_TrailersResponse {
	trailersExp := &extproctorv1.TrailersExpectation{}

	if resp != nil && resp.HeaderMutation != nil {
		trailersExp.SetTrailers = make(map[string]string)
		for _, h := range resp.HeaderMutation.SetHeaders {
			if h.Header != nil {
				trailersExp.SetTrailers[h.Header.Key] = h.Header.Value
			}
		}
		trailersExp.RemoveTrailers = resp.HeaderMutation.RemoveHeaders
	}

	return &extproctorv1.ExtProcExpectation_TrailersResponse{
		TrailersResponse: trailersExp,
	}
}

// convertEnvoyImmediateResponse converts an ExtProc immediate response to our expectation format.
func convertEnvoyImmediateResponse(resp *extprocv3.ImmediateResponse) *extproctorv1.ExtProcExpectation_ImmediateResponse {
	immExp := &extproctorv1.ImmediateExpectation{}

	if resp != nil {
		if resp.Status != nil {
			immExp.StatusCode = int32(resp.Status.Code)
		}
		immExp.Body = resp.Body
		immExp.Details = resp.Details

		if resp.Headers != nil {
			immExp.Headers = make(map[string]string)
			for _, h := range resp.Headers.SetHeaders {
				if h.Header != nil {
					immExp.Headers[h.Header.Key] = h.Header.Value
				}
			}
		}

		if resp.GrpcStatus != nil {
			immExp.GrpcStatus = &extproctorv1.GrpcStatus{
				Status: int32(resp.GrpcStatus.Status),
			}
		}
	}

	return &extproctorv1.ExtProcExpectation_ImmediateResponse{
		ImmediateResponse: immExp,
	}
}
