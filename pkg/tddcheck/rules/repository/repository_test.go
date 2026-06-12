package repository

import (
	"reflect"
	"slices"
	"testing"

	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/testkit"
)

func TestRulesBoundaryViolations(t *testing.T) {
	root := t.TempDir()

	testkit.WriteFile(t, root, "good/repository.go", `package good

import "database/sql"

type CommandRepository interface { Save(Entity) error }
type QueryRepository interface { Get() (*Entity, error) }
type bunRepository struct {}

func newRepository() *bunRepository { return &bunRepository{} }
func (r *bunRepository) Save(input RequestDTO) (*Entity, error) { return nil, sql.ErrNoRows }
`)
	testkit.WriteFile(t, root, "bad/repository.go", `package bad

import (
	"net/http"
	"github.com/danielgtaylor/huma/v2"
	"github.com/coder/websocket"
)

const mode = "bad"
var enabled = true

type QueryRepository interface { Get() (*Entity, error) }
type User struct {}
type userRepository struct {}
type badRepository = userRepository

func newRepository() *userRepository { return &userRepository{} }
func (r *Repository) List() ([]UserDTO, error) { return nil, nil }
func (r *Repository) Get() (*UserDTO, error) { return nil, nil }
func ToUserDTO(user User) UserDTO { return UserDTO{} }
func helper() {}
func useImports(_ http.ResponseWriter, _ huma.Context, _ *websocket.Conn) {}
func (r *Repository) Sort(db DB) {
	db.NewSelect().Order("created_at DESC")
	db.NewSelect().Order(`+"`"+`"user".username ASC`+"`"+`)
	db.NewSelect().Order("user.username ASC")
	db.NewSelect().Order("lower(username) ASC")
}
`)
	testkit.WriteFile(t, root, "bad/other.go", `package bad

type OtherRepository interface {}
func newRepository() {}
func (r *Repository) Outside() {}
`)

	violations, err := New(root).RepositoryBoundaryViolations()
	if err != nil {
		t.Fatal(err)
	}

	got := repositoryViolationMessages(violations)
	want := []string{
		"mapper function ToUserDTO must be declared in mapper.go",
		"*Repository type OtherRepository must be declared in repository.go",
		"newRepository must be declared in repository.go",
		"repository receiver method Outside must be declared in repository.go",
		"repository function Get must not return DTO",
		"repository function List must not return DTO",
		"repository function ToUserDTO must not return DTO",
		"repository.go must only declare repository interfaces, implementation types, constructors, and repository methods",
		"repository.go must only declare repository interfaces, implementation types, constructors, and repository methods",
		"repository.go package-level function ToUserDTO must be newRepository",
		"repository.go package-level function helper must be newRepository",
		"repository.go package-level function useImports must be newRepository",
		"repository.go type User must be a *Repository interface or repository implementation type",
		"repository.go type badRepository must be a *Repository interface or repository implementation type",
		"repository.go must use OrderExpr or OrderBy for complex order expressions",
		"repository.go must use OrderExpr or OrderBy for complex order expressions",
		"repository.go must use OrderExpr or OrderBy for complex order expressions",
		"repository.go must not import github.com/coder/websocket",
		"repository.go must not import github.com/danielgtaylor/huma/v2",
		"repository.go must not import net/http",
	}
	slices.Sort(got)
	slices.Sort(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %v, want %v", got, want)
	}
}

func TestRulesRequiresRoot(t *testing.T) {
	_, err := New("").RepositoryBoundaryViolations()
	if err == nil {
		t.Fatal("expected error for empty root")
	}
}

func repositoryViolationMessages(violations []RepositoryBoundaryViolation) []string {
	messages := make([]string, 0, len(violations))
	for _, violation := range violations {
		messages = append(messages, violation.Message)
	}
	return messages
}
