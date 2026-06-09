package tddcheck

import "fmt"

// Finding is one rule violation or warning.
type Finding struct {
	Rule     string
	Severity Severity
	File     string
	Line     int
	Message  string
}

func (f Finding) String() string {
	location := f.File
	if f.Line > 0 {
		location = fmt.Sprintf("%s:%d", f.File, f.Line)
	}
	if location == "" {
		location = "-"
	}

	return fmt.Sprintf("%s [%s] %s: %s", location, f.Severity, f.Rule, f.Message)
}
