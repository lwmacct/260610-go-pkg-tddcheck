package tddcheck

import (
	"strings"
	"testing"
)

func TestProjectRulesCheckPassesCleanProject(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "domain/identityuser/entity.go", `package identityuser

type User struct {}
`)

	result := (ProjectRules{Root: root}).Check()
	if result.Err != nil {
		t.Fatal(result.Err)
	}
	if !result.Passed {
		t.Fatalf("expected pass, got %s", result.Text())
	}
}

func TestProjectRulesCheckReportsViolations(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "domain/identityuser/service.go", `package user

const state = "bad"
`)

	result := (ProjectRules{Root: root}).Check()
	if result.Err != nil {
		t.Fatal(result.Err)
	}
	if result.Passed {
		t.Fatal("expected violations")
	}
	text := result.Text()
	for _, want := range []string{"tddcheck: failed", "[constants]", "[package-name]"} {
		if !strings.Contains(text, want) {
			t.Fatalf("result text missing %q:\n%s", want, text)
		}
	}
}

func TestProjectRulesAssert(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "domain/identityuser/entity.go", `package identityuser

type User struct {}
`)

	(ProjectRules{Root: root}).Assert(t)
}
