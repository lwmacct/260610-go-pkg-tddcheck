package tddcheck

import "testing"

func TestPolicyCheckPassesCleanProject(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "calc.go", `package calc

func Parse() {}
`)
	writeFile(t, root, "calc_test.go", `package calc

import "testing"

func TestParse(t *testing.T) {
	got := 1
	if got != 1 {
		t.Fatal(got)
	}
}
`)

	result := Policy{Root: root, Rules: DefaultRules()}.Check()
	if result.Err != nil {
		t.Fatal(result.Err)
	}
	if !result.Passed {
		t.Fatalf("expected pass, got %s", result.Text())
	}
}

func TestAssert(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "calc.go", "package calc\nfunc Parse() {}\n")
	writeFile(t, root, "calc_test.go", `package calc

import "testing"

func TestParse(t *testing.T) {
	t.Log("covered")
}
`)

	Assert(t, WithRoot(root))
}

func TestCheck(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "calc.go", "package calc\nfunc Parse() {}\n")
	writeFile(t, root, "calc_test.go", `package calc

import "testing"

func TestParse(t *testing.T) {
	t.Log("covered")
}
`)

	result := Check(WithRoot(root))
	if result.Err != nil {
		t.Fatal(result.Err)
	}
	if !result.Passed {
		t.Fatalf("expected pass, got %s", result.Text())
	}
}

func TestDefaultPolicy(t *testing.T) {
	policy := DefaultPolicy(WithRules())
	if len(policy.Rules) != 0 {
		t.Fatalf("expected rules override, got %d", len(policy.Rules))
	}
}

func TestPolicyAssert(t *testing.T) {
	TestPolicyCheckPassesCleanProject(t)
}

func TestWithRoot(t *testing.T) {
	policy := DefaultPolicy(WithRoot("/tmp/project"))
	if policy.Root != "/tmp/project" {
		t.Fatalf("unexpected root: %q", policy.Root)
	}
}

func TestWithChanged(t *testing.T) {
	policy := DefaultPolicy(WithChanged(true), WithRules())
	if !policy.Changed {
		t.Fatal("expected changed mode")
	}
	if len(policy.Rules) != 1 || policy.Rules[0].Name() != "changed-code-needs-test" {
		t.Fatalf("expected changed rule to be prepended, got %#v", policy.Rules)
	}
}

func TestWithRules(t *testing.T) {
	policy := DefaultPolicy(WithRules())
	if len(policy.Rules) != 0 {
		t.Fatalf("expected empty rules, got %d", len(policy.Rules))
	}
}

func TestWithDefaultRules(t *testing.T) {
	policy := DefaultPolicy(WithRules(), WithDefaultRules())
	if len(policy.Rules) != len(DefaultRules()) {
		t.Fatalf("expected default rules, got %d", len(policy.Rules))
	}
}

func TestWithIgnore(t *testing.T) {
	policy := DefaultPolicy(WithIgnore("gen/**"))
	if len(policy.Ignore) != 1 || policy.Ignore[0] != "gen/**" {
		t.Fatalf("unexpected ignore list: %#v", policy.Ignore)
	}
}

func TestWithCallerSkip(t *testing.T) {
	policy := DefaultPolicy(WithCallerSkip(7))
	if policy.CallerSkip != 7 {
		t.Fatalf("unexpected caller skip: %d", policy.CallerSkip)
	}
}

func TestDefaultRules(t *testing.T) {
	rules := DefaultRules()
	if len(rules) != 2 {
		t.Fatalf("expected two default rules, got %d", len(rules))
	}
}
