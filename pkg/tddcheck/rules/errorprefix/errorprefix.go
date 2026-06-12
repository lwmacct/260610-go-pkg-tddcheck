package errorprefix

import (
	"fmt"
	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rulekit"
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
	"strings"
	"testing"
)

// Rules declares naming rules for module errors.go files.
type Rules struct {
	root   string
	config rulekit.Config
}

// New creates rules for the supplied module root.
func New(root string, options ...rulekit.Option) Rules {
	values := rulekit.NewRuleOptions(root, options...)
	return Rules{root: values.Root, config: values.Config}
}

// ErrorPrefixViolation describes one error variable that does not match Err rules.
type ErrorPrefixViolation struct {
	File string
	Line int
	Name string
}

// AssertErrorPrefix fails the test when a package-level error variable in
// layered module errors.go does not start with Err.
func (r Rules) AssertErrorPrefix(t *testing.T) {
	t.Helper()

	violations, err := r.ErrorPrefixViolations()
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) == 0 {
		return
	}

	lines := make([]string, 0, len(violations))
	for _, violation := range violations {
		lines = append(lines, fmt.Sprintf(
			"%s:%d: error variable %s must start with Err",
			violation.File,
			violation.Line,
			violation.Name,
		))
	}

	t.Fatalf("invalid error variable names:\n  - %s", strings.Join(lines, "\n  - "))
}

// ErrorPrefixViolations returns all package-level error variables in module
// errors.go files that do not start with Err.
func (r Rules) ErrorPrefixViolations() ([]ErrorPrefixViolation, error) {
	matches, err := rulekit.ModuleFiles(r.root, "Rules", r.config, func(name string) bool { return name == "errors.go" })
	if err != nil {
		return nil, err
	}

	var violations []ErrorPrefixViolation
	for _, file := range matches {
		fileViolations, err := errorPrefixViolationsInFile(file)
		if err != nil {
			return nil, err
		}
		violations = append(violations, fileViolations...)
	}

	return violations, nil
}

func errorPrefixViolationsInFile(filename string) ([]ErrorPrefixViolation, error) {
	fileSet := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fileSet, filename, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}

	var violations []ErrorPrefixViolation
	for _, decl := range parsedFile.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			continue
		}

		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok || !isErrorValueSpec(valueSpec) {
				continue
			}

			for _, name := range valueSpec.Names {
				if strings.HasPrefix(name.Name, "Err") {
					continue
				}

				position := fileSet.Position(name.Pos())
				violations = append(violations, ErrorPrefixViolation{
					File: rulekit.DisplayFilename(filename),
					Line: position.Line,
					Name: name.Name,
				})
			}
		}
	}

	return violations, nil
}

func isErrorValueSpec(spec *ast.ValueSpec) bool {
	if isErrorType(spec.Type) {
		return true
	}
	return slices.ContainsFunc(spec.Values, isErrorConstructor)
}

func isErrorType(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == "error"
}

func isErrorConstructor(expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}

	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	pkg, ok := selector.X.(*ast.Ident)
	if !ok {
		return false
	}

	return (pkg.Name == "errors" && selector.Sel.Name == "New") ||
		(pkg.Name == "fmt" && selector.Sel.Name == "Errorf")
}
