package utils

import (
	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/testkit"
	"reflect"
	"testing"
)

func TestRulesBoundaryViolations(t *testing.T) {
	root := t.TempDir()

	testkit.WriteFile(t, root, "good/utils.go", `package good

import "strings"

func utilParse(value string) string {
	return strings.TrimSpace(value)
}
`)
	testkit.WriteFile(t, root, "bad/utils.go", `package bad

type helper struct {}
const helperValue = "x"
var helperState = 1

func Parse(value string) string {
	return value
}

func util(value string) string {
	return value
}

func (helper) utilMethod() {}
`)
	testkit.WriteFile(t, root, "bad/service.go", `package bad

func utilServiceHelper() {}
`)
	testkit.WriteFile(t, root, "bad/service_test.go", `package bad

func utilTestHelper() {}
`)

	violations, err := New(root).UtilsBoundaryViolations()
	if err != nil {
		t.Fatal(err)
	}

	got := utilsViolationMessages(violations)
	want := []string{
		"util* function utilServiceHelper must be declared in utils.go",
		"utils.go must only declare private package-level util* functions",
		"utils.go must only declare private package-level util* functions",
		"utils.go must only declare private package-level util* functions",
		"utils.go function Parse must use util* prefix",
		"utils.go function util must use util* prefix",
		"utils.go must not declare receiver method utilMethod",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %v, want %v", got, want)
	}
}

func TestRulesRequiresRoot(t *testing.T) {
	_, err := New("").UtilsBoundaryViolations()
	if err == nil {
		t.Fatal("expected error for empty root")
	}
}

func utilsViolationMessages(violations []UtilsBoundaryViolation) []string {
	messages := make([]string, 0, len(violations))
	for _, violation := range violations {
		messages = append(messages, violation.Message)
	}
	return messages
}
