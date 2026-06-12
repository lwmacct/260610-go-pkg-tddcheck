package databasetest

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rulekit"
)

type Rules struct {
	root   string
	config rulekit.Config
}

// New creates rules for the supplied module root.
func New(root string, options ...rulekit.Option) Rules {
	values := rulekit.NewRuleOptions(root, options...)
	return Rules{root: values.Root, config: values.Config}
}

type DatabaseTestViolation struct {
	File    string
	Line    int
	Message string
}

func (r Rules) Assert(t *testing.T) {
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

func (r Rules) DatabaseTestBoundaryViolations() ([]DatabaseTestViolation, error) {
	config := r.config.WithDefaults()
	root, err := r.resolveRoot()
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

func (r Rules) resolveRoot() (string, error) {
	if r.root == "" {
		return "", errors.New("Rules.Root is empty")
	}
	if filepath.IsAbs(r.root) {
		return r.root, nil
	}
	projectRoot, err := rulekit.FindProjectRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(projectRoot, r.root), nil
}

func databaseTestBoundaryAllowed(config rulekit.Config, path string) bool {
	return rulekit.StringIn(path, config.DatabaseTest.AllowedPaths)
}

func lineOf(content string, needle string) int {
	before, _, ok := strings.Cut(content, needle)
	if !ok {
		return 1
	}
	return strings.Count(before, "\n") + 1
}
