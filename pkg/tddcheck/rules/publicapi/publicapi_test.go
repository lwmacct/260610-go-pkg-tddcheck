package publicapi

import (
	"reflect"
	"testing"

	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/testkit"
)

func TestModulePublicAPIRuleViolations(t *testing.T) {
	root := t.TempDir()

	testkit.WriteFile(t, root, "good/service.go", `package good

func CheckName() error {
	return nil
}

func normalizeName(value string) string {
	return value
}

func (Service) Check() error {
	return nil
}

type Service struct {}
`)
	testkit.WriteFile(t, root, "bad/service.go", `package bad

func ValidateName() error {
	return nil
}

func NormalizeName(value string) string {
	return value
}

func (Service) Validate() error {
	return nil
}

func (Service) Normalize() error {
	return nil
}

type Service struct {}
`)
	testkit.WriteFile(t, root, "bad/service_test.go", `package bad

func ValidateTestName() error {
	return nil
}
`)

	violations, err := New(root).PublicAPINameViolations()
	if err != nil {
		t.Fatal(err)
	}

	got := publicAPIViolationMessages(violations)
	want := []string{
		"public API ValidateName must not use Validate or Normalize prefix",
		"public API NormalizeName must not use Validate or Normalize prefix",
		"public API Validate must not use Validate or Normalize prefix",
		"public API Normalize must not use Validate or Normalize prefix",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %v, want %v", got, want)
	}
}

func TestRulesRequiresRoot(t *testing.T) {
	_, err := New("").PublicAPINameViolations()
	if err == nil {
		t.Fatal("expected error for empty root")
	}
}

func publicAPIViolationMessages(violations []PublicAPINameViolation) []string {
	messages := make([]string, 0, len(violations))
	for _, violation := range violations {
		messages = append(messages, violation.Message)
	}
	return messages
}
