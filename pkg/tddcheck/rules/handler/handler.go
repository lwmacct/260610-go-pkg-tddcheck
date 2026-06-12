package handler

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

// ModuleHandlerRules declares boundary rules for module handler files.
type ModuleHandlerRules struct {
	// Root is the layered module root directory. Relative paths are resolved from go.mod.
	Root   string
	Config rulekit.Config
}

// HandlerBoundaryViolation describes one handler boundary violation.
type HandlerBoundaryViolation struct {
	File    string
	Line    int
	Message string
}

// AssertHandlerBoundary fails the test when module handler boundaries are violated.
func (r ModuleHandlerRules) AssertHandlerBoundary(t *testing.T) {
	t.Helper()

	violations, err := r.HandlerBoundaryViolations()
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

	t.Fatalf("invalid handler boundaries:\n  - %s", strings.Join(lines, "\n  - "))
}

// HandlerBoundaryViolations returns all module handler boundary violations.
func (r ModuleHandlerRules) HandlerBoundaryViolations() ([]HandlerBoundaryViolation, error) {
	moduleDirs, err := rulekit.ModulePackageDirs(r.Root, "ModuleHandlerRules", r.Config)
	if err != nil {
		return nil, err
	}

	var violations []HandlerBoundaryViolation
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
			fileViolations, err := handlerBoundaryViolationsInFile(r.Config, file)
			if err != nil {
				return nil, err
			}
			violations = append(violations, fileViolations...)
		}
	}

	return violations, nil
}

func handlerBoundaryViolationsInFile(config rulekit.Config, filename string) ([]HandlerBoundaryViolation, error) {
	fileSet := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fileSet, filename, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}

	var violations []HandlerBoundaryViolation
	base := filepath.Base(filename)
	if isHandlerBoundaryFile(base) {
		violations = append(violations, handlerDeclarationBoundaryViolations(fileSet, filename, parsedFile)...)
	} else {
		violations = append(violations, nonHandlerDeclarationBoundaryViolations(fileSet, filename, parsedFile)...)
		return violations, nil
	}
	for _, importSpec := range parsedFile.Imports {
		importPath := strings.Trim(importSpec.Path.Value, `"`)
		if !isForbiddenHandlerImport(config, importPath) {
			continue
		}
		position := fileSet.Position(importSpec.Pos())
		violations = append(violations, HandlerBoundaryViolation{
			File:    rulekit.DisplayFilename(filename),
			Line:    position.Line,
			Message: "handler file must not import " + importPath,
		})
	}

	ast.Inspect(parsedFile, func(node ast.Node) bool {
		selector, ok := node.(*ast.SelectorExpr)
		if !ok || !isForbiddenHandlerCall(config, selector.Sel.Name) {
			return true
		}
		position := fileSet.Position(selector.Sel.Pos())
		violations = append(violations, HandlerBoundaryViolation{
			File:    rulekit.DisplayFilename(filename),
			Line:    position.Line,
			Message: "handler file must not call " + selector.Sel.Name,
		})
		return true
	})

	ast.Inspect(parsedFile, func(node ast.Node) bool {
		funcDecl, ok := node.(*ast.FuncDecl)
		if !ok || !isHumaHandlerMethod(funcDecl) {
			return true
		}
		violations = append(violations, humaHandlerSignatureViolations(fileSet, filename, funcDecl)...)
		return true
	})

	ast.Inspect(parsedFile, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || !isHumaBadRequestCall(call) || !hasRequiredValidationMessage(call) {
			return true
		}
		position := fileSet.Position(call.Pos())
		violations = append(violations, HandlerBoundaryViolation{
			File:    rulekit.DisplayFilename(filename),
			Line:    position.Line,
			Message: "handler file must use Huma validation tags or Resolver for required request fields",
		})
		return true
	})

	ast.Inspect(parsedFile, func(node ast.Node) bool {
		expr, ok := node.(ast.Expr)
		if !ok {
			return true
		}
		bodyType, ok := httpapiBodyWrapperType(expr)
		if !ok {
			return true
		}
		typeName, typeOK := protocolDTOTypeName(bodyType)
		if typeOK && strings.HasSuffix(typeName, "DTO") {
			return true
		}
		position := fileSet.Position(bodyType.Pos())
		if typeName == "" {
			typeName = "anonymous"
		}
		violations = append(violations, HandlerBoundaryViolation{
			File:    rulekit.DisplayFilename(filename),
			Line:    position.Line,
			Message: "handler file HTTP body schema type " + typeName + " must use DTO suffix",
		})
		return true
	})

	return violations, nil
}

