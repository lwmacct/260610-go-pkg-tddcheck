package cqrs

import (
	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/testkit"
	"reflect"
	"testing"
)

func TestRulesStructSuffixViolations(t *testing.T) {
	root := t.TempDir()

	testkit.WriteFile(t, root, "user/cqrs.go", `package user

type CreateUserCommand struct {}
type UserQuery struct {}
type UserResult struct {}
type UserUseCase interface {}
type AuthenticateUserCommand = CreateUserCommand
`)
	testkit.WriteFile(t, root, "bad/cqrs.go", `package bad

type UserInput struct {}
type UserOutput struct {}
`)
	testkit.WriteFile(t, root, "bad/dto.go", `package bad

type Ignored struct {}
`)

	violations, err := New(root).StructSuffixViolations()
	if err != nil {
		t.Fatal(err)
	}

	got := cqrsViolationNames(violations)
	want := []string{"UserInput", "UserOutput"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %v, want %v", got, want)
	}
}

func TestRulesInterfaceNameViolations(t *testing.T) {
	root := t.TempDir()

	testkit.WriteFile(t, root, "user/cqrs.go", `package user

type UserUseCase interface {}
type CreateUserCommandHandler interface {}
type ListUsersQueryHandler interface {}
type OwnerResourceAccess interface {}
type AccessPolicy interface {}
type UserAuthorizer interface {}
`)
	testkit.WriteFile(t, root, "bad/cqrs.go", `package bad

type UserRepository interface {}
type UserService interface {}
type UserHandler interface {}
type UserProvider interface {}
type UserResolver interface {}
type UserInterface interface {}
type UserValidator interface {}
`)
	testkit.WriteFile(t, root, "bad/service.go", `package bad

type IgnoredRepository interface {}
`)

	violations, err := New(root).InterfaceNameViolations()
	if err != nil {
		t.Fatal(err)
	}

	got := cqrsInterfaceViolationNames(violations)
	want := []string{
		"UserRepository",
		"UserService",
		"UserHandler",
		"UserProvider",
		"UserResolver",
		"UserInterface",
		"UserValidator",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %v, want %v", got, want)
	}
}

func TestRulesRequiresRoot(t *testing.T) {
	_, err := New("").StructSuffixViolations()
	if err == nil {
		t.Fatal("expected error for empty root")
	}
}

func cqrsViolationNames(violations []CQRSSuffixViolation) []string {
	names := make([]string, 0, len(violations))
	for _, violation := range violations {
		names = append(names, violation.Name)
	}
	return names
}

func cqrsInterfaceViolationNames(violations []CQRSInterfaceNameViolation) []string {
	names := make([]string, 0, len(violations))
	for _, violation := range violations {
		names = append(names, violation.Name)
	}
	return names
}
