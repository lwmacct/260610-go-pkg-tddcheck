package mapper

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

// ModuleMapperRules declares boundary rules for module mapper.go files.
type ModuleMapperRules struct {
	// Root is the layered module root directory. Relative paths are resolved from go.mod.
	Root   string
	Config rulekit.Config
}

// MapperBoundaryViolation describes one mapper boundary violation.
type MapperBoundaryViolation struct {
	File    string
	Line    int
	Message string
}

// AssertMapperBoundary fails the test when module mapper boundaries are violated.
func (r ModuleMapperRules) AssertMapperBoundary(t *testing.T) {
	t.Helper()

	violations, err := r.MapperBoundaryViolations()
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

	t.Fatalf("invalid mapper boundaries:\n  - %s", strings.Join(lines, "\n  - "))
}

// MapperBoundaryViolations returns all module mapper boundary violations.
func (r ModuleMapperRules) MapperBoundaryViolations() ([]MapperBoundaryViolation, error) {
	moduleDirs, err := rulekit.ModulePackageDirs(r.Root, "ModuleMapperRules", r.Config)
	if err != nil {
		return nil, err
	}

	var violations []MapperBoundaryViolation
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
			fileViolations, err := mapperBoundaryViolationsInFile(r.Config, file)
			if err != nil {
				return nil, err
			}
			violations = append(violations, fileViolations...)
		}
	}

	return violations, nil
}

func mapperBoundaryViolationsInFile(config rulekit.Config, filename string) ([]MapperBoundaryViolation, error) {
	fileSet := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fileSet, filename, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}

	switch filepath.Base(filename) {
	case "mapper.go":
		return mapperFileBoundaryViolations(fileSet, filename, config.WithDefaults(), parsedFile), nil
	case "schema.go":
		return nil, nil
	}
	return nonMapperFileBoundaryViolations(fileSet, filename, parsedFile), nil
}

func mapperFileBoundaryViolations(fileSet *token.FileSet, filename string, config rulekit.Config, parsedFile *ast.File) []MapperBoundaryViolation {
	var violations []MapperBoundaryViolation
	for _, importSpec := range parsedFile.Imports {
		importPath := strings.Trim(importSpec.Path.Value, `"`)
		if !isForbiddenMapperImport(config, importPath) {
			continue
		}
		position := fileSet.Position(importSpec.Pos())
		violations = append(violations, MapperBoundaryViolation{
			File:    rulekit.DisplayFilename(filename),
			Line:    position.Line,
			Message: "mapper.go must not import " + importPath,
		})
	}

	for _, decl := range parsedFile.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		position := fileSet.Position(funcDecl.Pos())
		if funcDecl.Recv != nil {
			violations = append(violations, MapperBoundaryViolation{
				File:    rulekit.DisplayFilename(filename),
				Line:    position.Line,
				Message: fmt.Sprintf("mapper function %s must not use a receiver", funcDecl.Name.Name),
			})
		}
		if !strings.HasPrefix(funcDecl.Name.Name, "To") {
			violations = append(violations, MapperBoundaryViolation{
				File:    rulekit.DisplayFilename(filename),
				Line:    position.Line,
				Message: fmt.Sprintf("mapper function %s must start with To", funcDecl.Name.Name),
			})
		}
	}

	return violations
}

func nonMapperFileBoundaryViolations(fileSet *token.FileSet, filename string, parsedFile *ast.File) []MapperBoundaryViolation {
	var violations []MapperBoundaryViolation
	for _, decl := range parsedFile.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Recv != nil || !isMapperFunction(funcDecl) {
			continue
		}
		position := fileSet.Position(funcDecl.Pos())
		violations = append(violations, MapperBoundaryViolation{
			File:    rulekit.DisplayFilename(filename),
			Line:    position.Line,
			Message: fmt.Sprintf("mapper function %s must be declared in mapper.go", funcDecl.Name.Name),
		})
	}
	return violations
}

func isForbiddenMapperImport(config rulekit.Config, importPath string) bool {
	return rulekit.StringIn(importPath, config.WithDefaults().MapperForbiddenImports)
}

func isMapperFunctionName(name string) bool {
	return strings.HasPrefix(name, "To") &&
		(strings.HasSuffix(name, "DTO") ||
			strings.HasSuffix(name, "DTOs") ||
			strings.HasSuffix(name, "Schema") ||
			strings.HasSuffix(name, "Schemas"))
}

func isMapperFunction(funcDecl *ast.FuncDecl) bool {
	return isMapperFunctionName(funcDecl.Name.Name) ||
		isPureMapperSignature(funcDecl.Type)
}

func isPureMapperSignature(funcType *ast.FuncType) bool {
	return (hasMapperType(funcType.Params) || hasMapperType(funcType.Results)) &&
		!hasFlowType(funcType.Params) &&
		!hasFlowType(funcType.Results)
}

func hasMapperType(fields *ast.FieldList) bool {
	if fields == nil {
		return false
	}
	for _, field := range fields.List {
		if isMapperTypeExpr(field.Type) {
			return true
		}
	}
	return false
}

func hasFlowType(fields *ast.FieldList) bool {
	if fields == nil {
		return false
	}
	for _, field := range fields.List {
		if isFlowTypeExpr(field.Type) {
			return true
		}
	}
	return false
}

func isMapperTypeExpr(expr ast.Expr) bool {
	switch typed := expr.(type) {
	case *ast.Ident:
		return isMapperTypeName(typed.Name)
	case *ast.SelectorExpr:
		return isMapperTypeName(typed.Sel.Name)
	case *ast.StarExpr:
		return isMapperTypeExpr(typed.X)
	case *ast.ArrayType:
		return isMapperTypeExpr(typed.Elt)
	default:
		return false
	}
}

func isFlowTypeExpr(expr ast.Expr) bool {
	switch typed := expr.(type) {
	case *ast.Ident:
		return typed.Name == "error" || typed.Name == "bool"
	case *ast.SelectorExpr:
		return isFlowSelector(typed)
	case *ast.StarExpr:
		return isFlowTypeExpr(typed.X)
	case *ast.ArrayType:
		return isFlowTypeExpr(typed.Elt)
	default:
		return false
	}
}

func isFlowSelector(expr *ast.SelectorExpr) bool {
	pkg, ok := expr.X.(*ast.Ident)
	if !ok {
		return false
	}
	return (pkg.Name == "context" && expr.Sel.Name == "Context") ||
		(pkg.Name == "http" && (expr.Sel.Name == "Request" || expr.Sel.Name == "ResponseWriter")) ||
		(pkg.Name == "huma" && expr.Sel.Name == "Context")
}

func isMapperTypeName(name string) bool {
	return strings.HasSuffix(name, "DTO") ||
		strings.HasSuffix(name, "DTOs") ||
		strings.HasSuffix(name, "Schema") ||
		strings.HasSuffix(name, "Schemas")
}
