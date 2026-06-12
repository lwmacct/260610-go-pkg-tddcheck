package service

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

// Rules declares mechanical boundary rules for module Service files.
type Rules struct {
	root   string
	config rulekit.Config
}

// New creates rules for the supplied module root.
func New(root string, options ...rulekit.Option) Rules {
	values := rulekit.NewRuleOptions(root, options...)
	return Rules{root: values.Root, config: values.Config}
}

// ServiceBoundaryViolation describes one service boundary violation.
type ServiceBoundaryViolation struct {
	File    string
	Line    int
	Message string
}

// Assert fails the test when module service boundaries are violated.
func (r Rules) Assert(t *testing.T) {
	t.Helper()

	violations, err := r.ServiceBoundaryViolations()
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

	t.Fatalf("invalid service boundaries:\n  - %s", strings.Join(lines, "\n  - "))
}

// ServiceBoundaryViolations returns all module service boundary violations.
func (r Rules) ServiceBoundaryViolations() ([]ServiceBoundaryViolation, error) {
	moduleDirs, err := rulekit.ModulePackageDirs(r.root, "Rules", r.config)
	if err != nil {
		return nil, err
	}

	var violations []ServiceBoundaryViolation
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
			fileViolations, err := serviceBoundaryViolationsInFile(file)
			if err != nil {
				return nil, err
			}
			violations = append(violations, fileViolations...)
		}
	}

	return violations, nil
}

func serviceBoundaryViolationsInFile(filename string) ([]ServiceBoundaryViolation, error) {
	fileSet := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fileSet, filename, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}

	base := filepath.Base(filename)
	isServiceFile := base == "service.go"
	isServiceKindFile := base != "service.go" && strings.HasPrefix(base, "service.") && strings.HasSuffix(base, ".go")
	var violations []ServiceBoundaryViolation

	for _, decl := range parsedFile.Decls {
		switch typed := decl.(type) {
		case *ast.GenDecl:
			if typed.Tok == token.IMPORT {
				continue
			}
			if isServiceKindFile {
				position := fileSet.Position(typed.Pos())
				violations = append(violations, ServiceBoundaryViolation{
					File:    rulekit.DisplayFilename(filename),
					Line:    position.Line,
					Message: "service.*.go must only declare Service receiver methods",
				})
				continue
			}
			if isServiceFile {
				violations = append(violations, serviceFileGenDeclViolations(fileSet, filename, typed)...)
				continue
			}
			if typed.Tok != token.TYPE {
				continue
			}
			for _, spec := range typed.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || typeSpec.Name.Name != "Service" {
					continue
				}
				position := fileSet.Position(typeSpec.Pos())
				violations = append(violations, ServiceBoundaryViolation{
					File:    rulekit.DisplayFilename(filename),
					Line:    position.Line,
					Message: "type Service must be declared in service.go",
				})
			}
		case *ast.FuncDecl:
			position := fileSet.Position(typed.Pos())
			if isServiceKindFile && !hasServiceReceiver(typed) {
				violations = append(violations, ServiceBoundaryViolation{
					File:    rulekit.DisplayFilename(filename),
					Line:    position.Line,
					Message: "service.*.go must only declare Service receiver methods",
				})
				continue
			}
			if isServiceFile && typed.Name.Name != "NewService" && !hasServiceReceiver(typed) {
				violations = append(violations, ServiceBoundaryViolation{
					File:    rulekit.DisplayFilename(filename),
					Line:    position.Line,
					Message: "service.go must only declare NewService and Service receiver methods",
				})
				continue
			}
			if typed.Name.Name == "NewService" && !isServiceFile {
				violations = append(violations, ServiceBoundaryViolation{
					File:    rulekit.DisplayFilename(filename),
					Line:    position.Line,
					Message: "NewService must be declared in service.go",
				})
				continue
			}
			if hasServiceReceiver(typed) && !isServiceFile && !isServiceKindFile {
				violations = append(violations, ServiceBoundaryViolation{
					File:    rulekit.DisplayFilename(filename),
					Line:    position.Line,
					Message: "Service receiver method " + typed.Name.Name + " must be declared in service.go or service.*.go",
				})
			}
		}
	}

	return violations, nil
}

func serviceFileGenDeclViolations(fileSet *token.FileSet, filename string, decl *ast.GenDecl) []ServiceBoundaryViolation {
	if decl.Tok != token.TYPE {
		position := fileSet.Position(decl.Pos())
		return []ServiceBoundaryViolation{{
			File:    rulekit.DisplayFilename(filename),
			Line:    position.Line,
			Message: "service.go must not declare const or var",
		}}
	}

	var violations []ServiceBoundaryViolation
	for _, spec := range decl.Specs {
		typeSpec, ok := spec.(*ast.TypeSpec)
		if !ok || isServiceFileAllowedType(typeSpec) {
			continue
		}
		position := fileSet.Position(typeSpec.Pos())
		violations = append(violations, ServiceBoundaryViolation{
			File:    rulekit.DisplayFilename(filename),
			Line:    position.Line,
			Message: "service.go type " + typeSpec.Name.Name + " must be Service, Config, Dependencies, interface, or *Func dependency",
		})
	}
	return violations
}

func isServiceFileAllowedType(typeSpec *ast.TypeSpec) bool {
	switch typeSpec.Name.Name {
	case "Service", "Config", "Dependencies":
		return true
	}
	if _, ok := typeSpec.Type.(*ast.InterfaceType); ok {
		return true
	}
	if _, ok := typeSpec.Type.(*ast.FuncType); ok {
		return strings.HasSuffix(typeSpec.Name.Name, "Func")
	}
	return false
}

func hasServiceReceiver(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
		return false
	}
	return serviceReceiverTypeName(funcDecl.Recv.List[0].Type) == "Service"
}

func serviceReceiverTypeName(expr ast.Expr) string {
	switch typed := expr.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.StarExpr:
		return serviceReceiverTypeName(typed.X)
	}
	return ""
}
