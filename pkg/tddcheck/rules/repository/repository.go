package repository

import (
	"fmt"
	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rulekit"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// Rules declares boundary rules for module repository files.
type Rules struct {
	root   string
	config rulekit.Config
}

// New creates rules for the supplied module root.
func New(root string, options ...rulekit.Option) Rules {
	values := rulekit.NewRuleOptions(root, options...)
	return Rules{root: values.Root, config: values.Config}
}

// RepositoryBoundaryViolation describes one repository boundary violation.
type RepositoryBoundaryViolation struct {
	File    string
	Line    int
	Message string
}

// AssertRepositoryBoundary fails the test when module repository boundaries are violated.
func (r Rules) AssertRepositoryBoundary(t *testing.T) {
	t.Helper()

	violations, err := r.RepositoryBoundaryViolations()
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

	t.Fatalf("invalid repository boundaries:\n  - %s", strings.Join(lines, "\n  - "))
}

// RepositoryBoundaryViolations returns all module repository boundary violations.
func (r Rules) RepositoryBoundaryViolations() ([]RepositoryBoundaryViolation, error) {
	moduleDirs, err := rulekit.ModulePackageDirs(r.root, "Rules", r.config)
	if err != nil {
		return nil, err
	}

	var violations []RepositoryBoundaryViolation
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
			fileViolations, err := repositoryBoundaryViolationsInFile(r.config, file)
			if err != nil {
				return nil, err
			}
			violations = append(violations, fileViolations...)
		}
	}

	return violations, nil
}

func repositoryBoundaryViolationsInFile(config rulekit.Config, filename string) ([]RepositoryBoundaryViolation, error) {
	fileSet := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fileSet, filename, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}

	var violations []RepositoryBoundaryViolation
	isRepositoryFile := filepath.Base(filename) == "repository.go"
	if isRepositoryFile {
		violations = append(violations, repositoryDeclarationBoundaryViolations(fileSet, filename, config.WithDefaults(), parsedFile)...)
	} else {
		violations = append(violations, nonRepositoryDeclarationBoundaryViolations(fileSet, filename, config.WithDefaults(), parsedFile)...)
		return violations, nil
	}

	for _, importSpec := range parsedFile.Imports {
		importPath := strings.Trim(importSpec.Path.Value, `"`)
		if !isForbiddenRepositoryImport(config, importPath) {
			continue
		}
		position := fileSet.Position(importSpec.Pos())
		violations = append(violations, RepositoryBoundaryViolation{
			File:    rulekit.DisplayFilename(filename),
			Line:    position.Line,
			Message: "repository.go must not import " + importPath,
		})
	}

	for _, decl := range parsedFile.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		position := fileSet.Position(funcDecl.Pos())
		if isMapperFunctionName(funcDecl.Name.Name) {
			violations = append(violations, RepositoryBoundaryViolation{
				File:    rulekit.DisplayFilename(filename),
				Line:    position.Line,
				Message: fmt.Sprintf("mapper function %s must be declared in mapper.go", funcDecl.Name.Name),
			})
		}
		if returnsDTO(funcDecl.Type.Results) {
			violations = append(violations, RepositoryBoundaryViolation{
				File:    rulekit.DisplayFilename(filename),
				Line:    position.Line,
				Message: fmt.Sprintf("repository function %s must not return DTO", funcDecl.Name.Name),
			})
		}
	}

	ast.Inspect(parsedFile, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || !isBunOrderCall(call) {
			return true
		}
		for _, arg := range call.Args {
			if !isUnsafeBunOrderArg(arg) {
				continue
			}
			position := fileSet.Position(arg.Pos())
			violations = append(violations, RepositoryBoundaryViolation{
				File:    rulekit.DisplayFilename(filename),
				Line:    position.Line,
				Message: "repository.go must use OrderExpr or OrderBy for complex order expressions",
			})
		}
		return true
	})

	return violations, nil
}

func repositoryDeclarationBoundaryViolations(fileSet *token.FileSet, filename string, config rulekit.Config, parsedFile *ast.File) []RepositoryBoundaryViolation {
	var violations []RepositoryBoundaryViolation
	for _, decl := range parsedFile.Decls {
		switch typed := decl.(type) {
		case *ast.GenDecl:
			if typed.Tok == token.IMPORT {
				continue
			}
			if typed.Tok != token.TYPE {
				position := fileSet.Position(typed.Pos())
				violations = append(violations, RepositoryBoundaryViolation{
					File:    rulekit.DisplayFilename(filename),
					Line:    position.Line,
					Message: "repository.go must only declare repository interfaces, implementation types, constructors, and repository methods",
				})
				continue
			}
			for _, spec := range typed.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || isAllowedRepositoryType(config, typeSpec) {
					continue
				}
				position := fileSet.Position(typeSpec.Pos())
				violations = append(violations, RepositoryBoundaryViolation{
					File:    rulekit.DisplayFilename(filename),
					Line:    position.Line,
					Message: "repository.go type " + typeSpec.Name.Name + " must be a *Repository interface or repository implementation type",
				})
			}
		case *ast.FuncDecl:
			position := fileSet.Position(typed.Pos())
			if typed.Recv != nil {
				if isRepositoryReceiver(config, rulekit.ReceiverTypeName(typed.Recv)) {
					continue
				}
				violations = append(violations, RepositoryBoundaryViolation{
					File:    rulekit.DisplayFilename(filename),
					Line:    position.Line,
					Message: "repository.go receiver method " + typed.Name.Name + " must use a repository receiver",
				})
				continue
			}
			if typed.Name.Name != "newRepository" {
				violations = append(violations, RepositoryBoundaryViolation{
					File:    rulekit.DisplayFilename(filename),
					Line:    position.Line,
					Message: "repository.go package-level function " + typed.Name.Name + " must be newRepository",
				})
			}
		}
	}
	return violations
}

