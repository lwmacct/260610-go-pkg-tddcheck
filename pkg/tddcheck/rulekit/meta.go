package rulekit

// RuleKind classifies an architecture rule by the kind of boundary it checks.
type RuleKind string

const (
	RuleKindFile       RuleKind = "file"
	RuleKindName       RuleKind = "name"
	RuleKindDependency RuleKind = "dependency"
	RuleKindTest       RuleKind = "test"
)

// RuleMeta describes a rule package.
type RuleMeta struct {
	ID   string
	Kind RuleKind
	Name string
}
