package errorboundary

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

// ModuleErrorRules declares mechanical boundary rules for module errors.go files.
type ModuleErrorRules struct {
	// Root is the layered module root directory. Relative paths are resolved from go.mod.
	Root   string
	Config rulekit.Config
}

// ErrorsBoundaryViolation describes one errors.go boundary violation.
type ErrorsBoundaryViolation struct {
	File    string
	Line    int
	Message string
}

// AssertErrorsBoundaries fails the test when module errors boundaries are violated.
func (r ModuleErrorRules) AssertErrorsBoundaries(t *testing.T) {
	t.Helper()

	violations, err := r.ErrorsBoundaryViolations()
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

	t.Fatalf("invalid errors boundaries:\n  - %s", strings.Join(lines, "\n  - "))
}

// ErrorsBoundaryViolations returns all module errors boundary violations.
func (r ModuleErrorRules) ErrorsBoundaryViolations() ([]ErrorsBoundaryViolation, error) {
	moduleDirs, err := rulekit.ModulePackageDirs(r.Root, "ModuleErrorRules", r.Config)
	if err != nil {
		return nil, err
	}

	var violations []ErrorsBoundaryViolation
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
			fileViolations, err := errorsBoundaryViolationsInFile(file)
			if err != nil {
				return nil, err
			}
			violations = append(violations, fileViolations...)
		}
	}

	return violations, nil
}

func errorsBoundaryViolationsInFile(filename string) ([]ErrorsBoundaryViolation, error) {
	fileSet := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fileSet, filename, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}

	if filepath.Base(filename) == "errors.go" {
		return errorsFileBoundaryViolations(fileSet, filename, parsedFile), nil
	}
	return nonErrorsFileBoundaryViolations(fileSet, filename, parsedFile), nil
}

func errorsFileBoundaryViolations(fileSet *token.FileSet, filename string, parsedFile *ast.File) []ErrorsBoundaryViolation {
	errorTypes := errorsFileTypeNames(parsedFile)
	var violations []ErrorsBoundaryViolation
	for _, decl := range parsedFile.Decls {
		switch typed := decl.(type) {
		case *ast.GenDecl:
			if typed.Tok == token.IMPORT {
				continue
			}
			if typed.Tok == token.VAR {
				violations = append(violations, errorsFileVarViolations(fileSet, filename, typed)...)
				continue
			}
			if typed.Tok == token.TYPE {
				violations = append(violations, errorsFileTypeViolations(fileSet, filename, typed)...)
				continue
			}
			position := fileSet.Position(typed.Pos())
			violations = append(violations, ErrorsBoundaryViolation{
				File:    rulekit.DisplayFilename(filename),
				Line:    position.Line,
				Message: "errors.go must only declare error vars, *Error types, and error helpers",
			})
		case *ast.FuncDecl:
			position := fileSet.Position(typed.Pos())
			if typed.Recv != nil {
				receiver := rulekit.ReceiverTypeName(typed.Recv)
				if _, ok := errorTypes[receiver]; ok && isErrorMethodName(typed.Name.Name) {
					continue
				}
				violations = append(violations, ErrorsBoundaryViolation{
					File:    rulekit.DisplayFilename(filename),
					Line:    position.Line,
					Message: fmt.Sprintf("errors.go method %s must be Error, Is, As, or Unwrap on a local *Error type", typed.Name.Name),
				})
				continue
			}
			if !isErrorHelperName(typed.Name.Name) {
				violations = append(violations, ErrorsBoundaryViolation{
					File:    rulekit.DisplayFilename(filename),
					Line:    position.Line,
					Message: fmt.Sprintf("errors.go function %s must use Is*, As*, or Wrap* name", typed.Name.Name),
				})
			}
		}
	}
	return violations
}

func nonErrorsFileBoundaryViolations(fileSet *token.FileSet, filename string, parsedFile *ast.File) []ErrorsBoundaryViolation {
	var violations []ErrorsBoundaryViolation
	for _, decl := range parsedFile.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			continue
		}
		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok || !isErrorValueSpec(valueSpec) {
				continue
			}
			position := fileSet.Position(valueSpec.Pos())
			violations = append(violations, ErrorsBoundaryViolation{
				File:    rulekit.DisplayFilename(filename),
				Line:    position.Line,
				Message: "package-level error var must be declared in errors.go",
			})
		}
	}
	return violations
}

func errorsFileVarViolations(fileSet *token.FileSet, filename string, decl *ast.GenDecl) []ErrorsBoundaryViolation {
	var violations []ErrorsBoundaryViolation
	for _, spec := range decl.Specs {
		valueSpec, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		for _, name := range valueSpec.Names {
			if strings.HasPrefix(name.Name, "Err") && isErrorValueSpec(valueSpec) {
				continue
			}
			position := fileSet.Position(name.Pos())
			violations = append(violations, ErrorsBoundaryViolation{
				File:    rulekit.DisplayFilename(filename),
				Line:    position.Line,
				Message: fmt.Sprintf("errors.go var %s must be an Err* error value", name.Name),
			})
		}
	}
	return violations
}

func errorsFileTypeViolations(fileSet *token.FileSet, filename string, decl *ast.GenDecl) []ErrorsBoundaryViolation {
	var violations []ErrorsBoundaryViolation
	for _, spec := range decl.Specs {
		typeSpec, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		if strings.HasSuffix(typeSpec.Name.Name, "Error") && !typeSpec.Assign.IsValid() {
			continue
		}
		position := fileSet.Position(typeSpec.Pos())
		violations = append(violations, ErrorsBoundaryViolation{
			File:    rulekit.DisplayFilename(filename),
			Line:    position.Line,
			Message: fmt.Sprintf("errors.go type %s must be a non-alias *Error type", typeSpec.Name.Name),
		})
	}
	return violations
}

func errorsFileTypeNames(parsedFile *ast.File) map[string]struct{} {
	names := map[string]struct{}{}
	for _, decl := range parsedFile.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if ok && strings.HasSuffix(typeSpec.Name.Name, "Error") && !typeSpec.Assign.IsValid() {
				names[typeSpec.Name.Name] = struct{}{}
			}
		}
	}
	return names
}

func isErrorMethodName(name string) bool {
	return name == "Error" || name == "Is" || name == "As" || name == "Unwrap"
}

func isErrorHelperName(name string) bool {
	return hasErrorHelperPrefix(name, "Is") ||
		hasErrorHelperPrefix(name, "As") ||
		hasErrorHelperPrefix(name, "Wrap")
}

func hasErrorHelperPrefix(name string, prefix string) bool {
	return strings.HasPrefix(name, prefix) && len(name) > len(prefix)
}

func isErrorValueSpec(spec *ast.ValueSpec) bool {
	if isErrorType(spec.Type) {
		return true
	}
	return slices.ContainsFunc(spec.Values, isErrorConstructor)
}

func isErrorType(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == "error"
}

func isErrorConstructor(expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}

	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	pkg, ok := selector.X.(*ast.Ident)
	if !ok {
		return false
	}

	return (pkg.Name == "errors" && selector.Sel.Name == "New") ||
		(pkg.Name == "fmt" && selector.Sel.Name == "Errorf")
}
