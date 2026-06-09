package main

import (
	"os"
	"testing"
)

func TestRunVersion(t *testing.T) {
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	os.Args = []string{"tddcheck", "--version"}
	if code := run(); code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}
