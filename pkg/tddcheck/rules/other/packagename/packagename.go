package packagename

import (
	"fmt"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rulekit"
)

// Rules declares mechanical package naming rules.
type Rules struct {
	root   string
	config rulekit.Config
}

// New 为给定模块根目录创建规则。
func New(root string, options ...rulekit.Option) Rules {
	values := rulekit.NewRuleOptions(root, options...)
	return Rules{root: values.Root, config: values.Config}
}

// PackageNameViolation describes one package clause that does not match its directory.
type PackageNameViolation struct {
	File    string
	Line    int
	Message string
}

// Assert 在包名与目录名不匹配时让测试失败。
func (r Rules) Assert(t *testing.T) {
	t.Helper()

	violations, err := r.PackageNameViolations()
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

	t.Fatalf("invalid package names:\n  - %s", strings.Join(lines, "\n  - "))
}

// PackageNameViolations 返回所有包名违规。
func (r Rules) PackageNameViolations() ([]PackageNameViolation, error) {
	moduleDirs, err := rulekit.ModulePackageDirs(r.root, "Rules", r.config)
	if err != nil {
		return nil, err
	}

	var violations []PackageNameViolation
	for _, moduleDir := range moduleDirs {
		files, err := filepath.Glob(filepath.Join(moduleDir, "*.go"))
		if err != nil {
			return nil, err
		}
		slices.Sort(files)
		expected := utilPackageNameFromDir(moduleDir)
		for _, file := range files {
			fileViolations, err := packageNameViolationsInFile(file, expected)
			if err != nil {
				return nil, err
			}
			violations = append(violations, fileViolations...)
		}
	}

	return violations, nil
}

func packageNameViolationsInFile(filename string, expected string) ([]PackageNameViolation, error) {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, filename, nil, parser.PackageClauseOnly)
	if err != nil {
		return nil, err
	}

	actual := file.Name.Name
	if actual == expected || actual == expected+"_test" {
		return nil, nil
	}

	position := fileSet.Position(file.Name.Pos())
	return []PackageNameViolation{
		{
			File:    filename,
			Line:    position.Line,
			Message: fmt.Sprintf("package name must be %q or %q, got %q", expected, expected+"_test", actual),
		},
	}, nil
}

func utilPackageNameFromDir(dir string) string {
	name := filepath.Base(dir)
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, ".", "_")
	return name
}
