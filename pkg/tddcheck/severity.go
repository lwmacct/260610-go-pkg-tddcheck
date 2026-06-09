package tddcheck

// Severity describes how strongly a finding should affect the check result.
type Severity string

const (
	// SeverityError marks a finding that fails the check.
	SeverityError Severity = "error"
	// SeverityWarning marks a finding that should be reported but does not fail.
	SeverityWarning Severity = "warning"
)
