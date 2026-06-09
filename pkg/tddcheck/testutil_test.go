package tddcheck

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	requireNoError(t, os.MkdirAll(filepath.Dir(path), 0750))
	requireNoError(t, os.WriteFile(path, []byte(content), 0600))
}

func scanFixture(t *testing.T, files map[string]string) *Project {
	t.Helper()
	root := t.TempDir()
	for rel, content := range files {
		writeFile(t, root, rel, content)
	}
	project, err := Scan(context.Background(), ScanOptions{Root: root})
	requireNoError(t, err)

	return project
}

func requireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
