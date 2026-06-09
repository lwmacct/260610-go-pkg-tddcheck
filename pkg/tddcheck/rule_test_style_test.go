package tddcheck

import (
	"context"
	"testing"
)

func TestTestsAreNotEmptyReportsEmptyTestFunctions(t *testing.T) {
	project := scanFixture(t, map[string]string{
		"calc.go": `package calc

func Parse() {}
`,
		"calc_test.go": `package calc

import "testing"

func TestParse(t *testing.T) {}

func TestSkipped(t *testing.T) {
	t.Skip("later")
}
`,
	})

	findings, err := TestsAreNotEmpty().Check(context.Background(), project)
	requireNoError(t, err)
	if len(findings) != 1 {
		t.Fatalf("expected one finding, got %#v", findings)
	}
}
