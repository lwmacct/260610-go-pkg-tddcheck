package tddcheck

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// ModuleContextRules declares mechanical boundary rules for module context.go files.
type ModuleContextRules struct {
	// Root is the layered module root directory. Relative paths are resolved from go.mod.
	Root string
}

// ContextBoundaryViolation describes one context boundary violation.
type ContextBoundaryViolation struct {
	File    string
	Line    int
	Message string
}

// AssertContextBoundaries fails the test when module context boundaries are violated.
func (r ModuleContextRules) AssertContextBoundaries(t *testing.T) {
	t.Helper()

	violations, err := r.ContextBoundaryViolations()
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

	t.Fatalf("invalid context boundaries:\n  - %s", strings.Join(lines, "\n  - "))
}

// ContextBoundaryViolations returns all module context boundary violations.
func (r ModuleContextRules) ContextBoundaryViolations() ([]ContextBoundaryViolation, error) {
	moduleDirs, err := modulePackageDirs(r.Root, "ModuleContextRules")
	if err != nil {
		return nil, err
	}

	var violations []ContextBoundaryViolation
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
			fileViolations, err := contextBoundaryViolationsInFile(file)
			if err != nil {
				return nil, err
			}
			violations = append(violations, fileViolations...)
		}
	}

	return violations, nil
}

func contextBoundaryViolationsInFile(filename string) ([]ContextBoundaryViolation, error) {
	fileSet := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fileSet, filename, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}

	isContextFile := filepath.Base(filename) == "context.go"
	if isContextFile {
		return contextFileBoundaryViolations(fileSet, filename, parsedFile), nil
	}
	return nonContextFileBoundaryViolations(fileSet, filename, parsedFile), nil
}

func contextFileBoundaryViolations(fileSet *token.FileSet, filename string, parsedFile *ast.File) []ContextBoundaryViolation {
	localTypes := contextFileTypeNames(parsedFile)
	var violations []ContextBoundaryViolation
	for _, decl := range parsedFile.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if isContextHelperFunc(funcDecl) {
			continue
		}
		if funcDecl.Recv != nil && contextReceiverDeclaredInFile(funcDecl.Recv, localTypes) {
			continue
		}
		position := fileSet.Position(funcDecl.Pos())
		violations = append(violations, ContextBoundaryViolation{
			File:    displayFilename(filename),
			Line:    position.Line,
			Message: fmt.Sprintf("context.go function %s must be a context helper or local context type method", funcDecl.Name.Name),
		})
	}
	return violations
}

func nonContextFileBoundaryViolations(fileSet *token.FileSet, filename string, parsedFile *ast.File) []ContextBoundaryViolation {
	var violations []ContextBoundaryViolation
	for _, decl := range parsedFile.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || !isContextHelperFuncOutsideContextFile(funcDecl) {
			continue
		}
		position := fileSet.Position(funcDecl.Pos())
		violations = append(violations, ContextBoundaryViolation{
			File:    displayFilename(filename),
			Line:    position.Line,
			Message: fmt.Sprintf("context helper %s must be declared in context.go", funcDecl.Name.Name),
		})
	}

	ast.Inspect(parsedFile, func(node ast.Node) bool {
		selector, ok := node.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "WithValue" {
			return true
		}
		if ident, ok := selector.X.(*ast.Ident); !ok || ident.Name != "context" {
			return true
		}
		position := fileSet.Position(selector.Sel.Pos())
		violations = append(violations, ContextBoundaryViolation{
			File:    displayFilename(filename),
			Line:    position.Line,
			Message: "context.WithValue must be used in context.go",
		})
		return true
	})

	return violations
}

func contextFileTypeNames(parsedFile *ast.File) map[string]struct{} {
	names := map[string]struct{}{}
	for _, decl := range parsedFile.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if ok {
				names[typeSpec.Name.Name] = struct{}{}
			}
		}
	}
	return names
}

func contextReceiverDeclaredInFile(receiver *ast.FieldList, localTypes map[string]struct{}) bool {
	if receiver == nil || len(receiver.List) == 0 {
		return false
	}
	_, ok := localTypes[contextReceiverTypeName(receiver.List[0].Type)]
	return ok
}

func contextReceiverTypeName(expr ast.Expr) string {
	switch typed := expr.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.StarExpr:
		return contextReceiverTypeName(typed.X)
	}
	return ""
}

func isContextHelperFunc(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Recv != nil {
		return false
	}
	name := funcDecl.Name.Name
	return strings.HasPrefix(name, "ContextWith") ||
		strings.HasPrefix(name, "contextWith") ||
		(strings.HasSuffix(name, "FromContext") && hasSingleContextParam(funcDecl)) ||
		(strings.HasSuffix(name, "ContextFrom") && hasContextParam(funcDecl)) ||
		(strings.HasSuffix(name, "Context") && hasContextParam(funcDecl))
}

func isContextHelperFuncOutsideContextFile(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Recv != nil {
		return false
	}
	name := funcDecl.Name.Name
	return strings.HasPrefix(name, "ContextWith") ||
		strings.HasPrefix(name, "contextWith") ||
		(strings.HasSuffix(name, "FromContext") && hasSingleContextParam(funcDecl)) ||
		(strings.HasSuffix(name, "ContextFrom") && hasContextParam(funcDecl)) ||
		(strings.HasSuffix(name, "Context") && hasSingleContextParam(funcDecl)) ||
		(strings.HasSuffix(name, "Context") && ast.IsExported(name) && hasContextParam(funcDecl))
}

func hasSingleContextParam(funcDecl *ast.FuncDecl) bool {
	return funcDecl.Type.Params != nil &&
		len(funcDecl.Type.Params.List) == 1 &&
		isContextExpr(funcDecl.Type.Params.List[0].Type)
}

func hasContextParam(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Type.Params == nil || len(funcDecl.Type.Params.List) == 0 {
		return false
	}
	return isContextExpr(funcDecl.Type.Params.List[0].Type)
}

func isContextExpr(expr ast.Expr) bool {
	selector, ok := expr.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Context" {
		return false
	}
	ident, ok := selector.X.(*ast.Ident)
	return ok && ident.Name == "context"
}
