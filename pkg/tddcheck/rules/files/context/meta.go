package context

import "github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rulekit"

// Meta describes this rule.
var Meta = rulekit.RuleMeta{
	ID:   "file-context",
	Kind: rulekit.RuleKindFile,
	Name: "context helper file ownership",
}
