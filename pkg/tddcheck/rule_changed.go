package tddcheck

import (
	"context"
	"path/filepath"
	"strings"
)

// ChangedCodeHasTests requires staged production Go changes to include a staged
// test file in the same directory.
func ChangedCodeHasTests() Rule {
	return ruleFunc{
		name: "changed-code-needs-test",
		check: func(ctx context.Context, project *Project) ([]Finding, error) {
			_ = ctx
			if len(project.ChangedFiles) == 0 {
				return nil, nil
			}

			changedTestsByDir := make(map[string]bool)
			for changed := range project.ChangedFiles {
				if !strings.HasSuffix(changed, ".go") {
					continue
				}
				if isGoTestFile(changed) {
					changedTestsByDir[dirOf(changed)] = true
				}
			}

			var findings []Finding
			for _, file := range project.Files {
				if file.IsGenerated || file.IsTest || !project.ChangedFiles[file.Path] {
					continue
				}
				if changedTestsByDir[dirOf(file.Path)] {
					continue
				}
				findings = append(findings, Finding{
					Rule:     "changed-code-needs-test",
					Severity: SeverityError,
					File:     file.Path,
					Message:  "staged production Go change has no staged test file in the same package directory",
				})
			}

			return findings, nil
		},
	}
}

func dirOf(rel string) string {
	dir := filepath.ToSlash(filepath.Dir(rel))
	if dir == "." {
		return ""
	}

	return dir
}
