package mapper

import (
	"testing"

	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rulekit"
	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/testkit"
)

func TestMapperRulesCanDisableForbiddenImports(t *testing.T) {
	root := t.TempDir()
	testkit.WriteFile(t, root, "feature/mapper.go", `package feature

import "net/http"

func ToUserDTO(_ *http.Request) UserDTO { return UserDTO{} }
`)

	defaultViolations, err := (ModuleMapperRules{Root: root}).MapperBoundaryViolations()
	if err != nil {
		t.Fatal(err)
	}
	if len(defaultViolations) == 0 {
		t.Fatal("expected default forbidden import violation")
	}

	configuredViolations, err := (ModuleMapperRules{
		Root:   root,
		Config: rulekit.Config{MapperForbiddenImports: []string{}},
	}).MapperBoundaryViolations()
	if err != nil {
		t.Fatal(err)
	}
	if len(configuredViolations) != 0 {
		t.Fatalf("expected no configured violations, got %#v", configuredViolations)
	}
}
