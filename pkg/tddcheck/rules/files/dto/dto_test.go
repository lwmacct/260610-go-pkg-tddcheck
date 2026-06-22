package dto

import (
	"reflect"
	"testing"

	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/testkit"
)

func TestRulesStructSuffixViolations(t *testing.T) {
	root := t.TempDir()

	testkit.WriteFile(t, root, "user/dto.go", `package user

type UserDTO struct {}
type UserListDTOs struct {}
type userAlias = UserDTO
type userInterface interface {}
`)
	testkit.WriteFile(t, root, "bad/dto.go", `package bad

type User struct {}
type userPayload struct {}
`)
	testkit.WriteFile(t, root, "bad/handler.go", `package bad

type ignoredHandlerInput struct {}
`)

	violations, err := New(root).StructSuffixViolations()
	if err != nil {
		t.Fatal(err)
	}

	got := violationNames(violations)
	want := []string{"User", "userPayload"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %v, want %v", got, want)
	}
}

func TestRulesFuncViolations(t *testing.T) {
	root := t.TempDir()

	testkit.WriteFile(t, root, "user/dto.go", `package user

type UserDTO struct {}
`)
	testkit.WriteFile(t, root, "bad/dto.go", `package bad

type UserDTO struct {}

func ParseUserDTO() UserDTO {
	return UserDTO{}
}

func (dto UserDTO) Validate() error {
	return nil
}
`)
	testkit.WriteFile(t, root, "bad/service.go", `package bad

func ignoredServiceFunc() {}
`)

	violations, err := New(root).FuncViolations()
	if err != nil {
		t.Fatal(err)
	}

	got := dtoFuncViolationNames(violations)
	want := []string{"ParseUserDTO", "UserDTO.Validate"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %v, want %v", got, want)
	}
}

func TestRulesFileOwnershipViolations(t *testing.T) {
	root := t.TempDir()

	testkit.WriteFile(t, root, "user/dto.go", `package user

type UserDTO struct {}
`)
	testkit.WriteFile(t, root, "bad/handler.go", `package bad

type CreateUserDTO struct {}
type createUserDTO struct {}
type ignoredDTO = CreateUserDTO
`)

	violations, err := New(root).FileOwnershipViolations()
	if err != nil {
		t.Fatal(err)
	}

	got := dtoFileViolationNames(violations)
	want := []string{"CreateUserDTO", "createUserDTO"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %v, want %v", got, want)
	}
}

func TestRulesRequestsFileViolations(t *testing.T) {
	root := t.TempDir()

	testkit.WriteFile(t, root, "user/requests.go", `package user

type CreateUserRequest struct {}
`)
	testkit.WriteFile(t, root, "user/dto.go", `package user

type CreateUserDTO struct {}
`)

	violations, err := New(root).RequestsFileViolations()
	if err != nil {
		t.Fatal(err)
	}

	if len(violations) != 1 {
		t.Fatalf("violations len = %d, want 1", len(violations))
	}
}

func TestRulesProtocolTagViolations(t *testing.T) {
	root := t.TempDir()

	testkit.WriteFile(t, root, "user/dto.go", `package user

type CreateUserDTO struct {
	Name string `+"`json:\"name\"`"+`
}
`)
	testkit.WriteFile(t, root, "bad/service.go", `package bad

type CreateUserRequest struct {
	Name string `+"`json:\"name\"`"+`
}

type QueryInput struct {
	ID string `+"`query:\"id\"`"+`
}

type DomainCommand struct {
	Name string
}
`)

	violations, err := New(root).ProtocolTagViolations()
	if err != nil {
		t.Fatal(err)
	}

	got := protocolTagViolationNames(violations)
	want := []string{"CreateUserRequest", "QueryInput"}
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

func violationNames(violations []StructSuffixViolation) []string {
	names := make([]string, 0, len(violations))
	for _, violation := range violations {
		names = append(names, violation.Name)
	}
	return names
}

func dtoFuncViolationNames(violations []DTOFuncViolation) []string {
	names := make([]string, 0, len(violations))
	for _, violation := range violations {
		names = append(names, violation.Name)
	}
	return names
}

func dtoFileViolationNames(violations []DTOFileViolation) []string {
	names := make([]string, 0, len(violations))
	for _, violation := range violations {
		names = append(names, violation.Name)
	}
	return names
}

func protocolTagViolationNames(violations []ProtocolTagViolation) []string {
	names := make([]string, 0, len(violations))
	for _, violation := range violations {
		names = append(names, violation.Name)
	}
	return names
}
