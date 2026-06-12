package schema

import "github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rulekit"

// Meta describes this rule.
var Meta = rulekit.RuleMeta{
	ID:   "file-schema",
	Kind: rulekit.RuleKindFile,
	Name: "schema file ownership",
}
