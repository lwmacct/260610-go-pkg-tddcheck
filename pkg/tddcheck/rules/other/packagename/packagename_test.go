package packagename

import (
	"reflect"
	"testing"

	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/testkit"
)

func TestRulesViolations(t *testing.T) {
	root := t.TempDir()

	testkit.WriteFile(t, root, "domain/identityuser/service.go", `package user
`)
	testkit.WriteFile(t, root, "domain/nodeaccesspolicy/service.go", `package nodeaccesspolicy
`)
	testkit.WriteFile(t, root, "domain/nodeaccesspolicy/service_test.go", `package nodeaccesspolicy_test
`)

	violations, err := New(root).PackageNameViolations()
	if err != nil {
		t.Fatal(err)
	}

	got := packageNameViolationMessages(violations)
	want := []string{`package name must be "identityuser" or "identityuser_test", got "user"`}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %v, want %v", got, want)
	}
}

func TestRulesRequiresRoot(t *testing.T) {
	_, err := New("").PackageNameViolations()
	if err == nil {
		t.Fatal("expected error for empty root")
	}
}

func packageNameViolationMessages(violations []PackageNameViolation) []string {
	messages := make([]string, 0, len(violations))
	for _, violation := range violations {
		messages = append(messages, violation.Message)
	}
	return messages
}
