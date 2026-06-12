package validation

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

// Rules declares mechanical boundary rules for module validation.go files.
type Rules struct {
	root   string
	config rulekit.Config
}

// New 为给定模块根目录创建规则。
func New(root string, options ...rulekit.Option) Rules {
	values := rulekit.NewRuleOptions(root, options...)
	return Rules{root: values.Root, config: values.Config}
}

// ValidationBoundaryViolation describes one validation boundary violation.
type ValidationBoundaryViolation struct {
	File    string
	Line    int
	Message string
}

// Assert 在模块 validation 边界被违反时让测试失败。
func (r Rules) Assert(t *testing.T) {
	t.Helper()

	violations, err := r.ValidationBoundaryViolations()
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

	t.Fatalf("invalid validation boundaries:\n  - %s", strings.Join(lines, "\n  - "))
}

// ValidationBoundaryViolations 返回所有模块 validation 边界违规。
func (r Rules) ValidationBoundaryViolations() ([]ValidationBoundaryViolation, error) {
	moduleDirs, err := rulekit.ModulePackageDirs(r.root, "Rules", r.config)
	if err != nil {
		return nil, err
	}

	var violations []ValidationBoundaryViolation
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
			fileViolations, err := validationBoundaryViolationsInFile(r.config, file)
			if err != nil {
				return nil, err
			}
			violations = append(violations, fileViolations...)
		}
	}

	return violations, nil
}

func validationBoundaryViolationsInFile(config rulekit.Config, filename string) ([]ValidationBoundaryViolation, error) {
	fileSet := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fileSet, filename, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}

	if filepath.Base(filename) == "validation.go" {
		return validationFileBoundaryViolations(fileSet, filename, config.WithDefaults(), parsedFile), nil
	}
	return nonValidationFileBoundaryViolations(fileSet, filename, parsedFile), nil
}

func validationFileBoundaryViolations(fileSet *token.FileSet, filename string, config rulekit.Config, parsedFile *ast.File) []ValidationBoundaryViolation {
	var violations []ValidationBoundaryViolation
	for _, decl := range parsedFile.Decls {
		switch typed := decl.(type) {
		case *ast.GenDecl:
			if typed.Tok == token.IMPORT {
				continue
			}
			if typed.Tok == token.TYPE {
				position := fileSet.Position(typed.Pos())
				violations = append(violations, ValidationBoundaryViolation{
					File:    rulekit.DisplayFilename(filename),
					Line:    position.Line,
					Message: "validation.go must not declare type",
				})
				continue
			}
			for _, spec := range typed.Specs {
				for _, name := range validationSpecNames(spec) {
					if ast.IsExported(name.Name) {
						position := fileSet.Position(name.Pos())
						violations = append(violations, ValidationBoundaryViolation{
							File:    rulekit.DisplayFilename(filename),
							Line:    position.Line,
							Message: fmt.Sprintf("validation.go %s %s must be private", typed.Tok, name.Name),
						})
					}
				}
			}
		case *ast.FuncDecl:
			position := fileSet.Position(typed.Pos())
			if typed.Recv != nil {
				if typed.Name.Name == "Resolve" {
					violations = append(violations, validationResolveMethodViolations(fileSet, filename, config, typed)...)
					continue
				}
				violations = append(violations, ValidationBoundaryViolation{
					File:    rulekit.DisplayFilename(filename),
					Line:    position.Line,
					Message: "validation.go must not declare receiver method " + typed.Name.Name,
				})
				continue
			}
			if !isValidationFunctionName(typed.Name.Name) {
				violations = append(violations, ValidationBoundaryViolation{
					File:    rulekit.DisplayFilename(filename),
					Line:    position.Line,
					Message: fmt.Sprintf("validation.go function %s must start with validate or normalize", typed.Name.Name),
				})
			}
		}
	}
	return violations
}

func validationResolveMethodViolations(fileSet *token.FileSet, filename string, config rulekit.Config, funcDecl *ast.FuncDecl) []ValidationBoundaryViolation {
	if isValidationResolveSignature(config, funcDecl) {
		return nil
	}
	position := fileSet.Position(funcDecl.Pos())
	return []ValidationBoundaryViolation{{
		File:    rulekit.DisplayFilename(filename),
		Line:    position.Line,
		Message: config.ValidationResolve.Message,
	}}
}

func isValidationResolveSignature(config rulekit.Config, funcDecl *ast.FuncDecl) bool {
	if funcDecl.Type.Params == nil || funcDecl.Type.Results == nil || len(funcDecl.Type.Results.List) != 1 {
		return false
	}
	result, ok := funcDecl.Type.Results.List[0].Type.(*ast.ArrayType)
	if !ok || result.Len != nil || !isBuiltinErrorType(result.Elt) {
		return false
	}
	params := funcDecl.Type.Params.List
	if len(params) == 1 {
		return isConfiguredSelectorType(params[0].Type, config.ValidationResolve.ContextPackage, config.ValidationResolve.ContextType)
	}
	if len(params) == 2 {
		return isConfiguredSelectorType(params[0].Type, config.ValidationResolve.ContextPackage, config.ValidationResolve.ContextType) &&
			isConfiguredPointerSelectorType(params[1].Type, config.ValidationResolve.PathBufferPackage, config.ValidationResolve.PathBufferType)
	}
	return false
}

func isConfiguredSelectorType(expr ast.Expr, pkgName string, typeName string) bool {
	selector, ok := expr.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != typeName {
		return false
	}
	ident, ok := selector.X.(*ast.Ident)
	return ok && ident.Name == pkgName
}

func isConfiguredPointerSelectorType(expr ast.Expr, pkgName string, typeName string) bool {
	star, ok := expr.(*ast.StarExpr)
	if !ok {
		return false
	}
	return isConfiguredSelectorType(star.X, pkgName, typeName)
}

func isBuiltinErrorType(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == "error"
}

func nonValidationFileBoundaryViolations(fileSet *token.FileSet, filename string, parsedFile *ast.File) []ValidationBoundaryViolation {
	var violations []ValidationBoundaryViolation
	for _, decl := range parsedFile.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Recv != nil || !isValidationOwnershipName(funcDecl.Name.Name) {
			continue
		}
		position := fileSet.Position(funcDecl.Pos())
		violations = append(violations, ValidationBoundaryViolation{
			File:    rulekit.DisplayFilename(filename),
			Line:    position.Line,
			Message: fmt.Sprintf("validate*/normalize* function %s must be declared in validation.go", funcDecl.Name.Name),
		})
	}
	return violations
}

func validationSpecNames(spec ast.Spec) []*ast.Ident {
	switch typed := spec.(type) {
	case *ast.ValueSpec:
		return typed.Names
	case *ast.TypeSpec:
		return []*ast.Ident{typed.Name}
	}
	return nil
}

func isValidationFunctionName(name string) bool {
	return hasValidationPrefix(name, "validate") || hasValidationPrefix(name, "normalize")
}

func isValidationOwnershipName(name string) bool {
	return isValidationFunctionName(name) ||
		hasValidationPrefix(name, "Validate") ||
		hasValidationPrefix(name, "Normalize")
}

func hasValidationPrefix(name string, prefix string) bool {
	return strings.HasPrefix(name, prefix) && len(name) > len(prefix)
}
