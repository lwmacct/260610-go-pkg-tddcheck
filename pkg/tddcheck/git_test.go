package tddcheck

import (
	"context"
	"os/exec"
	"testing"
)

func TestGitStagedFiles(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	root := t.TempDir()
	runGit(t, root, "init")
	runGit(t, root, "config", "user.email", "test@example.com")
	runGit(t, root, "config", "user.name", "Test")
	writeFile(t, root, "calc.go", "package calc\n")
	runGit(t, root, "add", "calc.go")

	files, err := GitStagedFiles(context.Background(), root)
	requireNoError(t, err)
	if len(files) != 1 || files[0] != "calc.go" {
		t.Fatalf("unexpected staged files: %#v", files)
	}
}

func runGit(t *testing.T, root string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(output))
	}
}
