package schema

import (
	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/testkit"
	"reflect"
	"slices"
	"testing"
)

func TestModuleSchemaRulesBoundaryViolations(t *testing.T) {
	root := t.TempDir()

	testkit.WriteFile(t, root, "good/schema.go", `package good

type userSchema struct {}
type stringList []string

func Schema() string { return "" }
func CreateIndexes() error { return nil }
func (list *stringList) Scan(value any) error { return nil }
`)
	testkit.WriteFile(t, root, "bad/schema.go", `package bad

const table = "bad"
var enabled = true

type userSchema struct {}
type UserDTO struct {}
type User struct {}

func Schema() string { return "" }
func CreateIndexes() error { return nil }
func helper() {}
func (user User) Name() string { return "" }
`)
	testkit.WriteFile(t, root, "bad/model.go", `package bad

type auditSchema struct {}
func Schema() string { return "" }
func CreateIndexes() error { return nil }
func (userSchema) ID() string { return "" }
`)

	violations, err := (ModuleSchemaRules{Root: root}).SchemaBoundaryViolations()
	if err != nil {
		t.Fatal(err)
	}

	got := schemaViolationMessages(violations)
	want := []string{
		"Schema must be declared in schema.go",
		"CreateIndexes must be declared in schema.go",
		"schema receiver method ID must be declared in schema.go",
		"schema type auditSchema must be declared in schema.go",
		"schema.go must only declare schema types, Schema, CreateIndexes, and local type methods",
		"schema.go must only declare schema types, Schema, CreateIndexes, and local type methods",
		"schema.go type UserDTO must be a *Schema type or private schema helper type",
		"schema.go type User must be a *Schema type or private schema helper type",
		"schema.go package-level function helper must be Schema or CreateIndexes",
	}
	slices.Sort(got)
	slices.Sort(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %v, want %v", got, want)
	}
}

func TestModuleSchemaRulesRequiresRoot(t *testing.T) {
	_, err := (ModuleSchemaRules{}).SchemaBoundaryViolations()
	if err == nil {
		t.Fatal("expected error for empty root")
	}
}

func schemaViolationMessages(violations []SchemaBoundaryViolation) []string {
	messages := make([]string, 0, len(violations))
	for _, violation := range violations {
		messages = append(messages, violation.Message)
	}
	return messages
}
