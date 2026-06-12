package tddcheck

import "github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rulekit"

// Config describes project-specific architecture policy knobs.
//
// A zero Config uses DefaultConfig. Slice fields distinguish nil from empty:
// nil means use the default value, while an empty non-nil slice disables that
// portion of the policy.
type Config = rulekit.Config

// LayerDependencyRule rejects imports from SourceLayer to TargetLayer. When
// TargetRelPrefix is set, only matching target relative import paths are rejected.
type LayerDependencyRule = rulekit.LayerDependencyRule

// ValidationResolveConfig configures the accepted validation.go Resolve method
// signature used by protocol frameworks such as Huma.
type ValidationResolveConfig = rulekit.ValidationResolveConfig

// DatabaseTestConfig configures the optional database test boundary rule.
type DatabaseTestConfig = rulekit.DatabaseTestConfig

// DefaultConfig 返回默认架构策略。
func DefaultConfig() Config {
	return rulekit.DefaultConfig()
}
