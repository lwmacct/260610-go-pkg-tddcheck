package tddcheck

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ModuleDTORules declares naming rules for module dto.go files.
type ModuleDTORules struct {
	// Root is the layered module root directory. Relative paths are resolved from go.mod.
	Root   string
	Config Config
}

// StructSuffixViolation describes one struct name that does not match DTO rules.
type StructSuffixViolation struct {
	File string
	Line int
	Name string
}

// DTOFuncViolation describes one function declared in a dto.go file.
type DTOFuncViolation struct {
	File string
	Line int
	Name string
}

// DTOFileViolation describes one DTO struct declared outside dto.go.
type DTOFileViolation struct {
	File string
	Line int
	Name string
}

// AssertStructSuffix fails the test when a struct in layered module dto.go
// does not end with DTO or DTOs.
func (r ModuleDTORules) AssertStructSuffix(t *testing.T) {
	t.Helper()

	violations, err := r.StructSuffixViolations()
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) == 0 {
		return
	}

	lines := make([]string, 0, len(violations))
	for _, violation := range violations {
		lines = append(lines, fmt.Sprintf(
			"%s:%d: struct %s must end with DTO or DTOs",
			violation.File,
			violation.Line,
			violation.Name,
		))
	}

	t.Fatalf("invalid DTO struct names:\n  - %s", strings.Join(lines, "\n  - "))
}

// AssertNoFuncs fails the test when layered module dto.go files declare functions.
func (r ModuleDTORules) AssertNoFuncs(t *testing.T) {
	t.Helper()

	violations, err := r.FuncViolations()
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) == 0 {
		return
	}

	lines := make([]string, 0, len(violations))
	for _, violation := range violations {
		lines = append(lines, fmt.Sprintf(
			"%s:%d: dto.go must not declare func %s",
			violation.File,
			violation.Line,
			violation.Name,
		))
	}

	t.Fatalf("invalid DTO function declarations:\n  - %s", strings.Join(lines, "\n  - "))
}

// AssertDTOFileOwnership fails the test when DTO structs are declared outside dto.go.
func (r ModuleDTORules) AssertDTOFileOwnership(t *testing.T) {
	t.Helper()

	violations, err := r.FileOwnershipViolations()
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) == 0 {
		return
	}

	lines := make([]string, 0, len(violations))
	for _, violation := range violations {
		lines = append(lines, fmt.Sprintf(
			"%s:%d: DTO struct %s must be declared in dto.go",
			violation.File,
			violation.Line,
			violation.Name,
		))
	}

	t.Fatalf("invalid DTO file ownership:\n  - %s", strings.Join(lines, "\n  - "))
}

// StructSuffixViolations returns all struct names in module dto.go files that do
// not end with DTO or DTOs.
func (r ModuleDTORules) StructSuffixViolations() ([]StructSuffixViolation, error) {
	matches, err := moduleFilesWithConfig(r.Root, "ModuleDTORules", r.Config, func(name string) bool { return name == "dto.go" })
	if err != nil {
		return nil, err
	}

	var violations []StructSuffixViolation
	for _, file := range matches {
		fileViolations, err := structSuffixViolationsInFile(file)
		if err != nil {
			return nil, err
		}
		violations = append(violations, fileViolations...)
	}

	return violations, nil
}

// FuncViolations returns all functions declared in module dto.go files.
func (r ModuleDTORules) FuncViolations() ([]DTOFuncViolation, error) {
	matches, err := moduleFilesWithConfig(r.Root, "ModuleDTORules", r.Config, func(name string) bool { return name == "dto.go" })
	if err != nil {
		return nil, err
	}

	var violations []DTOFuncViolation
	for _, file := range matches {
		fileViolations, err := dtoFuncViolationsInFile(file)
		if err != nil {
			return nil, err
		}
		violations = append(violations, fileViolations...)
	}

	return violations, nil
}

