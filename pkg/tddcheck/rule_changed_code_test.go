package tddcheck

import (
	"context"
	"testing"
)

func TestRuleChangedCodeNeedsTest(t *testing.T) {
	project := scanFixture(t, map[string]string{
		"calc.go":      "package calc\nfunc Parse() {}\n",
		"calc_test.go": "package calc\n",
	})
	project.ChangedFiles = map[string]bool{"calc.go": true}

	findings, err := RuleChangedCodeNeedsTest().Check(context.Background(), project)
	requireNoError(t, err)
	if len(findings) != 1 {
		t.Fatalf("expected one finding, got %#v", findings)
	}

	project.ChangedFiles["calc_test.go"] = true
	findings, err = RuleChangedCodeNeedsTest().Check(context.Background(), project)
	requireNoError(t, err)
	if len(findings) != 0 {
		t.Fatalf("expected no findings, got %#v", findings)
	}
}
