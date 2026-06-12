package service

import "github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rulekit"

// Meta describes this rule.
var Meta = rulekit.RuleMeta{
	ID:   "file-service",
	Kind: rulekit.RuleKindFile,
	Name: "service file ownership",
}
