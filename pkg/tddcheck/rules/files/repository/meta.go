package repository

import "github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rulekit"

// Meta describes this rule.
var Meta = rulekit.RuleMeta{
	ID:   "file-repository",
	Kind: rulekit.RuleKindFile,
	Name: "repository file boundary",
}
