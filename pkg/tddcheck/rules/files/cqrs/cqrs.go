package cqrs

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rulekit"
)

// Rules declares naming rules for module cqrs.go files.
type Rules struct {
	root   string
	config rulekit.Config
}

// New 为给定模块根目录创建规则。
func New(root string, options ...rulekit.Option) Rules {
	values := rulekit.NewRuleOptions(root, options...)
	return Rules{root: values.Root, config: values.Config}
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

// Assert 在任意 CQRS 命名规则被违反时让测试失败。
func (r Rules) Assert(t *testing.T) {
	t.Helper()

	structs, err := r.StructSuffixViolations()
	if err != nil {
		t.Fatal(err)
	}
	interfaces, err := r.InterfaceNameViolations()
	if err != nil {
		t.Fatal(err)
	}
	if len(structs) == 0 && len(interfaces) == 0 {
		return
	}

	lines := make([]string, 0, len(structs)+len(interfaces))
	for _, violation := range structs {
		lines = append(lines, fmt.Sprintf(
			"%s:%d: struct %s must end with Query, Result, or Command",
			violation.File,
			violation.Line,
			violation.Name,
		))
	}
	for _, violation := range interfaces {
		lines = append(lines, fmt.Sprintf(
			"%s:%d: interface %s %s",
			violation.File,
			violation.Line,
			violation.Name,
			violation.Message,
		))
	}

	t.Fatalf("invalid CQRS boundaries:\n  - %s", strings.Join(lines, "\n  - "))
}

// StructSuffixViolations 返回模块 cqrs.go 文件中所有未以 Query、Result 或 Command
// 结尾的结构体名称。
func (r Rules) StructSuffixViolations() ([]CQRSSuffixViolation, error) {
	matches, err := rulekit.ModuleFiles(r.root, "Rules", r.config, func(name string) bool { return name == "cqrs.go" })
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

// InterfaceNameViolations 返回模块 cqrs.go 文件中所有未表达用例契约或命令/查询依赖的接口名称。
func (r Rules) InterfaceNameViolations() ([]CQRSInterfaceNameViolation, error) {
	matches, err := rulekit.ModuleFiles(r.root, "Rules", r.config, func(name string) bool { return name == "cqrs.go" })
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
			File: rulekit.DisplayFilename(filename),
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
			File:    rulekit.DisplayFilename(filename),
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
