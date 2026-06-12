package tddcheck

import "testing"

func TestModuleLayerRulesUsesConfiguredLayers(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "core/user/service.go", `package user

import "github.com/lwmacct/260610-go-pkg-tddcheck/internal/api/user"
`)

	config := Config{
		LayerDirs: []string{"core", "api"},
		LayerRules: []LayerDependencyRule{{
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

func TestMapperRulesCanDisableForbiddenImports(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "feature/mapper.go", `package feature

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
		Config: Config{MapperForbiddenImports: []string{}},
	}).MapperBoundaryViolations()
	if err != nil {
		t.Fatal(err)
	}
	if len(configuredViolations) != 0 {
		t.Fatalf("expected no configured violations, got %#v", configuredViolations)
	}
}

func TestDatabaseTestRulesUsesConfiguredNeedlesAndAllowedPaths(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "db/open_test.go", `package db

func TestOpen(t T) {
	OpenTestDB(tempPath())
}
`)
	config := Config{
		DatabaseTest: DatabaseTestConfig{
			AllowedPaths:      []string{"db/open_test.go"},
			OpenNeedle:        "OpenTestDB",
			TempDirNeedle:     "tempPath()",
			ConfigPathNeedle:  "SetDBPath(tempPath()",
			OpenMessage:       "use shared test db helper",
			ConfigPathMessage: "use shared test db config helper",
		},
	}

	violations, err := (DatabaseTestRules{Root: root, Config: config}).DatabaseTestBoundaryViolations()
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 0 {
		t.Fatalf("expected allowed path, got %#v", violations)
	}
}
