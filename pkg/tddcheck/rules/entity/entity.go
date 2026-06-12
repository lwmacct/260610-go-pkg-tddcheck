package entity

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

// Rules declares mechanical boundary rules for module entity.go files.
type Rules struct {
	root   string
	config rulekit.Config
}

// New creates rules for the supplied module root.
func New(root string, options ...rulekit.Option) Rules {
	values := rulekit.NewRuleOptions(root, options...)
	return Rules{root: values.Root, config: values.Config}
}

// EntityBoundaryViolation describes one entity boundary violation.
type EntityBoundaryViolation struct {
	File    string
	Line    int
	Message string
}

// Assert fails the test when module entity boundaries are violated.
func (r Rules) Assert(t *testing.T) {
	t.Helper()

	violations, err := r.EntityBoundaryViolations()
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

	t.Fatalf("invalid entity boundaries:\n  - %s", strings.Join(lines, "\n  - "))
}

// EntityBoundaryViolations returns all module entity boundary violations.
func (r Rules) EntityBoundaryViolations() ([]EntityBoundaryViolation, error) {
	moduleDirs, err := rulekit.ModulePackageDirs(r.root, "Rules", r.config)
	if err != nil {
		return nil, err
	}

	var violations []EntityBoundaryViolation
	for _, moduleDir := range moduleDirs {
		files, err := filepath.Glob(filepath.Join(moduleDir, "*.go"))
		if err != nil {
			return nil, err
		}
		slices.Sort(files)

		entityTypes, err := entityTypeNamesInDir(files)
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			if strings.HasSuffix(file, "_test.go") {
				continue
			}
			fileViolations, err := entityBoundaryViolationsInFile(file, entityTypes)
			if err != nil {
				return nil, err
			}
			violations = append(violations, fileViolations...)
		}
	}

	return violations, nil
}

func entityTypeNamesInDir(files []string) (map[string]struct{}, error) {
	entityTypes := map[string]struct{}{}
	for _, filename := range files {
		if filepath.Base(filename) != "entity.go" || strings.HasSuffix(filename, "_test.go") {
			continue
		}
		fileSet := token.NewFileSet()
		parsedFile, err := parser.ParseFile(fileSet, filename, nil, parser.SkipObjectResolution)
		if err != nil {
			return nil, err
		}
		for _, decl := range parsedFile.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}
			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if ok && isEntityConcreteType(typeSpec) {
					entityTypes[typeSpec.Name.Name] = struct{}{}
				}
			}
		}
	}
	return entityTypes, nil
}

func entityBoundaryViolationsInFile(filename string, entityTypes map[string]struct{}) ([]EntityBoundaryViolation, error) {
	fileSet := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fileSet, filename, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}

	if filepath.Base(filename) == "entity.go" {
		return entityFileBoundaryViolations(fileSet, filename, parsedFile, entityTypes), nil
	}
	return nonEntityFileBoundaryViolations(fileSet, filename, parsedFile, entityTypes), nil
}

func entityFileBoundaryViolations(fileSet *token.FileSet, filename string, parsedFile *ast.File, entityTypes map[string]struct{}) []EntityBoundaryViolation {
	var violations []EntityBoundaryViolation
	for _, decl := range parsedFile.Decls {
		switch typed := decl.(type) {
		case *ast.GenDecl:
			if typed.Tok == token.IMPORT {
				continue
			}
			if typed.Tok != token.TYPE {
				position := fileSet.Position(typed.Pos())
				violations = append(violations, EntityBoundaryViolation{
					File:    rulekit.DisplayFilename(filename),
					Line:    position.Line,
					Message: "entity.go must only declare concrete types and their methods",
				})
				continue
			}
			for _, spec := range typed.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || isEntityConcreteType(typeSpec) {
					continue
				}
				position := fileSet.Position(typeSpec.Pos())
				violations = append(violations, EntityBoundaryViolation{
					File:    rulekit.DisplayFilename(filename),
					Line:    position.Line,
					Message: fmt.Sprintf("entity.go type %s must be a concrete non-alias type", typeSpec.Name.Name),
				})
			}
		case *ast.FuncDecl:
			position := fileSet.Position(typed.Pos())
			if typed.Recv == nil {
				violations = append(violations, EntityBoundaryViolation{
					File:    rulekit.DisplayFilename(filename),
					Line:    position.Line,
					Message: "entity.go must not declare package-level function " + typed.Name.Name,
				})
				continue
			}
			receiver := rulekit.ReceiverTypeName(typed.Recv)
			if _, ok := entityTypes[receiver]; !ok {
				violations = append(violations, EntityBoundaryViolation{
					File:    rulekit.DisplayFilename(filename),
					Line:    position.Line,
					Message: fmt.Sprintf("entity.go method %s receiver %s must be declared in entity.go", typed.Name.Name, receiver),
				})
			}
		}
	}
	return violations
}

func nonEntityFileBoundaryViolations(fileSet *token.FileSet, filename string, parsedFile *ast.File, entityTypes map[string]struct{}) []EntityBoundaryViolation {
	var violations []EntityBoundaryViolation
	for _, decl := range parsedFile.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Recv == nil {
			continue
		}
		receiver := rulekit.ReceiverTypeName(funcDecl.Recv)
		if _, ok := entityTypes[receiver]; !ok {
			continue
		}
		position := fileSet.Position(funcDecl.Pos())
		violations = append(violations, EntityBoundaryViolation{
			File:    rulekit.DisplayFilename(filename),
			Line:    position.Line,
			Message: fmt.Sprintf("entity method %s.%s must be declared in entity.go", receiver, funcDecl.Name.Name),
		})
	}
	return violations
}

func isEntityConcreteType(typeSpec *ast.TypeSpec) bool {
	if typeSpec.Assign.IsValid() {
		return false
	}
	if _, ok := typeSpec.Type.(*ast.InterfaceType); ok {
		return false
	}
	return true
}
