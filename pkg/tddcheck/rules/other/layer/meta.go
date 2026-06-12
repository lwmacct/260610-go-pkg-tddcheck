package layer

import "github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rulekit"

// Meta describes this rule.
var Meta = rulekit.RuleMeta{
	ID:   "dependency-layer",
	Kind: rulekit.RuleKindDependency,
	Name: "layer dependency direction",
}
