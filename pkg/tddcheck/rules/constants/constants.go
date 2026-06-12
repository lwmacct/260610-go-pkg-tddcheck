package constants

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

// Rules declares mechanical boundary rules for module constants.go files.
type Rules struct {
	root   string
	config rulekit.Config
}

// New creates rules for the supplied module root.
func New(root string, options ...rulekit.Option) Rules {
	values := rulekit.NewRuleOptions(root, options...)
	return Rules{root: values.Root, config: values.Config}
}

// ConstantsBoundaryViolation describes one constants boundary violation.
type ConstantsBoundaryViolation struct {
	File    string
	Line    int
	Message string
}

// AssertConstantsBoundaries fails the test when module constants boundaries are violated.
func (r Rules) AssertConstantsBoundaries(t *testing.T) {
	t.Helper()

	violations, err := r.ConstantsBoundaryViolations()
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

	t.Fatalf("invalid constants boundaries:\n  - %s", strings.Join(lines, "\n  - "))
}

// ConstantsBoundaryViolations returns all module constants boundary violations.
func (r Rules) ConstantsBoundaryViolations() ([]ConstantsBoundaryViolation, error) {
	moduleDirs, err := rulekit.ModulePackageDirs(r.root, "Rules", r.config)
	if err != nil {
		return nil, err
	}

	var violations []ConstantsBoundaryViolation
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
			fileViolations, err := constantsBoundaryViolationsInFile(file)
			if err != nil {
				return nil, err
			}
			violations = append(violations, fileViolations...)
		}
	}

	return violations, nil
}

func constantsBoundaryViolationsInFile(filename string) ([]ConstantsBoundaryViolation, error) {
	fileSet := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fileSet, filename, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}

	if filepath.Base(filename) == "constants.go" {
		return constantsFileBoundaryViolations(fileSet, filename, parsedFile), nil
	}
	return nonConstantsFileBoundaryViolations(fileSet, filename, parsedFile), nil
}

func constantsFileBoundaryViolations(fileSet *token.FileSet, filename string, parsedFile *ast.File) []ConstantsBoundaryViolation {
	var violations []ConstantsBoundaryViolation
	for _, decl := range parsedFile.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if ok && (genDecl.Tok == token.IMPORT || genDecl.Tok == token.CONST) {
			continue
		}
		position := fileSet.Position(decl.Pos())
		violations = append(violations, ConstantsBoundaryViolation{
			File:    rulekit.DisplayFilename(filename),
			Line:    position.Line,
			Message: "constants.go must only declare const",
		})
	}
	return violations
}

func nonConstantsFileBoundaryViolations(fileSet *token.FileSet, filename string, parsedFile *ast.File) []ConstantsBoundaryViolation {
	var violations []ConstantsBoundaryViolation
	for _, decl := range parsedFile.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			continue
		}
		position := fileSet.Position(genDecl.Pos())
		violations = append(violations, ConstantsBoundaryViolation{
			File:    rulekit.DisplayFilename(filename),
			Line:    position.Line,
			Message: "package-level const must be declared in constants.go",
		})
	}
	return violations
}
