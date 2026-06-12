package database

import (
	"testing"

	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rulekit"
	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/testkit"
)

func TestDatabaseTestRulesUsesConfiguredNeedlesAndAllowedPaths(t *testing.T) {
	root := t.TempDir()
	testkit.WriteFile(t, root, "db/open_test.go", `package db

func TestOpen(t T) {
	OpenTestDB(tempPath())
}
`)
	config := rulekit.Config{
		DatabaseTest: rulekit.DatabaseTestConfig{
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