func isHandlerBoundaryFile(base string) bool {
	return base == "handler.go" || strings.HasPrefix(base, "handler.")
}

func handlerDeclarationBoundaryViolations(fileSet *token.FileSet, filename string, parsedFile *ast.File) []HandlerBoundaryViolation {
	base := filepath.Base(filename)
	var violations []HandlerBoundaryViolation
	for _, decl := range parsedFile.Decls {
		switch typed := decl.(type) {
		case *ast.GenDecl:
			if typed.Tok == token.IMPORT {
				continue
			}
			if base != "handler.go" {
				position := fileSet.Position(typed.Pos())
				violations = append(violations, HandlerBoundaryViolation{
					File:    rulekit.DisplayFilename(filename),
					Line:    position.Line,
					Message: "handler.*.go must only declare Handler receiver methods",
				})
				continue
			}
			violations = append(violations, handlerTypeBoundaryViolations(fileSet, filename, typed)...)
		case *ast.FuncDecl:
			position := fileSet.Position(typed.Pos())
			if typed.Recv != nil {
				if rulekit.ReceiverTypeName(typed.Recv) == "Handler" {
					continue
				}
				violations = append(violations, HandlerBoundaryViolation{
					File:    rulekit.DisplayFilename(filename),
					Line:    position.Line,
					Message: "handler file receiver method " + typed.Name.Name + " must use Handler receiver",
				})
				continue
			}
			if base == "handler.go" && typed.Name.Name == "RegisterRoutes" {
				continue
			}
			message := "handler.go must only declare RegisterRoutes and Handler receiver methods"
			if base != "handler.go" {
				message = "handler.*.go must only declare Handler receiver methods"
			}
			violations = append(violations, HandlerBoundaryViolation{
				File:    rulekit.DisplayFilename(filename),
				Line:    position.Line,
				Message: message,
			})
		}
	}
	return violations
}

func handlerTypeBoundaryViolations(fileSet *token.FileSet, filename string, decl *ast.GenDecl) []HandlerBoundaryViolation {
	if decl.Tok != token.TYPE {
		position := fileSet.Position(decl.Pos())
		return []HandlerBoundaryViolation{{
			File:    rulekit.DisplayFilename(filename),
			Line:    position.Line,
			Message: "handler.go must not declare const or var",
		}}
	}

	var violations []HandlerBoundaryViolation
	for _, spec := range decl.Specs {
		typeSpec, ok := spec.(*ast.TypeSpec)
		if !ok || isAllowedHandlerType(typeSpec) {
			continue
		}
		position := fileSet.Position(typeSpec.Pos())
		violations = append(violations, HandlerBoundaryViolation{
			File:    rulekit.DisplayFilename(filename),
			Line:    position.Line,
			Message: "handler.go type " + typeSpec.Name.Name + " must be Handler, RouteDeps, RouteConfig, *Func, dependency interface, or protocol input/output type",
		})
	}
	for _, spec := range decl.Specs {
		typeSpec, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		violations = append(violations, handlerProtocolWrapperViolations(fileSet, filename, typeSpec)...)
		violations = append(violations, handlerBodyFieldDTOViolations(fileSet, filename, typeSpec)...)
	}
	return violations
}

