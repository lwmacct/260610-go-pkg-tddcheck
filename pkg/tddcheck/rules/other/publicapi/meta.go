package publicapi

import "github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rulekit"

// Meta describes this rule.
var Meta = rulekit.RuleMeta{
	ID:   "name-public-api",
	Kind: rulekit.RuleKindName,
	Name: "public API name",
}
