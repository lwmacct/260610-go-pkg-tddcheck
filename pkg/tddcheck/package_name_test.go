package tddcheck

import (
	"reflect"
	"testing"
)

func TestModulePackageNameRulesViolations(t *testing.T) {
	root := t.TempDir()

	writeFile(t, root, "domain/identityuser/service.go", `package user
`)
	writeFile(t, root, "domain/nodeaccesspolicy/service.go", `package nodeaccesspolicy
`)
	writeFile(t, root, "domain/nodeaccesspolicy/service_test.go", `package nodeaccesspolicy_test
`)

	violations, err := (ModulePackageNameRules{Root: root}).PackageNameViolations()
	if err != nil {
		t.Fatal(err)
	}

	got := packageNameViolationMessages(violations)
	want := []string{`package name must be "identityuser" or "identityuser_test", got "user"`}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %v, want %v", got, want)
	}
}

func TestModulePackageNameRulesRequiresRoot(t *testing.T) {
	_, err := (ModulePackageNameRules{}).PackageNameViolations()
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
