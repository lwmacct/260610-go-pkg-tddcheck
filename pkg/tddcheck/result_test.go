package tddcheck

import (
	"strings"
	"testing"
)

func TestResult_Text(t *testing.T) {
	result := Result{}
	if result.Text() != "tddcheck: passed" {
		t.Fatalf("unexpected pass text: %q", result.Text())
	}

	result.Findings = []Finding{{
		Rule:     "rule",
		Severity: SeverityError,
		File:     "file.go",
		Line:     7,
		Message:  "message",
	}}
	text := result.Text()
	if !strings.Contains(text, "file.go:7") || !strings.Contains(text, "message") {
		t.Fatalf("unexpected finding text: %q", text)
	}
}

func TestFinding_String(t *testing.T) {
	finding := Finding{
		Rule:     "rule",
		Severity: SeverityError,
		File:     "file.go",
		Line:     1,
		Message:  "message",
	}
	if !strings.Contains(finding.String(), "file.go:1") {
		t.Fatalf("unexpected finding string: %q", finding.String())
	}
}
