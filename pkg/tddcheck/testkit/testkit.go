package testkit

import (
	"os"
	"path/filepath"
	"testing"
)

func WriteFile(t *testing.T, root string, name string, content string) {
	t.Helper()

	filename := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(filename), 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filename, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}
