package entity

import "github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rulekit"

// Meta describes this rule.
var Meta = rulekit.RuleMeta{
	ID:   "file-entity",
	Kind: rulekit.RuleKindFile,
	Name: "entity file ownership",
}
