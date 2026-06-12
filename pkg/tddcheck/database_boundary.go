package tddcheck

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type DatabaseTestRules struct {
	Root   string
	Config Config
}

type DatabaseTestViolation struct {
	File    string
	Line    int
	Message string
}

func (r DatabaseTestRules) AssertDatabaseTestBoundary(t *testing.T) {
	t.Helper()

	violations, err := r.DatabaseTestBoundaryViolations()
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) == 0 {
		return
	}

	lines := make([]string, 0, len(violations))
	for _, violation := range violations {
		lines = append(lines, fmt.Sprintf(
			"%s:%d: %s",
			violation.File,
			violation.Line,
			violation.Message,
		))
	}
	t.Fatalf("invalid database test boundaries:\n  - %s", strings.Join(lines, "\n  - "))
}

func (r DatabaseTestRules) DatabaseTestBoundaryViolations() ([]DatabaseTestViolation, error) {
	config := r.Config.withDefaults()
	root, err := r.root()
	if err != nil {
		return nil, err
	}

	rootFS := os.DirFS(root)
	var violations []DatabaseTestViolation
	err = fs.WalkDir(rootFS, ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", "node_modules", "dist", "build":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		repoPath := filepath.ToSlash(path)
		if databaseTestBoundaryAllowed(config, repoPath) {
			return nil
		}
		data, readErr := fs.ReadFile(rootFS, path)
		if readErr != nil {
			return readErr
		}
		content := string(data)
		if strings.Contains(content, config.DatabaseTest.OpenNeedle) && strings.Contains(content, config.DatabaseTest.TempDirNeedle) {
			violations = append(violations, DatabaseTestViolation{
				File:    repoPath,
				Line:    lineOf(content, config.DatabaseTest.OpenNeedle),
				Message: config.DatabaseTest.OpenMessage,
			})
		}
		if strings.Contains(content, config.DatabaseTest.ConfigPathNeedle) {
			violations = append(violations, DatabaseTestViolation{
				File:    repoPath,
				Line:    lineOf(content, config.DatabaseTest.ConfigPathNeedle),
				Message: config.DatabaseTest.ConfigPathMessage,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return violations, nil
}

func (r DatabaseTestRules) root() (string, error) {
	if r.Root == "" {
		return "", errors.New("DatabaseTestRules.Root is empty")
	}
	if filepath.IsAbs(r.Root) {
		return r.Root, nil
	}
	projectRoot, err := findProjectRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(projectRoot, r.Root), nil
}

func databaseTestBoundaryAllowed(config Config, path string) bool {
	return stringIn(path, config.DatabaseTest.AllowedPaths)
}

func lineOf(content string, needle string) int {
	before, _, ok := strings.Cut(content, needle)
	if !ok {
		return 1
	}
	return strings.Count(before, "\n") + 1
}
