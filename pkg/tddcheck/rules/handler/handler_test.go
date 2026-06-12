package handler

import (
	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/testkit"
	"reflect"
	"slices"
	"testing"
)

func TestRulesBoundaryViolations(t *testing.T) {
	root := t.TempDir()

	testkit.WriteFile(t, root, "good/handler.go", `package good

import "database/sql"

type CurrentUserFunc func() string
type RouteDeps struct {}
type RouteConfig struct {}
type Handler struct {}
type UserService interface { User() string }
type userInput struct {}
type userResponse struct {}

func RegisterRoutes() error { return sql.ErrNoRows }
func (h *Handler) list() {}
`)
	testkit.WriteFile(t, root, "bad/handler.go", `package bad

import (
	"context"

	"github.com/uptrace/bun"
)

const mode = "bad"
var enabled = true

type Handler struct {}
type CurrentUser = string
type User struct {}
type UserLookup interface { User() string }
type BadBodyDTO struct {}
type bodyInput struct { Body string }
type itemResponse struct { Body string }

func list(db bun.IDB) {
	_ = db.NewSelect()
	_ = db.NewInsert()
}

func (h *Handler) list() {}
func (h *Handler) empty(ctx context.Context, input *struct{}) (*itemResponse, error) { return nil, nil }
func (h *Handler) required(ctx context.Context, input *userInput) (*itemResponse, error) {
	return nil, huma.Error400BadRequest("name required")
}
func (x *User) show() {}
`)
	testkit.WriteFile(t, root, "bad/handler.extra.go", `package bad

type extraInput struct {}

func update(db interface{ RunInTx() error }) {
	_ = db.RunInTx()
}

func (h *Handler) update() {}
`)
	testkit.WriteFile(t, root, "bad/mapper.go", `package bad

func (h *Handler) mapUser() {}
`)
	testkit.WriteFile(t, root, "bad/handler_test.go", `package bad

import "github.com/uptrace/bun"

func seed(db bun.IDB) {
	_ = db.NewInsert()
}
`)

	violations, err := New(root).HandlerBoundaryViolations()
	if err != nil {
		t.Fatal(err)
	}

	got := handlerViolationMessages(violations)
	want := []string{
		"handler file must not call NewInsert",
		"handler file must not call NewSelect",
		"handler file must not call RunInTx",
		"handler file must not import github.com/uptrace/bun",
		"handler file must use Huma validation tags or Resolver for required request fields",
		"handler file receiver method show must use Handler receiver",
		"handler.go must not declare const or var",
		"handler.go must not declare const or var",
		"handler.go must only declare RegisterRoutes and Handler receiver methods",
		"handler.go type CurrentUser must be Handler, RouteDeps, RouteConfig, *Func, dependency interface, or protocol input/output type",
		"handler.go type BadBodyDTO must be Handler, RouteDeps, RouteConfig, *Func, dependency interface, or protocol input/output type",
		"handler.go type User must be Handler, RouteDeps, RouteConfig, *Func, dependency interface, or protocol input/output type",
		"handler.go type UserLookup must be Handler, RouteDeps, RouteConfig, *Func, dependency interface, or protocol input/output type",
		"handler.go body-only input type bodyInput must use httpapi.BodyInput[T]",
		"handler.go Body field type string must use DTO suffix",
		"handler.go Body field type string must use DTO suffix",
		"handler.go response type itemResponse must use httpapi.Body[T] unless it declares header or status fields",
		"handler.*.go must only declare Handler receiver methods",
		"handler.*.go must only declare Handler receiver methods",
		"Handler receiver method mapUser must be declared in handler.go or handler.*.go",
		"Huma handler empty must use httpapi.EmptyInput instead of *struct{}",
	}
	slices.Sort(got)
	slices.Sort(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %v, want %v", got, want)
	}
}

func TestRulesRequiresRoot(t *testing.T) {
	_, err := New("").HandlerBoundaryViolations()
	if err == nil {
		t.Fatal("expected error for empty root")
	}
}

func handlerViolationMessages(violations []HandlerBoundaryViolation) []string {
	messages := make([]string, 0, len(violations))
	for _, violation := range violations {
		messages = append(messages, violation.Message)
	}
	return messages
}
