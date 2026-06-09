package tddcheck

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// Policy describes a project-level TDD convention check.
type Policy struct {
	Root       string
	Changed    bool
	Rules      []Rule
	Ignore     []string
	CallerSkip int
}

// Option configures a Policy.
type Option func(*Policy)

// Assert runs the default policy and fails t when findings are reported.
func Assert(t testing.TB, opts ...Option) {
	t.Helper()
	DefaultPolicy(opts...).Assert(t)
}

// Check runs the default policy and returns the result.
func Check(opts ...Option) Result {
	return DefaultPolicy(opts...).Check()
}

// DefaultPolicy returns the package default policy.
func DefaultPolicy(opts ...Option) Policy {
	policy := Policy{
		Rules: append([]Rule(nil), DefaultRules()...),
		Ignore: []string{
			"vendor/**",
			"testdata/**",
			"**/testdata/**",
		},
		CallerSkip: 2,
	}
	for _, opt := range opts {
		opt(&policy)
	}
	if policy.Changed {
		policy.Rules = append([]Rule{ChangedCodeHasTests()}, policy.Rules...)
	}

	return policy
}

// Assert runs the policy and fails t when findings are reported.
func (p Policy) Assert(t testing.TB) {
	t.Helper()
	result := p.Check()
	if result.Err != nil {
		t.Fatal(result.Err)
	}
	if !result.Passed {
		t.Fatal(result.Text())
	}
}

// Check scans the project and runs policy rules.
func (p Policy) Check() Result {
	start := time.Now()
	root, err := p.resolveRoot()
	if err != nil {
		return Result{Passed: false, Err: err, Duration: time.Since(start)}
	}

	project, err := Scan(context.Background(), ScanOptions{
		Root:        root,
		StagedOnly:  p.Changed,
		IgnoreGlobs: p.Ignore,
	})
	if err != nil {
		return Result{Passed: false, Err: err, Duration: time.Since(start)}
	}

	var findings []Finding
	for _, rule := range p.Rules {
		if rule == nil {
			continue
		}
		ruleFindings, err := rule.Check(context.Background(), project)
		if err != nil {
			return Result{
				Passed:   false,
				Err:      fmt.Errorf("%s: %w", rule.Name(), err),
				Findings: findings,
				Duration: time.Since(start),
			}
		}
		findings = append(findings, ruleFindings...)
	}

	return Result{
		Passed:   findingsPassed(findings),
		Findings: findings,
		Duration: time.Since(start),
	}
}

func (p Policy) resolveRoot() (string, error) {
	if p.Root != "" {
		return p.Root, nil
	}

	return findModuleRoot(p.CallerSkip + 1)
}

func findingsPassed(findings []Finding) bool {
	for _, finding := range findings {
		if finding.Severity == SeverityError {
			return false
		}
	}

	return true
}

// WithRoot sets the project root. Empty root means auto-detect from caller.
func WithRoot(root string) Option {
	return func(p *Policy) {
		p.Root = root
	}
}

// WithChanged enables git staged-change checks.
func WithChanged(enabled bool) Option {
	return func(p *Policy) {
		p.Changed = enabled
	}
}

// WithRules replaces the enabled rule set.
func WithRules(rules ...Rule) Option {
	return func(p *Policy) {
		p.Rules = append([]Rule(nil), rules...)
	}
}

// WithDefaultRules replaces the enabled rule set with the default unit-test
// friendly rules.
func WithDefaultRules() Option {
	return func(p *Policy) {
		p.Rules = append([]Rule(nil), DefaultRules()...)
	}
}

// WithIgnore replaces the ignore glob list.
func WithIgnore(patterns ...string) Option {
	return func(p *Policy) {
		p.Ignore = append([]string(nil), patterns...)
	}
}

// WithCallerSkip adjusts automatic module-root detection for wrappers.
func WithCallerSkip(skip int) Option {
	return func(p *Policy) {
		p.CallerSkip = skip
	}
}
