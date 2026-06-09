package tddcheck

import (
	"context"
	"fmt"
	"time"
)

// Check scans a Go project and runs the configured rules.
func Check(ctx context.Context, opts ...Option) (*Result, error) {
	start := time.Now()
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	project, err := Scan(ctx, ScanOptions{
		Root:        cfg.root,
		StagedOnly:  cfg.stagedOnly,
		IgnoreGlobs: cfg.ignoreGlobs,
	})
	if err != nil {
		return nil, err
	}

	var findings []Finding
	for _, rule := range cfg.rules {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if rule == nil {
			continue
		}
		ruleFindings, err := rule.Check(ctx, project)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", rule.Name(), err)
		}
		findings = append(findings, ruleFindings...)
	}

	result := &Result{
		Passed:   findingsPassed(findings),
		Findings: findings,
		Duration: time.Since(start),
	}

	return result, nil
}

func findingsPassed(findings []Finding) bool {
	for _, finding := range findings {
		if finding.Severity == SeverityError {
			return false
		}
	}

	return true
}