func nonRepositoryDeclarationBoundaryViolations(fileSet *token.FileSet, filename string, config rulekit.Config, parsedFile *ast.File) []RepositoryBoundaryViolation {
	var violations []RepositoryBoundaryViolation
	for _, decl := range parsedFile.Decls {
		switch typed := decl.(type) {
		case *ast.GenDecl:
			if typed.Tok != token.TYPE {
				continue
			}
			for _, spec := range typed.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || !strings.HasSuffix(typeSpec.Name.Name, "Repository") {
					continue
				}
				position := fileSet.Position(typeSpec.Pos())
				violations = append(violations, RepositoryBoundaryViolation{
					File:    rulekit.DisplayFilename(filename),
					Line:    position.Line,
					Message: "*Repository type " + typeSpec.Name.Name + " must be declared in repository.go",
				})
			}
		case *ast.FuncDecl:
			position := fileSet.Position(typed.Pos())
			if typed.Recv == nil {
				if typed.Name.Name == "newRepository" {
					violations = append(violations, RepositoryBoundaryViolation{
						File:    rulekit.DisplayFilename(filename),
						Line:    position.Line,
						Message: "newRepository must be declared in repository.go",
					})
				}
				continue
			}
			if !isRepositoryReceiver(config, rulekit.ReceiverTypeName(typed.Recv)) {
				continue
			}
			violations = append(violations, RepositoryBoundaryViolation{
				File:    rulekit.DisplayFilename(filename),
				Line:    position.Line,
				Message: "repository receiver method " + typed.Name.Name + " must be declared in repository.go",
			})
		}
	}
	return violations
}

func isAllowedRepositoryType(config rulekit.Config, typeSpec *ast.TypeSpec) bool {
	if typeSpec.Assign.IsValid() {
		return false
	}
	name := typeSpec.Name.Name
	if rulekit.StringIn(name, config.RepositoryImplementationNames) ||
		strings.HasSuffix(name, "repository") ||
		strings.HasSuffix(name, "RepositoryImpl") ||
		(!ast.IsExported(name) && strings.HasSuffix(name, "Repository")) {
		return true
	}
	if strings.HasSuffix(name, "Repository") {
		_, ok := typeSpec.Type.(*ast.InterfaceType)
		return ok
	}
	return false
}

func isRepositoryReceiver(config rulekit.Config, name string) bool {
	return rulekit.StringIn(name, config.RepositoryImplementationNames) ||
		strings.HasSuffix(name, "RepositoryImpl") ||
		strings.HasSuffix(name, "repository") ||
		strings.HasSuffix(name, "Repository")
}

func isForbiddenRepositoryImport(config rulekit.Config, importPath string) bool {
	return rulekit.StringIn(importPath, config.WithDefaults().RepositoryForbiddenImports)
}

func isMapperFunctionName(name string) bool {
	return strings.HasPrefix(name, "To") &&
		(strings.HasSuffix(name, "DTO") ||
			strings.HasSuffix(name, "DTOs") ||
			strings.HasSuffix(name, "Schema") ||
			strings.HasSuffix(name, "Schemas"))
}

func returnsDTO(results *ast.FieldList) bool {
	if results == nil {
		return false
	}
	for _, field := range results.List {
		if exprContainsDTO(field.Type) {
			return true
		}
	}
	return false
}

func exprContainsDTO(expr ast.Expr) bool {
	switch typed := expr.(type) {
	case *ast.Ident:
		return strings.HasSuffix(typed.Name, "DTO") || strings.HasSuffix(typed.Name, "DTOs")
	case *ast.SelectorExpr:
		return strings.HasSuffix(typed.Sel.Name, "DTO") || strings.HasSuffix(typed.Sel.Name, "DTOs")
	case *ast.StarExpr:
		return exprContainsDTO(typed.X)
	case *ast.ArrayType:
		return exprContainsDTO(typed.Elt)
	case *ast.MapType:
		return exprContainsDTO(typed.Key) || exprContainsDTO(typed.Value)
	case *ast.ChanType:
		return exprContainsDTO(typed.Value)
	}
	return false
}

func isBunOrderCall(call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	return ok && selector.Sel.Name == "Order"
}

func isUnsafeBunOrderArg(expr ast.Expr) bool {
	literal, ok := expr.(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return false
	}
	value, err := strconv.Unquote(literal.Value)
	if err != nil {
		return false
	}
	column := value
	if index := strings.IndexByte(column, ' '); index >= 0 {
		column = column[:index]
	}
	return strings.ContainsAny(column, `".()`)
}
