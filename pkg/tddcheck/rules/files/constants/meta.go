package constants

import "github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rulekit"

// Meta describes this rule.
var Meta = rulekit.RuleMeta{
	ID:   "file-constants",
	Kind: rulekit.RuleKindFile,
	Name: "constants file ownership",
}
