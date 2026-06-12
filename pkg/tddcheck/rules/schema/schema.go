package schema

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

// Rules declares mechanical boundary rules for module schema.go files.
type Rules struct {
	root   string
	config rulekit.Config
}

// New creates rules for the supplied module root.
func New(root string, options ...rulekit.Option) Rules {
	values := rulekit.NewRuleOptions(root, options...)
	return Rules{root: values.Root, config: values.Config}
}

// SchemaBoundaryViolation describes one schema boundary violation.
type SchemaBoundaryViolation struct {
	File    string
	Line    int
	Message string
}

// AssertSchemaBoundaries fails the test when module schema boundaries are violated.
func (r Rules) AssertSchemaBoundaries(t *testing.T) {
	t.Helper()

	violations, err := r.SchemaBoundaryViolations()
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

	t.Fatalf("invalid schema boundaries:\n  - %s", strings.Join(lines, "\n  - "))
}

// SchemaBoundaryViolations returns all module schema boundary violations.
func (r Rules) SchemaBoundaryViolations() ([]SchemaBoundaryViolation, error) {
	moduleDirs, err := rulekit.ModulePackageDirs(r.root, "Rules", r.config)
	if err != nil {
		return nil, err
	}

	var violations []SchemaBoundaryViolation
	for _, moduleDir := range moduleDirs {
		files, err := filepath.Glob(filepath.Join(moduleDir, "*.go"))
		if err != nil {
			return nil, err
		}
		slices.Sort(files)

		schemaTypes, err := schemaTypeNamesInDir(files)
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			if strings.HasSuffix(file, "_test.go") {
				continue
			}
			fileViolations, err := schemaBoundaryViolationsInFile(file, schemaTypes)
			if err != nil {
				return nil, err
			}
			violations = append(violations, fileViolations...)
		}
	}

	return violations, nil
}

func schemaTypeNamesInDir(files []string) (map[string]struct{}, error) {
	schemaTypes := map[string]struct{}{}
	for _, filename := range files {
		if filepath.Base(filename) != "schema.go" || strings.HasSuffix(filename, "_test.go") {
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
				if ok && isSchemaFileType(typeSpec.Name.Name) {
					schemaTypes[typeSpec.Name.Name] = struct{}{}
				}
			}
		}
	}
	return schemaTypes, nil
}

func schemaBoundaryViolationsInFile(filename string, schemaTypes map[string]struct{}) ([]SchemaBoundaryViolation, error) {
	fileSet := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fileSet, filename, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}

	if filepath.Base(filename) == "schema.go" {
		return schemaFileBoundaryViolations(fileSet, filename, parsedFile, schemaTypes), nil
	}
	return nonSchemaFileBoundaryViolations(fileSet, filename, parsedFile, schemaTypes), nil
}

func schemaFileBoundaryViolations(fileSet *token.FileSet, filename string, parsedFile *ast.File, schemaTypes map[string]struct{}) []SchemaBoundaryViolation {
	var violations []SchemaBoundaryViolation
	localTypes := schemaLocalTypeNames(parsedFile)
	for _, decl := range parsedFile.Decls {
		switch typed := decl.(type) {
		case *ast.GenDecl:
			if typed.Tok == token.IMPORT {
				continue
			}
			if typed.Tok != token.TYPE {
				position := fileSet.Position(typed.Pos())
				violations = append(violations, SchemaBoundaryViolation{
					File:    rulekit.DisplayFilename(filename),
					Line:    position.Line,
					Message: "schema.go must only declare schema types, Schema, CreateIndexes, and local type methods",
				})
				continue
			}
			for _, spec := range typed.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || isAllowedSchemaFileType(typeSpec.Name.Name) {
					continue
				}
				position := fileSet.Position(typeSpec.Pos())
				violations = append(violations, SchemaBoundaryViolation{
					File:    rulekit.DisplayFilename(filename),
					Line:    position.Line,
					Message: "schema.go type " + typeSpec.Name.Name + " must be a *Schema type or private schema helper type",
				})
			}
		case *ast.FuncDecl:
			position := fileSet.Position(typed.Pos())
			if typed.Recv == nil {
				if isSchemaEntryFunction(typed.Name.Name) {
					continue
				}
				violations = append(violations, SchemaBoundaryViolation{
					File:    rulekit.DisplayFilename(filename),
					Line:    position.Line,
					Message: "schema.go package-level function " + typed.Name.Name + " must be Schema or CreateIndexes",
				})
				continue
			}
			receiver := rulekit.ReceiverTypeName(typed.Recv)
			if _, ok := localTypes[receiver]; ok {
				continue
			}
			violations = append(violations, SchemaBoundaryViolation{
				File:    rulekit.DisplayFilename(filename),
				Line:    position.Line,
				Message: "schema.go method " + typed.Name.Name + " receiver " + receiver + " must be declared in schema.go",
			})
		}
	}
	_ = schemaTypes
	return violations
}

func nonSchemaFileBoundaryViolations(fileSet *token.FileSet, filename string, parsedFile *ast.File, schemaTypes map[string]struct{}) []SchemaBoundaryViolation {
	var violations []SchemaBoundaryViolation
	for _, decl := range parsedFile.Decls {
		switch typed := decl.(type) {
		case *ast.GenDecl:
			if typed.Tok != token.TYPE {
				continue
			}
			for _, spec := range typed.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || !isSchemaFileType(typeSpec.Name.Name) {
					continue
				}
				position := fileSet.Position(typeSpec.Pos())
				violations = append(violations, SchemaBoundaryViolation{
					File:    rulekit.DisplayFilename(filename),
					Line:    position.Line,
					Message: "schema type " + typeSpec.Name.Name + " must be declared in schema.go",
				})
			}
		case *ast.FuncDecl:
			position := fileSet.Position(typed.Pos())
			if typed.Recv == nil {
				if isSchemaEntryFunction(typed.Name.Name) {
					violations = append(violations, SchemaBoundaryViolation{
						File:    rulekit.DisplayFilename(filename),
						Line:    position.Line,
						Message: typed.Name.Name + " must be declared in schema.go",
					})
				}
				continue
			}
			receiver := rulekit.ReceiverTypeName(typed.Recv)
			if _, ok := schemaTypes[receiver]; !ok {
				continue
			}
			violations = append(violations, SchemaBoundaryViolation{
				File:    rulekit.DisplayFilename(filename),
				Line:    position.Line,
				Message: "schema receiver method " + typed.Name.Name + " must be declared in schema.go",
			})
		}
	}
	return violations
}

func schemaLocalTypeNames(parsedFile *ast.File) map[string]struct{} {
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

func isAllowedSchemaFileType(name string) bool {
	return isSchemaFileType(name) || !ast.IsExported(name)
}

func isSchemaFileType(name string) bool {
	return name != "Schema" && (strings.HasSuffix(name, "Schema") || strings.HasSuffix(name, "schema"))
}

func isSchemaEntryFunction(name string) bool {
	return name == "Schema" || name == "CreateIndexes"
}
