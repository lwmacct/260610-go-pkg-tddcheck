package mapper

import "github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rulekit"

// Meta describes this rule.
var Meta = rulekit.RuleMeta{
	ID:   "file-mapper",
	Kind: rulekit.RuleKindFile,
	Name: "mapper file ownership",
}
