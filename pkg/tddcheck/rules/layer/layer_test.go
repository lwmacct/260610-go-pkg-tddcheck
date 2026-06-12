package layer

import (
	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rulekit"
	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/testkit"
	"reflect"
	"slices"
	"testing"
)

func TestModuleLayerRulesViolations(t *testing.T) {
	root := t.TempDir()
	importPrefix, err := rulekit.ModulePath()
	if err != nil {
		t.Fatal(err)
	}

	testkit.WriteFile(t, root, "domain/user/service.go", `package user

import "`+importPrefix+`/internal/adapter/httpauth"
`)
	testkit.WriteFile(t, root, "usecase/report/service.go", `package report

import "`+importPrefix+`/internal/adapter/sshcmd"
`)
	testkit.WriteFile(t, root, "infra/hooks/service.go", `package hooks

import "`+importPrefix+`/internal/domain/identityuser"
`)
	testkit.WriteFile(t, root, "runtime/nodepool/service.go", `package nodepool

import "`+importPrefix+`/internal/adapter/httpauth"
`)
	testkit.WriteFile(t, root, "adapter/httpauth/service.go", `package httpauth

import "`+importPrefix+`/internal/usecase/nodequery"
`)

	violations, err := (ModuleLayerRules{Root: root}).LayerDependencyViolations()
	if err != nil {
		t.Fatal(err)
	}

	got := layerViolationMessages(violations)
	slices.Sort(got)
	want := []string{
		"domain must not import adapter",
		"infra must not import business layers",
		"runtime must not import HTTP API adapter",
		"usecase must not import adapter",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %v, want %v", got, want)
	}
}

func TestModuleLayerRulesRequiresRoot(t *testing.T) {
	_, err := (ModuleLayerRules{}).LayerDependencyViolations()
	if err == nil {
		t.Fatal("expected error for empty root")
	}
}

func layerViolationMessages(violations []LayerDependencyViolation) []string {
	messages := make([]string, 0, len(violations))
	for _, violation := range violations {
		messages = append(messages, violation.Message)
	}
	return messages
}
