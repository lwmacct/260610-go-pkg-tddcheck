package tddcheck

import "context"

// TestsAreNotEmpty rejects empty tests. It also rejects skipped tests in
// production files, while allowing environment guards in _test.go files.
func TestsAreNotEmpty() Rule {
	return ruleFunc{
		name: "no-skipped-or-empty-tests",
		check: func(ctx context.Context, project *Project) ([]Finding, error) {
			_ = ctx
			var findings []Finding
			for _, file := range project.Files {
				for _, test := range file.Tests {
					if test.Empty {
						findings = append(findings, Finding{
							Rule:     "no-skipped-or-empty-tests",
							Severity: SeverityError,
							File:     test.File,
							Line:     test.Line,
							Message:  "test function is empty",
						})
					}
					if test.Skipped && !isGoTestFile(test.File) {
						findings = append(findings, Finding{
							Rule:     "no-skipped-or-empty-tests",
							Severity: SeverityError,
							File:     test.File,
							Line:     test.Line,
							Message:  "test function contains t.Skip",
						})
					}
				}
			}

			return findings, nil
		},
	}
}
