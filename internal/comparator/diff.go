// SPDX-FileCopyrightText: 2025 Thibault NORMAND
// SPDX-License-Identifier: MIT

package comparator

import (
	"fmt"
	"strings"

	extproctorv1 "zntr.io/extproctor/gen/extproctor/v1"
)

// FormatDifferences formats differences for human-readable output.
func FormatDifferences(diffs []Difference) string {
	if len(diffs) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("Differences:\n")

	for _, d := range diffs {
		sb.WriteString(fmt.Sprintf("  [%s] %s:\n", phaseName(d.Phase), d.Path))
		sb.WriteString(fmt.Sprintf("    expected: %s\n", d.Expected))
		sb.WriteString(fmt.Sprintf("    actual:   %s\n", d.Actual))
	}

	return sb.String()
}

// FormatUnmatched formats unmatched expectations for human-readable output.
func FormatUnmatched(unmatched []*extproctorv1.ExtProcExpectation) string {
	if len(unmatched) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("Unmatched expectations:\n")

	for _, exp := range unmatched {
		sb.WriteString(fmt.Sprintf("  - Phase: %s\n", phaseName(exp.Phase)))
		sb.WriteString(fmt.Sprintf("    Type: %T\n", exp.Response))
	}

	return sb.String()
}

// phaseName returns a human-readable name for a processing phase.
func phaseName(phase extproctorv1.ProcessingPhase) string {
	switch phase {
	case extproctorv1.ProcessingPhase_REQUEST_HEADERS:
		return "REQUEST_HEADERS"
	case extproctorv1.ProcessingPhase_REQUEST_BODY:
		return "REQUEST_BODY"
	case extproctorv1.ProcessingPhase_REQUEST_TRAILERS:
		return "REQUEST_TRAILERS"
	case extproctorv1.ProcessingPhase_RESPONSE_HEADERS:
		return "RESPONSE_HEADERS"
	case extproctorv1.ProcessingPhase_RESPONSE_BODY:
		return "RESPONSE_BODY"
	case extproctorv1.ProcessingPhase_RESPONSE_TRAILERS:
		return "RESPONSE_TRAILERS"
	default:
		return "UNKNOWN"
	}
}