func handlerProtocolWrapperViolations(fileSet *token.FileSet, filename string, typeSpec *ast.TypeSpec) []HandlerBoundaryViolation {
	structType, ok := typeSpec.Type.(*ast.StructType)
	if !ok {
		return nil
	}
	name := typeSpec.Name.Name
	if strings.HasSuffix(name, "Response") && hasOnlyBodyField(structType) {
		position := fileSet.Position(typeSpec.Pos())
		return []HandlerBoundaryViolation{{
			File:    rulekit.DisplayFilename(filename),
			Line:    position.Line,
			Message: "handler.go response type " + name + " must use httpapi.Body[T] unless it declares header or status fields",
		}}
	}
	if strings.HasSuffix(name, "Input") && hasOnlyBodyField(structType) {
		position := fileSet.Position(typeSpec.Pos())
		return []HandlerBoundaryViolation{{
			File:    rulekit.DisplayFilename(filename),
			Line:    position.Line,
			Message: "handler.go body-only input type " + name + " must use httpapi.BodyInput[T]",
		}}
	}
	return nil
}

func handlerBodyFieldDTOViolations(fileSet *token.FileSet, filename string, typeSpec *ast.TypeSpec) []HandlerBoundaryViolation {
	structType, ok := typeSpec.Type.(*ast.StructType)
	if !ok || structType.Fields == nil {
		return nil
	}
	var violations []HandlerBoundaryViolation
	for _, field := range structType.Fields.List {
		if fieldName(field) != "Body" {
			continue
		}
		typeName, ok := protocolDTOTypeName(field.Type)
		if ok && strings.HasSuffix(typeName, "DTO") {
			continue
		}
		if typeName == "" {
			typeName = "anonymous"
		}
		position := fileSet.Position(field.Type.Pos())
		violations = append(violations, HandlerBoundaryViolation{
			File:    rulekit.DisplayFilename(filename),
			Line:    position.Line,
			Message: "handler.go Body field type " + typeName + " must use DTO suffix",
		})
	}
	return violations
}

func hasOnlyBodyField(structType *ast.StructType) bool {
	fields := namedStructFields(structType)
	return len(fields) == 1 && fieldName(fields[0]) == "Body"
}

func namedStructFields(structType *ast.StructType) []*ast.Field {
	if structType.Fields == nil {
		return nil
	}
	fields := make([]*ast.Field, 0, len(structType.Fields.List))
	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			continue
		}
		fields = append(fields, field)
	}
	return fields
}

func fieldName(field *ast.Field) string {
	if field == nil || len(field.Names) == 0 {
		return ""
	}
	return field.Names[0].Name
}

func isHumaHandlerMethod(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Recv == nil || rulekit.ReceiverTypeName(funcDecl.Recv) != "Handler" {
		return false
	}
	if funcDecl.Type.Params == nil || len(funcDecl.Type.Params.List) != 2 {
		return false
	}
	if funcDecl.Type.Results == nil || len(funcDecl.Type.Results.List) != 2 {
		return false
	}
	return isContextType(funcDecl.Type.Params.List[0].Type) &&
		isPointerType(funcDecl.Type.Params.List[1].Type) &&
		isPointerType(funcDecl.Type.Results.List[0].Type) &&
		isHandlerErrorType(funcDecl.Type.Results.List[1].Type)
}

func humaHandlerSignatureViolations(fileSet *token.FileSet, filename string, funcDecl *ast.FuncDecl) []HandlerBoundaryViolation {
	input := funcDecl.Type.Params.List[1]
	if _, ok := input.Type.(*ast.StarExpr).X.(*ast.StructType); ok {
		position := fileSet.Position(input.Type.Pos())
		return []HandlerBoundaryViolation{{
			File:    rulekit.DisplayFilename(filename),
			Line:    position.Line,
			Message: "Huma handler " + funcDecl.Name.Name + " must use httpapi.EmptyInput instead of *struct{}",
		}}
	}
	return nil
}

func isPointerType(expr ast.Expr) bool {
	_, ok := expr.(*ast.StarExpr)
	return ok
}

func httpapiBodyWrapperType(expr ast.Expr) (ast.Expr, bool) {
	switch typed := expr.(type) {
	case *ast.StarExpr:
		return httpapiBodyWrapperType(typed.X)
	case *ast.IndexExpr:
		if isHTTPAPIBodyWrapper(typed.X) {
			return typed.Index, true
		}
	case *ast.IndexListExpr:
		if isHTTPAPIBodyWrapper(typed.X) && len(typed.Indices) == 1 {
			return typed.Indices[0], true
		}
	}
	return nil, false
}

