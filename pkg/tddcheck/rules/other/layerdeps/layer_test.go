package layerdeps

import (
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rulekit"
	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/testkit"
)

func TestRulesViolations(t *testing.T) {
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

	violations, err := New(root).LayerDependencyViolations()
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

func TestRulesDetectsLayerBelowInternalFromModuleRoot(t *testing.T) {
	root := t.TempDir()
	importPrefix, err := rulekit.ModulePath()
	if err != nil {
		t.Fatal(err)
	}

	testkit.WriteFile(t, root, "internal/domain/user/service.go", `package user

import "`+importPrefix+`/internal/adapter/httpauth"
`)

	violations, err := New(root).LayerDependencyViolations()
	if err != nil {
		t.Fatal(err)
	}

	got := layerViolationMessages(violations)
	want := []string{"domain must not import adapter"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %v, want %v", got, want)
	}
}

func TestRulesRequiresRoot(t *testing.T) {
	_, err := New("").LayerDependencyViolations()
	if err == nil {
		t.Fatal("expected error for empty root")
	}
}

func TestModuleImportPrefixesIncludesConfiguredRoot(t *testing.T) {
	projectRoot, err := rulekit.FindProjectRoot()
	if err != nil {
		t.Fatal(err)
	}

	prefixes := moduleImportPrefixes("example.com/project", filepath.Join(projectRoot, "app"))
	_, rel, ok := moduleImportLayer(prefixes, "example.com/project/app/adapter/httpapi")
	if !ok {
		t.Fatalf("expected configured root import prefix in %v", prefixes)
	}
	if rel != "adapter/httpapi" {
		t.Fatalf("rel = %q, want %q", rel, "adapter/httpapi")
	}
}

func layerViolationMessages(violations []LayerDependencyViolation) []string {
	messages := make([]string, 0, len(violations))
	for _, violation := range violations {
		messages = append(messages, violation.Message)
	}
	return messages
}
