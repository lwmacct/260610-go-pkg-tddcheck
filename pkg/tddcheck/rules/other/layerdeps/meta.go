package layerdeps

import "github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rulekit"

// Meta describes this rule.
var Meta = rulekit.RuleMeta{
	ID:   "dependency-layerdeps",
	Kind: rulekit.RuleKindDependency,
	Name: "layer dependency direction",
}
