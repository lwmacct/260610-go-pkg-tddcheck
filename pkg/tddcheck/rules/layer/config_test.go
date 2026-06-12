package layer

import (
	"testing"

	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rulekit"
	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/testkit"
)

func TestModuleLayerRulesUsesConfiguredLayers(t *testing.T) {
	root := t.TempDir()
	testkit.WriteFile(t, root, "core/user/service.go", `package user

import "github.com/lwmacct/260610-go-pkg-tddcheck/internal/api/user"
`)

	config := rulekit.Config{
		LayerDirs: []string{"core", "api"},
		LayerRules: []rulekit.LayerDependencyRule{{
			SourceLayer: "core",
			TargetLayer: "api",
			Message:     "core must not import api",
		}},
	}
	violations, err := (ModuleLayerRules{Root: root, Config: config}).LayerDependencyViolations()
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 1 || violations[0].Message != "core must not import api" {
		t.Fatalf("unexpected violations: %#v", violations)
	}
}
