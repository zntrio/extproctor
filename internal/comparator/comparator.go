// SPDX-FileCopyrightText: 2025 Thibault NORMAND
// SPDX-License-Identifier: MIT

package comparator

import (
	"fmt"
	"strings"

	extprocv3 "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"

	extproctorv1 "zntr.io/extproctor/gen/extproctor/v1"
	"zntr.io/extproctor/internal/client"
)

// ComparisonResult contains the result of comparing expected vs actual responses.
type ComparisonResult struct {
	Passed      bool
	Differences []Difference
	Matched     []*MatchedExpectation
	Unmatched   []*extproctorv1.ExtProcExpectation
}

// MatchedExpectation represents an expectation that was matched.
type MatchedExpectation struct {
	Expectation *extproctorv1.ExtProcExpectation
	Response    *client.PhaseResponse
}

// Difference represents a single difference between expected and actual values.
type Difference struct {
	Phase    extproctorv1.ProcessingPhase
	Path     string
	Expected string
	Actual   string
}

// Comparator compares expected expectations against actual responses.
type Comparator struct{}

// New creates a new comparator.
func New() *Comparator {
	return &Comparator{}
}

// Compare compares expectations against actual responses using unordered matching.
// All expectations must be satisfied by some response.
func (c *Comparator) Compare(expectations []*extproctorv1.ExtProcExpectation, result *client.ProcessingResult) *ComparisonResult {
	cr := &ComparisonResult{
		Passed: true,
	}

	// Track which expectations have been matched
	matchedExpectations := make(map[int]bool)

	// Try to match each expectation with a response
	for i, exp := range expectations {
		matched := false

		for _, resp := range result.Responses {
			// Phase must match
			if resp.Phase != exp.Phase {
				continue
			}

			// Try to match this expectation with this response
			diffs := c.compareExpectation(exp, resp.Response)
			if len(diffs) == 0 {
				// Match found
				matched = true
				matchedExpectations[i] = true
				cr.Matched = append(cr.Matched, &MatchedExpectation{
					Expectation: exp,
					Response:    resp,
				})
				break
			} else {
				// Record differences but continue looking for a match
				cr.Differences = append(cr.Differences, diffs...)
			}
		}

		if !matched {
			cr.Unmatched = append(cr.Unmatched, exp)
			cr.Passed = false
		}
	}

	return cr
}

// compareExpectation compares a single expectation against a response.
func (c *Comparator) compareExpectation(exp *extproctorv1.ExtProcExpectation, resp *extprocv3.ProcessingResponse) []Difference {
	var diffs []Difference

	switch r := exp.Response.(type) {
	case *extproctorv1.ExtProcExpectation_HeadersResponse:
		diffs = c.compareHeadersResponse(exp.Phase, r.HeadersResponse, resp)
	case *extproctorv1.ExtProcExpectation_BodyResponse:
		diffs = c.compareBodyResponse(exp.Phase, r.BodyResponse, resp)
	case *extproctorv1.ExtProcExpectation_TrailersResponse:
		diffs = c.compareTrailersResponse(exp.Phase, r.TrailersResponse, resp)
	case *extproctorv1.ExtProcExpectation_ImmediateResponse:
		diffs = c.compareImmediateResponse(exp.Phase, r.ImmediateResponse, resp)
	}

	return diffs
}

// compareHeadersResponse compares expected headers response against actual.
func (c *Comparator) compareHeadersResponse(phase extproctorv1.ProcessingPhase, exp *extproctorv1.HeadersExpectation, resp *extprocv3.ProcessingResponse) []Difference {
	var diffs []Difference

	actual := resp.GetRequestHeaders()
	if actual == nil {
		actual = resp.GetResponseHeaders()
	}

	if actual == nil {
		diffs = append(diffs, Difference{
			Phase:    phase,
			Path:     "response_type",
			Expected: "headers_response",
			Actual:   fmt.Sprintf("%T", resp.Response),
		})
		return diffs
	}

	// Compare header mutations
	if exp.CommonResponse != nil && exp.CommonResponse.HeaderMutation != nil {
		diffs = append(diffs, c.compareHeaderMutation(phase, exp.CommonResponse.HeaderMutation, actual.Response)...)
	}

	// Compare set headers
	if len(exp.SetHeaders) > 0 {
		diffs = append(diffs, c.compareSetHeaders(phase, exp.SetHeaders, actual.Response)...)
	}

	// Compare remove headers
	if len(exp.RemoveHeaders) > 0 {
		diffs = append(diffs, c.compareRemoveHeaders(phase, exp.RemoveHeaders, actual.Response)...)
	}

	return diffs
}

