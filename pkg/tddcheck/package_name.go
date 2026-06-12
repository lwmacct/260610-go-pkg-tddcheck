package tddcheck

import (
	"fmt"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// ModulePackageNameRules declares mechanical package naming rules.
type ModulePackageNameRules struct {
	// Root is the layered module root directory. Relative paths are resolved from go.mod.
	Root string
}

// PackageNameViolation describes one package clause that does not match its directory.
type PackageNameViolation struct {
	File    string
	Line    int
	Message string
}

// AssertPackageNames fails the test when package names do not match directory names.
func (r ModulePackageNameRules) AssertPackageNames(t *testing.T) {
	t.Helper()

	violations, err := r.PackageNameViolations()
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

	t.Fatalf("invalid package names:\n  - %s", strings.Join(lines, "\n  - "))
}

// PackageNameViolations returns all package name violations.
func (r ModulePackageNameRules) PackageNameViolations() ([]PackageNameViolation, error) {
	moduleDirs, err := modulePackageDirs(r.Root, "ModulePackageNameRules")
	if err != nil {
		return nil, err
	}

	var violations []PackageNameViolation
	for _, moduleDir := range moduleDirs {
		files, err := filepath.Glob(filepath.Join(moduleDir, "*.go"))
		if err != nil {
			return nil, err
		}
		slices.Sort(files)
		expected := utilPackageNameFromDir(moduleDir)
		for _, file := range files {
			fileViolations, err := packageNameViolationsInFile(file, expected)
			if err != nil {
				return nil, err
			}
			violations = append(violations, fileViolations...)
		}
	}

	return violations, nil
}

func packageNameViolationsInFile(filename string, expected string) ([]PackageNameViolation, error) {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, filename, nil, parser.PackageClauseOnly)
	if err != nil {
		return nil, err
	}

	actual := file.Name.Name
	if actual == expected || actual == expected+"_test" {
		return nil, nil
	}

	position := fileSet.Position(file.Name.Pos())
	return []PackageNameViolation{
		{
			File:    filename,
			Line:    position.Line,
			Message: fmt.Sprintf("package name must be %q or %q, got %q", expected, expected+"_test", actual),
		},
	}, nil
}

func utilPackageNameFromDir(dir string) string {
	name := filepath.Base(dir)
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, ".", "_")
	return name
}
