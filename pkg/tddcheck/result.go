package tddcheck

import (
	"strings"
	"time"
)

// Result is the output of a check run.
type Result struct {
	Passed   bool
	Err      error
	Findings []Finding
	Duration time.Duration
}

// Text renders findings in a stable, line-oriented format suitable for CLI use.
func (r Result) Text() string {
	if r.Err != nil {
		return "tddcheck: " + r.Err.Error()
	}
	if len(r.Findings) == 0 {
		return "tddcheck: passed"
	}

	lines := make([]string, 0, len(r.Findings)+1)
	lines = append(lines, "tddcheck: failed")
	for _, finding := range r.Findings {
		lines = append(lines, finding.String())
	}

	return strings.Join(lines, "\n")
}
