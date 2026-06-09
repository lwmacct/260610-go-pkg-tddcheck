package tddcheck

import (
	"context"
	"fmt"
)

// RuleExportedDeclsNeedTests requires exported production functions and methods
// to have candidate tests in the same package.
func RuleExportedDeclsNeedTests() Rule {
	return ruleFunc{
		name: "exported-decls-need-tests",
		check: func(ctx context.Context, project *Project) ([]Finding, error) {
			_ = ctx
			var findings []Finding
			for _, pkg := range project.Packages {
				for _, decl := range pkg.Decls {
					if shouldSkipExportedDecl(decl) {
						continue
					}
					if hasCandidateTest(pkg, decl) {
						continue
					}
					findings = append(findings, Finding{
						Rule:     "exported-decls-need-tests",
						Severity: SeverityError,
						File:     decl.File,
						Line:     decl.Line,
						Message:  fmt.Sprintf("exported %s %q has no candidate test", decl.Kind, declDisplayName(decl)),
					})
				}
			}

			return findings, nil
		},
	}
}

func shouldSkipExportedDecl(decl Decl) bool {
	if !decl.Exported || isGoTestFile(decl.File) || isTestDecl(decl) {
		return true
	}
	if decl.Kind == DeclType {
		return true
	}
	if decl.Kind == DeclMethod && (decl.Recv == "" || !isExportedName(decl.Recv)) {
		return true
	}

	return false
}

func isTestDecl(decl Decl) bool {
	return decl.Kind == DeclFunc && testNameExists(decl.Name)
}

func testNameExists(name string) bool {
	_, ok := testKind(name)

	return ok
}

func hasCandidateTest(pkg *Package, decl Decl) bool {
	for _, name := range candidateTestNames(decl) {
		for testName := range pkg.TestNames {
			if testMatchesCandidate(testName, name) {
				return true
			}
		}
	}

	return false
}

func testMatchesCandidate(testName, candidate string) bool {
	if testName == candidate {
		return true
	}
	if len(testName) <= len(candidate) || testName[:len(candidate)] != candidate {
		return false
	}

	next := testName[len(candidate)]
	return next == '_' || next >= 'A' && next <= 'Z'
}

func candidateTestNames(decl Decl) []string {
	switch decl.Kind {
	case DeclMethod:
		if decl.Recv == "" {
			return []string{"Test" + decl.Name}
		}
		return []string{"Test" + decl.Recv + "_" + decl.Name, "Test" + decl.Name}
	default:
		return []string{"Test" + decl.Name}
	}
}

func declDisplayName(decl Decl) string {
	if decl.Recv == "" {
		return decl.Name
	}

	return decl.Recv + "." + decl.Name
}

func isExportedName(name string) bool {
	if name == "" {
		return false
	}

	return name[0] >= 'A' && name[0] <= 'Z'
}
