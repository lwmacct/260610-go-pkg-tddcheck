package errors

import "github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rulekit"

// Meta describes this rule.
var Meta = rulekit.RuleMeta{
	ID:   "file-errors",
	Kind: rulekit.RuleKindFile,
	Name: "errors file ownership",
}