func isHTTPAPIBodyWrapper(expr ast.Expr) bool {
	selector, ok := expr.(*ast.SelectorExpr)
	if !ok || (selector.Sel.Name != "Body" && selector.Sel.Name != "BodyInput") {
		return false
	}
	ident, ok := selector.X.(*ast.Ident)
	return ok && ident.Name == "httpapi"
}

func protocolDTOTypeName(expr ast.Expr) (string, bool) {
	switch typed := expr.(type) {
	case *ast.StarExpr:
		return protocolDTOTypeName(typed.X)
	case *ast.ArrayType:
		return protocolDTOTypeName(typed.Elt)
	case *ast.Ident:
		return typed.Name, true
	case *ast.SelectorExpr:
		return typed.Sel.Name, true
	}
	return "", false
}

func hasHTTPBodySchemaName(name string) bool {
	return strings.HasSuffix(name, "DTO") &&
		(strings.Contains(name, "Body") || strings.Contains(name, "Request"))
}

func isContextType(expr ast.Expr) bool {
	selector, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := selector.X.(*ast.Ident)
	return ok && ident.Name == "context" && selector.Sel.Name == "Context"
}

func isHandlerErrorType(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == "error"
}

func isHumaBadRequestCall(call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Error400BadRequest" {
		return false
	}
	ident, ok := selector.X.(*ast.Ident)
	return ok && ident.Name == "huma"
}

func hasRequiredValidationMessage(call *ast.CallExpr) bool {
	for _, arg := range call.Args {
		lit, ok := arg.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			continue
		}
		message := strings.Trim(lit.Value, "`\"")
		if strings.HasPrefix(message, "missing ") ||
			strings.Contains(message, " missing ") ||
			strings.Contains(message, " required") ||
			strings.Contains(message, "required ") {
			return true
		}
	}
	return false
}

func nonHandlerDeclarationBoundaryViolations(fileSet *token.FileSet, filename string, parsedFile *ast.File) []HandlerBoundaryViolation {
	var violations []HandlerBoundaryViolation
	for _, decl := range parsedFile.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Recv == nil || rulekit.ReceiverTypeName(funcDecl.Recv) != "Handler" {
			continue
		}
		position := fileSet.Position(funcDecl.Pos())
		violations = append(violations, HandlerBoundaryViolation{
			File:    rulekit.DisplayFilename(filename),
			Line:    position.Line,
			Message: "Handler receiver method " + funcDecl.Name.Name + " must be declared in handler.go or handler.*.go",
		})
	}
	return violations
}

func isAllowedHandlerType(typeSpec *ast.TypeSpec) bool {
	if typeSpec.Assign.IsValid() {
		return false
	}
	name := typeSpec.Name.Name
	switch name {
	case "Handler", "RouteDeps", "RouteConfig":
		return true
	}
	if strings.HasSuffix(name, "Func") {
		_, ok := typeSpec.Type.(*ast.FuncType)
		return ok
	}
	if _, ok := typeSpec.Type.(*ast.InterfaceType); ok {
		return hasAnySuffix(name, "Service", "Access", "Authorizer", "Policy", "Resolver")
	}
	if hasAnySuffix(name, "DTO") {
		return !hasHTTPBodySchemaName(name)
	}
	return hasAnySuffix(name, "Input", "Response", "ErrorModel")
}

func hasAnySuffix(value string, suffixes ...string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(value, suffix) {
			return true
		}
	}
	return false
}

func isForbiddenHandlerImport(config rulekit.Config, importPath string) bool {
	return rulekit.StringIn(importPath, config.WithDefaults().HandlerForbiddenImports)
}

func isForbiddenHandlerCall(config rulekit.Config, name string) bool {
	return rulekit.StringIn(name, config.WithDefaults().HandlerForbiddenCalls)
}
