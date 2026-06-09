package tddcheck

import "context"

// Rule checks a scanned Go project and returns findings.
type Rule interface {
	Name() string
	Check(context.Context, *Project) ([]Finding, error)
}

type ruleFunc struct {
	name  string
	check func(context.Context, *Project) ([]Finding, error)
}

func (r ruleFunc) Name() string {
	return r.name
}

func (r ruleFunc) Check(ctx context.Context, project *Project) ([]Finding, error) {
	return r.check(ctx, project)
}
