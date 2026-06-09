package tddcheck

import (
	"context"
	"testing"
)

func TestScan(t *testing.T) {
	project := scanFixture(t, map[string]string{
		"calc.go": `package calc

func Parse() {}
`,
		"calc_test.go": `package calc

import "testing"

func TestParse(t *testing.T) {}
`,
	})

	if len(project.Files) != 2 {
		t.Fatalf("expected two files, got %d", len(project.Files))
	}
	if len(project.Packages) != 1 {
		t.Fatalf("expected one package, got %d", len(project.Packages))
	}
}

func TestScanOptions(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "ignored/calc.go", "package calc\nfunc Parse() {}\n")

	project, err := Scan(context.Background(), ScanOptions{
		Root:        root,
		IgnoreGlobs: []string{"ignored/**"},
	})
	requireNoError(t, err)
	if len(project.Files) != 0 {
		t.Fatalf("expected ignored files, got %#v", project.Files)
	}
}
