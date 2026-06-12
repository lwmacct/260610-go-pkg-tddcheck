package validation

import "github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rulekit"

// Meta describes this rule.
var Meta = rulekit.RuleMeta{
	ID:   "file-validation",
	Kind: rulekit.RuleKindFile,
	Name: "validation file ownership",
}
