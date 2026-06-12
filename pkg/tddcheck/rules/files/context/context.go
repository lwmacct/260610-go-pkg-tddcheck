package context

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rulekit"
)

// Rules declares mechanical boundary rules for module context.go files.
type Rules struct {
	root   string
	config rulekit.Config
}

// New 为给定模块根目录创建规则。
func New(root string, options ...rulekit.Option) Rules {
	values := rulekit.NewRuleOptions(root, options...)
	return Rules{root: values.Root, config: values.Config}
}

// ContextBoundaryViolation describes one context boundary violation.
type ContextBoundaryViolation struct {
	File    string
	Line    int
	Message string
}

// Assert 在模块 context 边界被违反时让测试失败。
func (r Rules) Assert(t *testing.T) {
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

// ContextBoundaryViolations 返回所有模块 context 边界违规。
func (r Rules) ContextBoundaryViolations() ([]ContextBoundaryViolation, error) {
	moduleDirs, err := rulekit.ModulePackageDirs(r.root, "Rules", r.config)
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
	contextImports := contextImportNames(parsedFile)
	if isContextFile {
		return contextFileBoundaryViolations(fileSet, filename, parsedFile, contextImports), nil
	}
	return nonContextFileBoundaryViolations(fileSet, filename, parsedFile, contextImports), nil
}

func contextFileBoundaryViolations(fileSet *token.FileSet, filename string, parsedFile *ast.File, contextImports map[string]struct{}) []ContextBoundaryViolation {
	localTypes := contextFileTypeNames(parsedFile)
	var violations []ContextBoundaryViolation
	for _, decl := range parsedFile.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if isContextHelperFunc(funcDecl, contextImports) {
			continue
		}
		if funcDecl.Recv != nil && contextReceiverDeclaredInFile(funcDecl.Recv, localTypes) {
			continue
		}
		position := fileSet.Position(funcDecl.Pos())
		violations = append(violations, ContextBoundaryViolation{
			File:    rulekit.DisplayFilename(filename),
			Line:    position.Line,
			Message: fmt.Sprintf("context.go function %s must be a context helper or local context type method", funcDecl.Name.Name),
		})
	}
	return violations
}

func nonContextFileBoundaryViolations(fileSet *token.FileSet, filename string, parsedFile *ast.File, contextImports map[string]struct{}) []ContextBoundaryViolation {
	var violations []ContextBoundaryViolation
	for _, decl := range parsedFile.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || !isContextHelperFuncOutsideContextFile(funcDecl, contextImports) {
			continue
		}
		position := fileSet.Position(funcDecl.Pos())
		violations = append(violations, ContextBoundaryViolation{
			File:    rulekit.DisplayFilename(filename),
			Line:    position.Line,
			Message: fmt.Sprintf("context helper %s must be declared in context.go", funcDecl.Name.Name),
		})
	}

	ast.Inspect(parsedFile, func(node ast.Node) bool {
		selector, ok := node.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "WithValue" {
			return true
		}
		if ident, ok := selector.X.(*ast.Ident); !ok || !isContextImportName(ident.Name, contextImports) {
			return true
		}
		position := fileSet.Position(selector.Sel.Pos())
		violations = append(violations, ContextBoundaryViolation{
			File:    rulekit.DisplayFilename(filename),
			Line:    position.Line,
			Message: "context.WithValue must be used in context.go",
		})
		return true
	})

	return violations
}

func contextImportNames(parsedFile *ast.File) map[string]struct{} {
	names := map[string]struct{}{"context": {}}
	for _, importSpec := range parsedFile.Imports {
		if strings.Trim(importSpec.Path.Value, `"`) != "context" {
			continue
		}
		if importSpec.Name == nil {
			names["context"] = struct{}{}
			continue
		}
		if importSpec.Name.Name == "." || importSpec.Name.Name == "_" {
			continue
		}
		names[importSpec.Name.Name] = struct{}{}
	}
	return names
}

func isContextImportName(name string, contextImports map[string]struct{}) bool {
	_, ok := contextImports[name]
	return ok
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

func isContextHelperFunc(funcDecl *ast.FuncDecl, contextImports map[string]struct{}) bool {
	if funcDecl.Recv != nil {
		return false
	}
	name := funcDecl.Name.Name
	return strings.HasPrefix(name, "ContextWith") ||
		strings.HasPrefix(name, "contextWith") ||
		(strings.HasSuffix(name, "FromContext") && hasSingleContextParam(funcDecl, contextImports)) ||
		(strings.HasSuffix(name, "ContextFrom") && hasContextParam(funcDecl, contextImports)) ||
		(strings.HasSuffix(name, "Context") && hasContextParam(funcDecl, contextImports))
}

func isContextHelperFuncOutsideContextFile(funcDecl *ast.FuncDecl, contextImports map[string]struct{}) bool {
	if funcDecl.Recv != nil {
		return false
	}
	name := funcDecl.Name.Name
	return strings.HasPrefix(name, "ContextWith") ||
		strings.HasPrefix(name, "contextWith") ||
		(strings.HasSuffix(name, "FromContext") && hasSingleContextParam(funcDecl, contextImports)) ||
		(strings.HasSuffix(name, "ContextFrom") && hasContextParam(funcDecl, contextImports)) ||
		(strings.HasSuffix(name, "Context") && hasSingleContextParam(funcDecl, contextImports)) ||
		(strings.HasSuffix(name, "Context") && ast.IsExported(name) && hasContextParam(funcDecl, contextImports))
}

func hasSingleContextParam(funcDecl *ast.FuncDecl, contextImports map[string]struct{}) bool {
	return funcDecl.Type.Params != nil &&
		len(funcDecl.Type.Params.List) == 1 &&
		isContextExpr(funcDecl.Type.Params.List[0].Type, contextImports)
}

func hasContextParam(funcDecl *ast.FuncDecl, contextImports map[string]struct{}) bool {
	if funcDecl.Type.Params == nil || len(funcDecl.Type.Params.List) == 0 {
		return false
	}
	return isContextExpr(funcDecl.Type.Params.List[0].Type, contextImports)
}

func isContextExpr(expr ast.Expr, contextImports map[string]struct{}) bool {
	selector, ok := expr.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Context" {
		return false
	}
	ident, ok := selector.X.(*ast.Ident)
	return ok && isContextImportName(ident.Name, contextImports)
}
