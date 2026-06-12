package tddcheck

import (
	"reflect"
	"testing"
)

func TestModuleContextRulesBoundaryViolations(t *testing.T) {
	root := t.TempDir()

	writeFile(t, root, "good/context.go", `package good

import "context"

type key struct{}
type wrappedContext struct{}

func ContextWithUser(ctx context.Context, user string) context.Context {
	return context.WithValue(ctx, key{}, user)
}

func UserFromContext(ctx context.Context) (string, bool) {
	user, ok := ctx.Value(key{}).(string)
	return user, ok
}

func commandContextFrom(ctx context.Context) context.Context {
	return ctx
}

func requestContext(ctx context.Context) context.Context {
	return ctx
}

func (ctx wrappedContext) Deadline() {}
`)
	writeFile(t, root, "bad/context.go", `package bad

func unrelated() {}
`)
	writeFile(t, root, "bad/service.go", `package bad

import "context"

type key struct{}

func ContextWithUser(ctx context.Context, user string) context.Context {
	return context.WithValue(ctx, key{}, user)
}

func UserFromContext(ctx context.Context) (string, bool) {
	user, ok := ctx.Value(key{}).(string)
	return user, ok
}

func requestContext(ctx context.Context) context.Context {
	return ctx
}
`)
	writeFile(t, root, "bad/handler.go", `package bad

import "context"

type key struct{}

func setValue(ctx context.Context, user string) context.Context {
	return context.WithValue(ctx, key{}, user)
}

func nodeAccessDecisionFromContext(ctx context.Context, node string) bool {
	return node != ""
}
`)
	writeFile(t, root, "bad/service_test.go", `package bad

import "context"

type key struct{}

func ContextWithTest(ctx context.Context, user string) context.Context {
	return context.WithValue(ctx, key{}, user)
}
`)

	violations, err := (ModuleContextRules{Root: root}).ContextBoundaryViolations()
	if err != nil {
		t.Fatal(err)
	}

	got := contextViolationMessages(violations)
	want := []string{
		"context.go function unrelated must be a context helper or local context type method",
		"context.WithValue must be used in context.go",
		"context helper ContextWithUser must be declared in context.go",
		"context helper UserFromContext must be declared in context.go",
		"context helper requestContext must be declared in context.go",
		"context.WithValue must be used in context.go",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %v, want %v", got, want)
	}
}

func TestModuleContextRulesRequiresRoot(t *testing.T) {
	_, err := (ModuleContextRules{}).ContextBoundaryViolations()
	if err == nil {
		t.Fatal("expected error for empty root")
	}
}

func contextViolationMessages(violations []ContextBoundaryViolation) []string {
	messages := make([]string, 0, len(violations))
	for _, violation := range violations {
		messages = append(messages, violation.Message)
	}
	return messages
}
