package errorboundary

import (
	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/testkit"
	"reflect"
	"testing"
)

func TestModuleErrorRulesBoundaryViolations(t *testing.T) {
	root := t.TempDir()

	testkit.WriteFile(t, root, "good/errors.go", `package good

import "errors"

var ErrInvalid = errors.New("invalid")

type InvalidError struct {}

func (err InvalidError) Error() string {
	return "invalid"
}

func IsInvalid(err error) bool {
	return errors.Is(err, ErrInvalid)
}
`)
	testkit.WriteFile(t, root, "bad/errors.go", `package bad

import "errors"

const invalid = "invalid"
var invalid = errors.New("invalid")
type invalidError struct {}
type Invalid = invalidError

func invalid() {}
func (err invalidError) Message() string { return "" }
`)
	testkit.WriteFile(t, root, "bad/service.go", `package bad

import "errors"

var ErrService = errors.New("service")
`)

	violations, err := (ModuleErrorRules{Root: root}).ErrorsBoundaryViolations()
	if err != nil {
		t.Fatal(err)
	}

	got := errorsViolationMessages(violations)
	want := []string{
		"errors.go must only declare error vars, *Error types, and error helpers",
		"errors.go var invalid must be an Err* error value",
		"errors.go type Invalid must be a non-alias *Error type",
		"errors.go function invalid must use Is*, As*, or Wrap* name",
		"errors.go method Message must be Error, Is, As, or Unwrap on a local *Error type",
		"package-level error var must be declared in errors.go",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %v, want %v", got, want)
	}
}

func errorsViolationMessages(violations []ErrorsBoundaryViolation) []string {
	messages := make([]string, 0, len(violations))
	for _, violation := range violations {
		messages = append(messages, violation.Message)
	}
	return messages
}
