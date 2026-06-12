package tddcheck

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// ModuleCQRSRules declares naming rules for module cqrs.go files.
type ModuleCQRSRules struct {
	// Root is the layered module root directory. Relative paths are resolved from go.mod.
	Root string
}

// CQRSSuffixViolation describes one struct name that does not match CQRS rules.
type CQRSSuffixViolation struct {
	File string
	Line int
	Name string
}

// CQRSInterfaceNameViolation describes one interface name that does not match CQRS rules.
type CQRSInterfaceNameViolation struct {
	File    string
	Line    int
	Name    string
	Message string
}

// AssertStructSuffix fails the test when a struct in layered module cqrs.go
// does not end with Query, Result, or Command.
func (r ModuleCQRSRules) AssertStructSuffix(t *testing.T) {
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
			"%s:%d: struct %s must end with Query, Result, or Command",
			violation.File,
			violation.Line,
			violation.Name,
		))
	}

	t.Fatalf("invalid CQRS struct names:\n  - %s", strings.Join(lines, "\n  - "))
}

// AssertInterfaceNames fails the test when an interface in layered module
// cqrs.go does not express a use case contract or command/query dependency.
func (r ModuleCQRSRules) AssertInterfaceNames(t *testing.T) {
	t.Helper()

	violations, err := r.InterfaceNameViolations()
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) == 0 {
		return
	}

	lines := make([]string, 0, len(violations))
	for _, violation := range violations {
		lines = append(lines, fmt.Sprintf(
			"%s:%d: interface %s %s",
			violation.File,
			violation.Line,
			violation.Name,
			violation.Message,
		))
	}

	t.Fatalf("invalid CQRS interface names:\n  - %s", strings.Join(lines, "\n  - "))
}

// StructSuffixViolations returns all struct names in module cqrs.go files that
// do not end with Query, Result, or Command.
func (r ModuleCQRSRules) StructSuffixViolations() ([]CQRSSuffixViolation, error) {
	matches, err := moduleFiles(r.Root, "ModuleCQRSRules", func(name string) bool { return name == "cqrs.go" })
	if err != nil {
		return nil, err
	}

	var violations []CQRSSuffixViolation
	for _, file := range matches {
		fileViolations, err := cqrsSuffixViolationsInFile(file)
		if err != nil {
			return nil, err
		}
		violations = append(violations, fileViolations...)
	}

	return violations, nil
}

// InterfaceNameViolations returns all interface names in module cqrs.go files
// that do not express a use case contract or command/query dependency.
func (r ModuleCQRSRules) InterfaceNameViolations() ([]CQRSInterfaceNameViolation, error) {
	matches, err := moduleFiles(r.Root, "ModuleCQRSRules", func(name string) bool { return name == "cqrs.go" })
	if err != nil {
		return nil, err
	}

	var violations []CQRSInterfaceNameViolation
	for _, file := range matches {
		fileViolations, err := cqrsInterfaceNameViolationsInFile(file)
		if err != nil {
			return nil, err
		}
		violations = append(violations, fileViolations...)
	}

	return violations, nil
}

func cqrsSuffixViolationsInFile(filename string) ([]CQRSSuffixViolation, error) {
	fileSet := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fileSet, filename, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}

	var violations []CQRSSuffixViolation
	ast.Inspect(parsedFile, func(node ast.Node) bool {
		typeSpec, ok := node.(*ast.TypeSpec)
		if !ok {
			return true
		}
		if _, ok := typeSpec.Type.(*ast.StructType); !ok {
			return true
		}
		if isCQRSName(typeSpec.Name.Name) {
			return true
		}

		position := fileSet.Position(typeSpec.Pos())
		violations = append(violations, CQRSSuffixViolation{
			File: displayFilename(filename),
			Line: position.Line,
			Name: typeSpec.Name.Name,
		})

		return true
	})

	return violations, nil
}

func cqrsInterfaceNameViolationsInFile(filename string) ([]CQRSInterfaceNameViolation, error) {
	fileSet := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fileSet, filename, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}

	var violations []CQRSInterfaceNameViolation
	ast.Inspect(parsedFile, func(node ast.Node) bool {
		typeSpec, ok := node.(*ast.TypeSpec)
		if !ok {
			return true
		}
		if _, ok := typeSpec.Type.(*ast.InterfaceType); !ok {
			return true
		}
		if isCQRSInterfaceName(typeSpec.Name.Name) {
			return true
		}

		position := fileSet.Position(typeSpec.Pos())
		violations = append(violations, CQRSInterfaceNameViolation{
			File:    displayFilename(filename),
			Line:    position.Line,
			Name:    typeSpec.Name.Name,
			Message: cqrsInterfaceNameViolationMessage(typeSpec.Name.Name),
		})

		return true
	})

	return violations, nil
}

func isCQRSName(name string) bool {
	return strings.HasSuffix(name, "Query") ||
		strings.HasSuffix(name, "Result") ||
		strings.HasSuffix(name, "Command")
}

func isCQRSInterfaceName(name string) bool {
	return strings.HasSuffix(name, "UseCase") ||
		strings.HasSuffix(name, "CommandHandler") ||
		strings.HasSuffix(name, "QueryHandler") ||
		strings.HasSuffix(name, "Access") ||
		strings.HasSuffix(name, "Policy") ||
		strings.HasSuffix(name, "Authorizer")
}

func cqrsInterfaceNameViolationMessage(name string) string {
	if hasForbiddenCQRSInterfaceSuffix(name) {
		return "must not be a Repository, Service, generic Handler, Provider, Resolver, or Interface in cqrs.go"
	}
	return "must end with UseCase, CommandHandler, QueryHandler, Access, Policy, or Authorizer"
}

func hasForbiddenCQRSInterfaceSuffix(name string) bool {
	return strings.HasSuffix(name, "Repository") ||
		strings.HasSuffix(name, "Service") ||
		strings.HasSuffix(name, "Handler") ||
		strings.HasSuffix(name, "Provider") ||
		strings.HasSuffix(name, "Resolver") ||
		strings.HasSuffix(name, "Interface")
}
