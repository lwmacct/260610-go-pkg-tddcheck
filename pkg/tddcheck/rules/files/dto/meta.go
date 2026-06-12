package dto

import "github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rulekit"

// Meta describes this rule.
var Meta = rulekit.RuleMeta{
	ID:   "file-dto",
	Kind: rulekit.RuleKindFile,
	Name: "DTO file ownership",
}
