package handler

import "github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rulekit"

// Meta describes this rule.
var Meta = rulekit.RuleMeta{
	ID:   "file-handler",
	Kind: rulekit.RuleKindFile,
	Name: "handler file boundary",
}
