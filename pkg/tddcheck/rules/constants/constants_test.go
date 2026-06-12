package constants

import (
	"reflect"
	"testing"

	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/testkit"
)

func TestRulesBoundaryViolations(t *testing.T) {
	root := t.TempDir()

	testkit.WriteFile(t, root, "good/constants.go", `package good

const State = "ready"
`)
	testkit.WriteFile(t, root, "bad/constants.go", `package bad

const State = "ready"
var current = 1
type StateValue string
func StateName() string { return State }
`)
	testkit.WriteFile(t, root, "bad/service.go", `package bad

const serviceState = "bad"
`)

	violations, err := New(root).ConstantsBoundaryViolations()
	if err != nil {
		t.Fatal(err)
	}

	got := constantsViolationMessages(violations)
	want := []string{
		"constants.go must only declare const",
		"constants.go must only declare const",
		"constants.go must only declare const",
		"package-level const must be declared in constants.go",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %v, want %v", got, want)
	}
}

func TestRulesRequiresRoot(t *testing.T) {
	_, err := New("").ConstantsBoundaryViolations()
	if err == nil {
		t.Fatal("expected error for empty root")
	}
}

func constantsViolationMessages(violations []ConstantsBoundaryViolation) []string {
	messages := make([]string, 0, len(violations))
	for _, violation := range violations {
		messages = append(messages, violation.Message)
	}
	return messages
}
