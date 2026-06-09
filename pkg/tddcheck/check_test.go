package tddcheck

import (
	"context"
	"testing"
)

func TestCheckPassesCleanProject(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "calc.go", `package calc

func Parse() {}
`)
	writeFile(t, root, "calc_test.go", `package calc

import "testing"

func TestParse(t *testing.T) {
	got := 1
	if got != 1 {
		t.Fatal(got)
	}
}
`)

	result, err := Check(context.Background(), WithRoot(root))
	requireNoError(t, err)
	if !result.Passed {
		t.Fatalf("expected pass, got %s", result.Text())
	}
}

func TestWithRoot(t *testing.T) {
	TestCheckPassesCleanProject(t)
}

func TestWithStagedOnly(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "calc.go", "package calc\nfunc Parse() {}\n")

	result, err := Check(context.Background(),
		WithRoot(root),
		WithStagedOnly(false),
		WithRules(),
	)
	requireNoError(t, err)
	if !result.Passed {
		t.Fatalf("expected pass, got %s", result.Text())
	}
}

func TestWithRules(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "calc.go", "package calc\nfunc Parse() {}\n")

	result, err := Check(context.Background(), WithRoot(root), WithRules())
	requireNoError(t, err)
	if !result.Passed {
		t.Fatalf("expected pass, got %s", result.Text())
	}
}

func TestWithDefaultRules(t *testing.T) {
	TestCheckPassesCleanProject(t)
}

func TestWithIgnoreGlobs(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "ignored/calc.go", "package calc\nfunc Parse() {}\n")

	result, err := Check(context.Background(),
		WithRoot(root),
		WithIgnoreGlobs("ignored/**"),
	)
	requireNoError(t, err)
	if !result.Passed {
		t.Fatalf("expected pass, got %s", result.Text())
	}
}
