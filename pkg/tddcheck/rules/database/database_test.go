package database

import (
	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/testkit"
	"reflect"
	"slices"
	"testing"
)

func TestRulesBoundaryViolations(t *testing.T) {
	root := t.TempDir()

	testkit.WriteFile(t, root, "good/service_test.go", `package good

func testDB() {}
`)
	testkit.WriteFile(t, root, "bad/service_test.go", `package bad

func testDB(t T) {
	_, _ = database.OpenSQLite(ctx, filepath.Join(t.TempDir(), "test.db"))
	cfg.Server.DB.SQLite = filepath.Join(t.TempDir(), "test.db")
}
`)
	testkit.WriteFile(t, root, "internal/infra/database/database_test.go", `package database

func TestSQLitePath(t T) {
	_, _ = database.OpenSQLite(ctx, filepath.Join(t.TempDir(), "test.db"))
}
`)

	violations, err := New(root).DatabaseTestBoundaryViolations()
	if err != nil {
		t.Fatal(err)
	}

	got := databaseTestViolationMessages(violations)
	want := []string{
		"ordinary SQLite config tests must use dbtest.Open or explicit test exemption",
		"ordinary SQLite tests must use dbtest.Open",
	}
	slices.Sort(got)
	slices.Sort(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %v, want %v", got, want)
	}
}

func TestRulesRequiresRoot(t *testing.T) {
	_, err := New("").DatabaseTestBoundaryViolations()
	if err == nil {
		t.Fatal("expected error for empty root")
	}
}

func databaseTestViolationMessages(violations []DatabaseTestViolation) []string {
	messages := make([]string, 0, len(violations))
	for _, violation := range violations {
		messages = append(messages, violation.Message)
	}
	return messages
}
