package dto

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rulekit"
)

// Rules declares naming rules for module dto.go files.
type Rules struct {
	root   string
	config rulekit.Config
}

// New 为给定模块根目录创建规则。
func New(root string, options ...rulekit.Option) Rules {
	values := rulekit.NewRuleOptions(root, options...)
	return Rules{root: values.Root, config: values.Config}
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

// RequestsFileViolation describes one legacy requests.go file.
type RequestsFileViolation struct {
	File string
	Line int
}

// ProtocolTagViolation describes one protocol-tagged struct declared outside dto.go.
type ProtocolTagViolation struct {
	File string
	Line int
	Name string
}

// Assert 在任意 DTO 命名或文件归属规则被违反时让测试失败。
func (r Rules) Assert(t *testing.T) {
	t.Helper()

	structs, err := r.StructSuffixViolations()
	if err != nil {
		t.Fatal(err)
	}
	funcs, err := r.FuncViolations()
	if err != nil {
		t.Fatal(err)
	}
	owned, err := r.FileOwnershipViolations()
	if err != nil {
		t.Fatal(err)
	}
	requests, err := r.RequestsFileViolations()
	if err != nil {
		t.Fatal(err)
	}
	protocolTagged, err := r.ProtocolTagViolations()
	if err != nil {
		t.Fatal(err)
	}
	if len(structs) == 0 && len(funcs) == 0 && len(owned) == 0 && len(requests) == 0 && len(protocolTagged) == 0 {
		return
	}

	lines := make([]string, 0, len(structs)+len(funcs)+len(owned)+len(requests)+len(protocolTagged))
	for _, violation := range structs {
		lines = append(lines, fmt.Sprintf(
			"%s:%d: struct %s must end with DTO or DTOs",
			violation.File,
			violation.Line,
			violation.Name,
		))
	}
	for _, violation := range funcs {
		lines = append(lines, fmt.Sprintf(
			"%s:%d: dto.go must not declare func %s",
			violation.File,
			violation.Line,
			violation.Name,
		))
	}
	for _, violation := range owned {
		lines = append(lines, fmt.Sprintf(
			"%s:%d: DTO struct %s must be declared in dto.go",
			violation.File,
			violation.Line,
			violation.Name,
		))
	}
	for _, violation := range requests {
		lines = append(lines, fmt.Sprintf(
			"%s:%d: requests.go is obsolete; protocol DTOs must be declared in dto.go",
			violation.File,
			violation.Line,
		))
	}
	for _, violation := range protocolTagged {
		lines = append(lines, fmt.Sprintf(
			"%s:%d: protocol-tagged struct %s must be declared in dto.go",
			violation.File,
			violation.Line,
			violation.Name,
		))
	}

	t.Fatalf("invalid DTO boundaries:\n  - %s", strings.Join(lines, "\n  - "))
}

// StructSuffixViolations 返回模块 dto.go 文件中所有未以 DTO 或 DTOs 结尾的结构体名称。
func (r Rules) StructSuffixViolations() ([]StructSuffixViolation, error) {
	matches, err := rulekit.ModuleFiles(r.root, "Rules", r.config, func(name string) bool { return name == "dto.go" })
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

// FuncViolations 返回模块 dto.go 文件中声明的所有函数。
func (r Rules) FuncViolations() ([]DTOFuncViolation, error) {
	matches, err := rulekit.ModuleFiles(r.root, "Rules", r.config, func(name string) bool { return name == "dto.go" })
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

// FileOwnershipViolations 返回所有在 dto.go 之外声明的 DTO 结构体。
func (r Rules) FileOwnershipViolations() ([]DTOFileViolation, error) {
	matches, err := rulekit.ModuleFiles(r.root, "Rules", r.config, func(name string) bool {
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

// RequestsFileViolations 返回所有过时的 requests.go 文件。
func (r Rules) RequestsFileViolations() ([]RequestsFileViolation, error) {
	matches, err := rulekit.ModuleFiles(r.root, "Rules", r.config, func(name string) bool { return name == "requests.go" })
	if err != nil {
		return nil, err
	}

	violations := make([]RequestsFileViolation, 0, len(matches))
	for _, file := range matches {
		violations = append(violations, RequestsFileViolation{
			File: rulekit.DisplayFilename(file),
			Line: 1,
		})
	}
	return violations, nil
}

// ProtocolTagViolations 返回所有在 dto.go 之外声明的协议 tagged 结构体。
func (r Rules) ProtocolTagViolations() ([]ProtocolTagViolation, error) {
	matches, err := rulekit.ModuleFiles(r.root, "Rules", r.config, func(name string) bool {
		return strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
	})
	if err != nil {
		return nil, err
	}

	var violations []ProtocolTagViolation
	for _, file := range matches {
		if filepath.Base(file) == "dto.go" {
			continue
		}
		fileViolations, err := protocolTagViolationsInFile(file)
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
			File: rulekit.DisplayFilename(filename),
			Line: position.Line,
			Name: typeSpec.Name.Name,
		})

		return true
	})

	return violations, nil
}

func protocolTagViolationsInFile(filename string) ([]ProtocolTagViolation, error) {
	fileSet := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fileSet, filename, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}

	var violations []ProtocolTagViolation
	ast.Inspect(parsedFile, func(node ast.Node) bool {
		typeSpec, ok := node.(*ast.TypeSpec)
		if !ok {
			return true
		}
		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok || !structHasProtocolTag(structType) {
			return true
		}
		position := fileSet.Position(typeSpec.Pos())
		violations = append(violations, ProtocolTagViolation{
			File: rulekit.DisplayFilename(filename),
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
			File: rulekit.DisplayFilename(filename),
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
			File: rulekit.DisplayFilename(filename),
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

func structHasProtocolTag(structType *ast.StructType) bool {
	if structType.Fields == nil {
		return false
	}
	for _, field := range structType.Fields.List {
		if field.Tag == nil {
			continue
		}
		tag := strings.Trim(field.Tag.Value, "`")
		if hasProtocolTag(tag) {
			return true
		}
	}
	return false
}

func hasProtocolTag(tag string) bool {
	for _, key := range []string{"json", "query", "path", "header", "form"} {
		if strings.Contains(tag, key+":") {
			return true
		}
	}
	return false
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
