package errorprefix

import (
	"reflect"
	"testing"

	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/testkit"
)

func TestRulesErrorPrefixViolations(t *testing.T) {
	root := t.TempDir()

	testkit.WriteFile(t, root, "user/errors.go", `package user

import "errors"

var ErrInvalid = errors.New("invalid")
var ErrExplicit error = errors.New("explicit")

type UserError struct {}

func (err UserError) Error() string { return "user" }
`)
	testkit.WriteFile(t, root, "bad/errors.go", `package bad

import (
	"errors"
	"fmt"
)

var invalid = errors.New("invalid")
var wrapped = fmt.Errorf("wrapped")
var explicit error
var ErrValid = errors.New("valid")
`)
	testkit.WriteFile(t, root, "bad/service.go", `package bad

import "errors"

var ignored = errors.New("ignored")
`)

	violations, err := New(root).ErrorPrefixViolations()
	if err != nil {
		t.Fatal(err)
	}

	got := errorViolationNames(violations)
	want := []string{"invalid", "wrapped", "explicit"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %v, want %v", got, want)
	}
}

func TestRulesRequiresRoot(t *testing.T) {
	_, err := New("").ErrorPrefixViolations()
	if err == nil {
		t.Fatal("expected error for empty root")
	}
}

func errorViolationNames(violations []ErrorPrefixViolation) []string {
	names := make([]string, 0, len(violations))
	for _, violation := range violations {
		names = append(names, violation.Name)
	}
	return names
}
