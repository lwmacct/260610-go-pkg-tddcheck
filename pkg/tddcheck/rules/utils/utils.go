package utils

import (
	"fmt"
	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rulekit"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// Rules declares mechanical boundary rules for module utils.go files.
type Rules struct {
	root   string
	config rulekit.Config
}

// New creates rules for the supplied module root.
func New(root string, options ...rulekit.Option) Rules {
	values := rulekit.NewRuleOptions(root, options...)
	return Rules{root: values.Root, config: values.Config}
}

// UtilsBoundaryViolation describes one utils.go boundary violation.
type UtilsBoundaryViolation struct {
	File    string
	Line    int
	Message string
}

// AssertUtilsBoundaries fails the test when module utils boundaries are violated.
func (r Rules) AssertUtilsBoundaries(t *testing.T) {
	t.Helper()

	violations, err := r.UtilsBoundaryViolations()
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

	t.Fatalf("invalid utils boundaries:\n  - %s", strings.Join(lines, "\n  - "))
}

// UtilsBoundaryViolations returns all module utils boundary violations.
func (r Rules) UtilsBoundaryViolations() ([]UtilsBoundaryViolation, error) {
	moduleDirs, err := rulekit.ModulePackageDirs(r.root, "Rules", r.config)
	if err != nil {
		return nil, err
	}

	var violations []UtilsBoundaryViolation
	for _, moduleDir := range moduleDirs {
		files, err := filepath.Glob(filepath.Join(moduleDir, "*.go"))
		if err != nil {
			return nil, err
		}
		slices.Sort(files)
		for _, file := range files {
			if strings.HasSuffix(file, "_test.go") {
				continue
			}
			fileViolations, err := utilsBoundaryViolationsInFile(file)
			if err != nil {
				return nil, err
			}
			violations = append(violations, fileViolations...)
		}
	}

	return violations, nil
}

func utilsBoundaryViolationsInFile(filename string) ([]UtilsBoundaryViolation, error) {
	fileSet := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fileSet, filename, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}

	if filepath.Base(filename) == "utils.go" {
		return utilsFileBoundaryViolations(fileSet, filename, parsedFile), nil
	}
	return nonUtilsFileBoundaryViolations(fileSet, filename, parsedFile), nil
}

func utilsFileBoundaryViolations(fileSet *token.FileSet, filename string, parsedFile *ast.File) []UtilsBoundaryViolation {
	var violations []UtilsBoundaryViolation
	for _, decl := range parsedFile.Decls {
		switch typed := decl.(type) {
		case *ast.GenDecl:
			if typed.Tok == token.IMPORT {
				continue
			}
			position := fileSet.Position(typed.Pos())
			violations = append(violations, UtilsBoundaryViolation{
				File:    rulekit.DisplayFilename(filename),
				Line:    position.Line,
				Message: "utils.go must only declare private package-level util* functions",
			})
		case *ast.FuncDecl:
			position := fileSet.Position(typed.Pos())
			if typed.Recv != nil {
				violations = append(violations, UtilsBoundaryViolation{
					File:    rulekit.DisplayFilename(filename),
					Line:    position.Line,
					Message: "utils.go must not declare receiver method " + typed.Name.Name,
				})
				continue
			}
			if !isUtilFunctionName(typed.Name.Name) {
				violations = append(violations, UtilsBoundaryViolation{
					File:    rulekit.DisplayFilename(filename),
					Line:    position.Line,
					Message: fmt.Sprintf("utils.go function %s must use util* prefix", typed.Name.Name),
				})
			}
		}
	}
	return violations
}

func nonUtilsFileBoundaryViolations(fileSet *token.FileSet, filename string, parsedFile *ast.File) []UtilsBoundaryViolation {
	var violations []UtilsBoundaryViolation
	for _, decl := range parsedFile.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || !isUtilFunctionName(funcDecl.Name.Name) {
			continue
		}
		position := fileSet.Position(funcDecl.Pos())
		violations = append(violations, UtilsBoundaryViolation{
			File:    rulekit.DisplayFilename(filename),
			Line:    position.Line,
			Message: fmt.Sprintf("util* function %s must be declared in utils.go", funcDecl.Name.Name),
		})
	}
	return violations
}

func isUtilFunctionName(name string) bool {
	return strings.HasPrefix(name, "util") && len(name) > len("util")
}