// compareHeaderMutation compares header mutation expectations.
func (c *Comparator) compareHeaderMutation(phase extproctorv1.ProcessingPhase, exp *extproctorv1.HeaderMutation, resp *extprocv3.CommonResponse) []Difference {
	var diffs []Difference

	if resp == nil || resp.HeaderMutation == nil {
		if len(exp.SetHeaders) > 0 || len(exp.RemoveHeaders) > 0 {
			diffs = append(diffs, Difference{
				Phase:    phase,
				Path:     "header_mutation",
				Expected: "present",
				Actual:   "nil",
			})
		}
		return diffs
	}

	// Compare set headers
	for k, v := range exp.SetHeaders {
		found := false
		for _, h := range resp.HeaderMutation.SetHeaders {
			if h.Header != nil && h.Header.Key == k {
				found = true
				if h.Header.Value != v {
					diffs = append(diffs, Difference{
						Phase:    phase,
						Path:     fmt.Sprintf("header_mutation.set_headers[%s]", k),
						Expected: v,
						Actual:   h.Header.Value,
					})
				}
				break
			}
		}
		if !found {
			diffs = append(diffs, Difference{
				Phase:    phase,
				Path:     fmt.Sprintf("header_mutation.set_headers[%s]", k),
				Expected: v,
				Actual:   "<not set>",
			})
		}
	}

	// Compare remove headers
	for _, k := range exp.RemoveHeaders {
		found := false
		for _, h := range resp.HeaderMutation.RemoveHeaders {
			if h == k {
				found = true
				break
			}
		}
		if !found {
			diffs = append(diffs, Difference{
				Phase:    phase,
				Path:     fmt.Sprintf("header_mutation.remove_headers[%s]", k),
				Expected: "removed",
				Actual:   "<not removed>",
			})
		}
	}

	return diffs
}

// compareSetHeaders compares set headers expectations.
func (c *Comparator) compareSetHeaders(phase extproctorv1.ProcessingPhase, exp map[string]string, resp *extprocv3.CommonResponse) []Difference {
	var diffs []Difference

	if resp == nil || resp.HeaderMutation == nil {
		if len(exp) > 0 {
			diffs = append(diffs, Difference{
				Phase:    phase,
				Path:     "set_headers",
				Expected: fmt.Sprintf("%v", exp),
				Actual:   "<no header mutation>",
			})
		}
		return diffs
	}

	for k, v := range exp {
		found := false
		for _, h := range resp.HeaderMutation.SetHeaders {
			if h.Header != nil && h.Header.Key == k {
				found = true
				if h.Header.Value != v {
					diffs = append(diffs, Difference{
						Phase:    phase,
						Path:     fmt.Sprintf("set_headers[%s]", k),
						Expected: v,
						Actual:   h.Header.Value,
					})
				}
				break
			}
		}
		if !found {
			diffs = append(diffs, Difference{
				Phase:    phase,
				Path:     fmt.Sprintf("set_headers[%s]", k),
				Expected: v,
				Actual:   "<not set>",
			})
		}
	}

	return diffs
}

// compareRemoveHeaders compares remove headers expectations.
func (c *Comparator) compareRemoveHeaders(phase extproctorv1.ProcessingPhase, exp []string, resp *extprocv3.CommonResponse) []Difference {
	var diffs []Difference

	if resp == nil || resp.HeaderMutation == nil {
		if len(exp) > 0 {
			diffs = append(diffs, Difference{
				Phase:    phase,
				Path:     "remove_headers",
				Expected: strings.Join(exp, ", "),
				Actual:   "<no header mutation>",
			})
		}
		return diffs
	}

	for _, k := range exp {
		found := false
		for _, h := range resp.HeaderMutation.RemoveHeaders {
			if h == k {
				found = true
				break
			}
		}
		if !found {
			diffs = append(diffs, Difference{
				Phase:    phase,
				Path:     fmt.Sprintf("remove_headers[%s]", k),
				Expected: "removed",
				Actual:   "<not removed>",
			})
		}
	}

	return diffs
}

