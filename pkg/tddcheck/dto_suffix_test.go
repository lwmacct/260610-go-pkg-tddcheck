package tddcheck

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestModuleDTORulesStructSuffixViolations(t *testing.T) {
	root := t.TempDir()

	writeFile(t, root, "user/dto.go", `package user

type UserDTO struct {}
type UserListDTOs struct {}
type userAlias = UserDTO
type userInterface interface {}
`)
	writeFile(t, root, "bad/dto.go", `package bad

type User struct {}
type userPayload struct {}
`)
	writeFile(t, root, "bad/handler.go", `package bad

type ignoredHandlerInput struct {}
`)

	violations, err := (ModuleDTORules{Root: root}).StructSuffixViolations()
	if err != nil {
		t.Fatal(err)
	}

	got := violationNames(violations)
	want := []string{"User", "userPayload"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %v, want %v", got, want)
	}
}

func TestModuleDTORulesFuncViolations(t *testing.T) {
	root := t.TempDir()

	writeFile(t, root, "user/dto.go", `package user

type UserDTO struct {}
`)
	writeFile(t, root, "bad/dto.go", `package bad

type UserDTO struct {}

func ParseUserDTO() UserDTO {
	return UserDTO{}
}

func (dto UserDTO) Validate() error {
	return nil
}
`)
	writeFile(t, root, "bad/service.go", `package bad

func ignoredServiceFunc() {}
`)

	violations, err := (ModuleDTORules{Root: root}).FuncViolations()
	if err != nil {
		t.Fatal(err)
	}

	got := dtoFuncViolationNames(violations)
	want := []string{"ParseUserDTO", "UserDTO.Validate"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %v, want %v", got, want)
	}
}

func TestModuleDTORulesFileOwnershipViolations(t *testing.T) {
	root := t.TempDir()

	writeFile(t, root, "user/dto.go", `package user

type UserDTO struct {}
`)
	writeFile(t, root, "bad/handler.go", `package bad

type CreateUserDTO struct {}
type createUserDTO struct {}
type ignoredDTO = CreateUserDTO
`)

	violations, err := (ModuleDTORules{Root: root}).FileOwnershipViolations()
	if err != nil {
		t.Fatal(err)
	}

	got := dtoFileViolationNames(violations)
	want := []string{"CreateUserDTO", "createUserDTO"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %v, want %v", got, want)
	}
}

func TestModuleDTORulesRequiresRoot(t *testing.T) {
	_, err := (ModuleDTORules{}).StructSuffixViolations()
	if err == nil {
		t.Fatal("expected error for empty root")
	}
}

func writeFile(t *testing.T, root string, name string, content string) {
	t.Helper()

	filename := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(filename), 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filename, []byte(content), 0600); err != nil {
		t.Fatal(err)
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
