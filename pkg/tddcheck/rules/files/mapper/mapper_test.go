package mapper

import (
	"reflect"
	"slices"
	"testing"

	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/testkit"
)

func TestRulesBoundaryViolations(t *testing.T) {
	root := t.TempDir()

	testkit.WriteFile(t, root, "good/mapper.go", `package good

func ToUserDTO(user User) UserDTO { return UserDTO{} }
func ToUserSchema(user User) userSchema { return userSchema{} }
`)
	testkit.WriteFile(t, root, "bad/mapper.go", `package bad

import (
	"context"
	"github.com/uptrace/bun"
)

func (h *Handler) ToUserDTO(user User) UserDTO { return UserDTO{} }
func buildUserDTO(user User) UserDTO { return UserDTO{} }
`)
	testkit.WriteFile(t, root, "bad/handler.go", `package bad

func ToUserDTO(user User) UserDTO { return UserDTO{} }
func ToUserSchema(user User) userSchema { return userSchema{} }
func utilUserDTO(user User) UserDTO { return UserDTO{} }
func utilUserDTOs(user User) []other.UserDTO { return nil }
func utilUserSchema(user User) *userSchema { return nil }
func buildUser(dto UserDTO) User { return User{} }
func parseUser(dto UserDTO) (User, error) { return User{}, nil }
func requestFromContext(ctx context.Context) (UserDTO, bool) { return UserDTO{}, false }
`)
	testkit.WriteFile(t, root, "bad/service_test.go", `package bad

func ToIgnoredDTO(user User) UserDTO { return UserDTO{} }
`)

	violations, err := New(root).MapperBoundaryViolations()
	if err != nil {
		t.Fatal(err)
	}

	got := violationMessages(violations)
	want := []string{
		"mapper.go must not import context",
		"mapper.go must not import github.com/uptrace/bun",
		"mapper function ToUserDTO must not use a receiver",
		"mapper function buildUserDTO must start with To",
		"mapper function ToUserDTO must be declared in mapper.go",
		"mapper function ToUserSchema must be declared in mapper.go",
		"mapper function utilUserDTO must be declared in mapper.go",
		"mapper function utilUserDTOs must be declared in mapper.go",
		"mapper function utilUserSchema must be declared in mapper.go",
		"mapper function buildUser must be declared in mapper.go",
	}
	slices.Sort(got)
	slices.Sort(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %v, want %v", got, want)
	}
}

func TestRulesRequiresRoot(t *testing.T) {
	_, err := New("").MapperBoundaryViolations()
	if err == nil {
		t.Fatal("expected error for empty root")
	}
}

func violationMessages(violations []MapperBoundaryViolation) []string {
	messages := make([]string, 0, len(violations))
	for _, violation := range violations {
		messages = append(messages, violation.Message)
	}
	return messages
}
