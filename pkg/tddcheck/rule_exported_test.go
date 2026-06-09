package tddcheck

import (
	"context"
	"testing"
)

func TestRuleExportedDeclsNeedTestsPassesWithCandidates(t *testing.T) {
	project := scanFixture(t, map[string]string{
		"calc.go": `package calc

type Client struct{}

func Parse() {}

func (c *Client) Do() {}
`,
		"calc_test.go": `package calc

import "testing"

func TestClient(t *testing.T) {}
func TestParse(t *testing.T) {}
func TestClient_Do(t *testing.T) {}
`,
	})

	findings, err := RuleExportedDeclsNeedTests().Check(context.Background(), project)
	requireNoError(t, err)
	if len(findings) != 0 {
		t.Fatalf("expected no findings, got %#v", findings)
	}
}

func TestRuleExportedDeclsNeedTestsReportsMissingCandidates(t *testing.T) {
	project := scanFixture(t, map[string]string{
		"calc.go": `package calc

func Parse() {}
`,
	})

	findings, err := RuleExportedDeclsNeedTests().Check(context.Background(), project)
	requireNoError(t, err)
	if len(findings) != 1 {
		t.Fatalf("expected one finding, got %#v", findings)
	}
	if findings[0].File != "calc.go" || findings[0].Line != 3 {
		t.Fatalf("unexpected finding location: %#v", findings[0])
	}
}
