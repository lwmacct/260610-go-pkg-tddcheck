package tddcheck

import (
	"reflect"
	"testing"
)

func TestModuleValidationRulesBoundaryViolations(t *testing.T) {
	root := t.TempDir()

	writeFile(t, root, "good/validation.go", `package good

import "github.com/danielgtaylor/huma/v2"

const minLength = 1
var domainPattern = "x"

func validateName(value string) error {
	return nil
}

func normalizeName(value string) string {
	return value
}

func (*requestBody) Resolve(_ huma.Context, _ *huma.PathBuffer) []error {
	return nil
}
`)
	writeFile(t, root, "bad/validation.go", `package bad

type Validator struct {}
const MaxLength = 1
var DomainPattern = "x"

func ValidateName(value string) error {
	return nil
}

func validName(value string) bool {
	return true
}

func normalize(value string) string {
	return value
}

func (Validator) Resolve() {}
func (Validator) validateMethod() {}
`)
	writeFile(t, root, "bad/service.go", `package bad

func validateServiceInput() error {
	return nil
}

func normalizeServiceInput() string {
	return ""
}

func ValidateServiceInput() error {
	return nil
}
`)
	writeFile(t, root, "bad/service_test.go", `package bad

func validateTestInput() error {
	return nil
}
`)

	violations, err := (ModuleValidationRules{Root: root}).ValidationBoundaryViolations()
	if err != nil {
		t.Fatal(err)
	}

	got := validationViolationMessages(violations)
	want := []string{
		"validate*/normalize* function validateServiceInput must be declared in validation.go",
		"validate*/normalize* function normalizeServiceInput must be declared in validation.go",
		"validate*/normalize* function ValidateServiceInput must be declared in validation.go",
		"validation.go must not declare type",
		"validation.go const MaxLength must be private",
		"validation.go var DomainPattern must be private",
		"validation.go function ValidateName must start with validate or normalize",
		"validation.go function validName must start with validate or normalize",
		"validation.go function normalize must start with validate or normalize",
		"validation.go Resolve receiver method must implement huma.Resolver or huma.ResolverWithPath",
		"validation.go must not declare receiver method validateMethod",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %v, want %v", got, want)
	}
}

func TestModuleValidationRulesRequiresRoot(t *testing.T) {
	_, err := (ModuleValidationRules{}).ValidationBoundaryViolations()
	if err == nil {
		t.Fatal("expected error for empty root")
	}
}

func validationViolationMessages(violations []ValidationBoundaryViolation) []string {
	messages := make([]string, 0, len(violations))
	for _, violation := range violations {
		messages = append(messages, violation.Message)
	}
	return messages
}
