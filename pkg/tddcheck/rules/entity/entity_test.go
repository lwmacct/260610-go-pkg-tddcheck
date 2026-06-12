package entity

import (
	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/testkit"
	"reflect"
	"testing"
)

func TestRulesBoundaryViolations(t *testing.T) {
	root := t.TempDir()

	testkit.WriteFile(t, root, "good/entity.go", `package good

type User struct {}
type Status string

func (user User) Active() bool {
	return true
}
`)
	testkit.WriteFile(t, root, "bad/entity.go", `package bad

const state = "bad"
var current = 1

type User struct {}
type UserAlias = User
type Loader interface {}

func BuildUser() User {
	return User{}
}

func (user User) Active() bool {
	return true
}
`)
	testkit.WriteFile(t, root, "bad/service.go", `package bad

func (user User) Disable() {}
`)

	violations, err := New(root).EntityBoundaryViolations()
	if err != nil {
		t.Fatal(err)
	}

	got := entityViolationMessages(violations)
	want := []string{
		"entity.go must only declare concrete types and their methods",
		"entity.go must only declare concrete types and their methods",
		"entity.go type UserAlias must be a concrete non-alias type",
		"entity.go type Loader must be a concrete non-alias type",
		"entity.go must not declare package-level function BuildUser",
		"entity method User.Disable must be declared in entity.go",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %v, want %v", got, want)
	}
}

func TestRulesRequiresRoot(t *testing.T) {
	_, err := New("").EntityBoundaryViolations()
	if err == nil {
		t.Fatal("expected error for empty root")
	}
}

func entityViolationMessages(violations []EntityBoundaryViolation) []string {
	messages := make([]string, 0, len(violations))
	for _, violation := range violations {
		messages = append(messages, violation.Message)
	}
	return messages
}
