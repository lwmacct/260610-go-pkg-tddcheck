package publicapi

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

// Rules declares mechanical public API naming rules.
type Rules struct {
	root   string
	config rulekit.Config
}

// New creates rules for the supplied module root.
func New(root string, options ...rulekit.Option) Rules {
	values := rulekit.NewRuleOptions(root, options...)
	return Rules{root: values.Root, config: values.Config}
}

// PublicAPINameViolation describes one public API naming violation.
type PublicAPINameViolation struct {
	File    string
	Line    int
	Message string
}

// Assert fails the test when public API names use reserved responsibility prefixes.
func (r Rules) Assert(t *testing.T) {
	t.Helper()

	violations, err := r.PublicAPINameViolations()
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

	t.Fatalf("invalid public API names:\n  - %s", strings.Join(lines, "\n  - "))
}

// PublicAPINameViolations returns all public API naming violations.
func (r Rules) PublicAPINameViolations() ([]PublicAPINameViolation, error) {
	moduleDirs, err := rulekit.ModulePackageDirs(r.root, "Rules", r.config)
	if err != nil {
		return nil, err
	}

	var violations []PublicAPINameViolation
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
			fileViolations, err := publicAPINameViolationsInFile(file)
			if err != nil {
				return nil, err
			}
			violations = append(violations, fileViolations...)
		}
	}

	return violations, nil
}

func publicAPINameViolationsInFile(filename string) ([]PublicAPINameViolation, error) {
	fileSet := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fileSet, filename, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}

	var violations []PublicAPINameViolation
	for _, decl := range parsedFile.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || !isReservedPublicAPIName(funcDecl.Name.Name) {
			continue
		}
		position := fileSet.Position(funcDecl.Pos())
		violations = append(violations, PublicAPINameViolation{
			File:    rulekit.DisplayFilename(filename),
			Line:    position.Line,
			Message: fmt.Sprintf("public API %s must not use Validate or Normalize prefix", funcDecl.Name.Name),
		})
	}
	return violations, nil
}

func isReservedPublicAPIName(name string) bool {
	return strings.HasPrefix(name, "Validate") || strings.HasPrefix(name, "Normalize")
}