// compareBodyResponse compares expected body response against actual.
func (c *Comparator) compareBodyResponse(phase extproctorv1.ProcessingPhase, exp *extproctorv1.BodyExpectation, resp *extprocv3.ProcessingResponse) []Difference {
	var diffs []Difference

	actual := resp.GetRequestBody()
	if actual == nil {
		actual = resp.GetResponseBody()
	}

	if actual == nil {
		diffs = append(diffs, Difference{
			Phase:    phase,
			Path:     "response_type",
			Expected: "body_response",
			Actual:   fmt.Sprintf("%T", resp.Response),
		})
		return diffs
	}

	if exp.ClearBody && actual.Response != nil {
		bodyMut := actual.Response.BodyMutation
		if bodyMut == nil || !bodyMut.GetClearBody() {
			diffs = append(diffs, Difference{
				Phase:    phase,
				Path:     "body.clear_body",
				Expected: "true",
				Actual:   "false",
			})
		}
	}

	if len(exp.Body) > 0 && actual.Response != nil {
		bodyMut := actual.Response.BodyMutation
		if bodyMut == nil {
			diffs = append(diffs, Difference{
				Phase:    phase,
				Path:     "body.body_mutation",
				Expected: string(exp.Body),
				Actual:   "<nil>",
			})
		} else if string(bodyMut.GetBody()) != string(exp.Body) {
			diffs = append(diffs, Difference{
				Phase:    phase,
				Path:     "body.body_mutation.body",
				Expected: string(exp.Body),
				Actual:   string(bodyMut.GetBody()),
			})
		}
	}

	return diffs
}

// compareTrailersResponse compares expected trailers response against actual.
func (c *Comparator) compareTrailersResponse(phase extproctorv1.ProcessingPhase, exp *extproctorv1.TrailersExpectation, resp *extprocv3.ProcessingResponse) []Difference {
	var diffs []Difference

	actual := resp.GetRequestTrailers()
	if actual == nil {
		actual = resp.GetResponseTrailers()
	}

	if actual == nil {
		diffs = append(diffs, Difference{
			Phase:    phase,
			Path:     "response_type",
			Expected: "trailers_response",
			Actual:   fmt.Sprintf("%T", resp.Response),
		})
		return diffs
	}

	// Compare set trailers
	if len(exp.SetTrailers) > 0 && actual.HeaderMutation != nil {
		for k, v := range exp.SetTrailers {
			found := false
			for _, h := range actual.HeaderMutation.SetHeaders {
				if h.Header != nil && h.Header.Key == k {
					found = true
					if h.Header.Value != v {
						diffs = append(diffs, Difference{
							Phase:    phase,
							Path:     fmt.Sprintf("set_trailers[%s]", k),
							Expected: v,
							Actual:   h.Header.Value,
						})
					}
					break
				}
			}
			if !found {
				diffs = append(diffs, Difference{
					Phase:    phase,
					Path:     fmt.Sprintf("set_trailers[%s]", k),
					Expected: v,
					Actual:   "<not set>",
				})
			}
		}
	}

	return diffs
}

// compareImmediateResponse compares expected immediate response against actual.
func (c *Comparator) compareImmediateResponse(phase extproctorv1.ProcessingPhase, exp *extproctorv1.ImmediateExpectation, resp *extprocv3.ProcessingResponse) []Difference {
	var diffs []Difference

	actual := resp.GetImmediateResponse()
	if actual == nil {
		diffs = append(diffs, Difference{
			Phase:    phase,
			Path:     "response_type",
			Expected: "immediate_response",
			Actual:   fmt.Sprintf("%T", resp.Response),
		})
		return diffs
	}

	// Compare status code
	if exp.StatusCode > 0 {
		actualStatus := int32(0)
		if actual.Status != nil {
			actualStatus = int32(actual.Status.Code)
		}
		if actualStatus != exp.StatusCode {
			diffs = append(diffs, Difference{
				Phase:    phase,
				Path:     "immediate_response.status_code",
				Expected: fmt.Sprintf("%d", exp.StatusCode),
				Actual:   fmt.Sprintf("%d", actualStatus),
			})
		}
	}

	// Compare body
	if len(exp.Body) > 0 {
		if string(actual.Body) != string(exp.Body) {
			diffs = append(diffs, Difference{
				Phase:    phase,
				Path:     "immediate_response.body",
				Expected: string(exp.Body),
				Actual:   string(actual.Body),
			})
		}
	}

	// Compare headers
	if len(exp.Headers) > 0 && actual.Headers != nil {
		for k, v := range exp.Headers {
			found := false
			for _, h := range actual.Headers.SetHeaders {
				if h.Header != nil && h.Header.Key == k {
					found = true
					if h.Header.Value != v {
						diffs = append(diffs, Difference{
							Phase:    phase,
							Path:     fmt.Sprintf("immediate_response.headers[%s]", k),
							Expected: v,
							Actual:   h.Header.Value,
						})
					}
					break
				}
			}
			if !found {
				diffs = append(diffs, Difference{
					Phase:    phase,
					Path:     fmt.Sprintf("immediate_response.headers[%s]", k),
					Expected: v,
					Actual:   "<not set>",
				})
			}
		}
	}

	return diffs
}