// FileOwnershipViolations returns all DTO structs declared outside dto.go.
func (r ModuleDTORules) FileOwnershipViolations() ([]DTOFileViolation, error) {
	matches, err := moduleFilesWithConfig(r.Root, "ModuleDTORules", r.Config, func(name string) bool {
		return strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
	})
	if err != nil {
		return nil, err
	}

	var violations []DTOFileViolation
	for _, file := range matches {
		if filepath.Base(file) == "dto.go" {
			continue
		}
		fileViolations, err := dtoFileOwnershipViolationsInFile(file)
		if err != nil {
			return nil, err
		}
		violations = append(violations, fileViolations...)
	}

	return violations, nil
}

func structSuffixViolationsInFile(filename string) ([]StructSuffixViolation, error) {
	fileSet := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fileSet, filename, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}

	var violations []StructSuffixViolation
	ast.Inspect(parsedFile, func(node ast.Node) bool {
		typeSpec, ok := node.(*ast.TypeSpec)
		if !ok {
			return true
		}
		if _, ok := typeSpec.Type.(*ast.StructType); !ok {
			return true
		}
		if isDTOName(typeSpec.Name.Name) {
			return true
		}

		position := fileSet.Position(typeSpec.Pos())
		violations = append(violations, StructSuffixViolation{
			File: displayFilename(filename),
			Line: position.Line,
			Name: typeSpec.Name.Name,
		})

		return true
	})

	return violations, nil
}

func dtoFuncViolationsInFile(filename string) ([]DTOFuncViolation, error) {
	fileSet := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fileSet, filename, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}

	var violations []DTOFuncViolation
	for _, decl := range parsedFile.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		position := fileSet.Position(funcDecl.Pos())
		violations = append(violations, DTOFuncViolation{
			File: displayFilename(filename),
			Line: position.Line,
			Name: dtoFuncName(funcDecl),
		})
	}

	return violations, nil
}

func dtoFileOwnershipViolationsInFile(filename string) ([]DTOFileViolation, error) {
	fileSet := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fileSet, filename, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}

	var violations []DTOFileViolation
	ast.Inspect(parsedFile, func(node ast.Node) bool {
		typeSpec, ok := node.(*ast.TypeSpec)
		if !ok || !isDTOName(typeSpec.Name.Name) {
			return true
		}
		if _, ok := typeSpec.Type.(*ast.StructType); !ok {
			return true
		}
		position := fileSet.Position(typeSpec.Pos())
		violations = append(violations, DTOFileViolation{
			File: displayFilename(filename),
			Line: position.Line,
			Name: typeSpec.Name.Name,
		})
		return true
	})

	return violations, nil
}

func isDTOName(name string) bool {
	return strings.HasSuffix(name, "DTO") || strings.HasSuffix(name, "DTOs")
}

func dtoFuncName(funcDecl *ast.FuncDecl) string {
	if funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
		return funcDecl.Name.Name
	}
	return receiverName(funcDecl.Recv.List[0].Type) + "." + funcDecl.Name.Name
}

func receiverName(expr ast.Expr) string {
	switch typed := expr.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.StarExpr:
		return "*" + receiverName(typed.X)
	case *ast.IndexExpr:
		return receiverName(typed.X)
	case *ast.IndexListExpr:
		return receiverName(typed.X)
	case *ast.SelectorExpr:
		return receiverName(typed.X) + "." + typed.Sel.Name
	}
	return "receiver"
}

func displayFilename(filename string) string {
	projectRoot, err := findProjectRoot()
	if err == nil {
		if relative, relErr := filepath.Rel(projectRoot, filename); relErr == nil && !strings.HasPrefix(relative, "..") {
			return filepath.ToSlash(relative)
		}
	}

	return filepath.ToSlash(filename)
}

func findProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd, nil
		}

		parent := filepath.Dir(wd)
		if parent == wd {
			return "", fmt.Errorf("go.mod not found from %s", wd)
		}
		wd = parent
	}
}
