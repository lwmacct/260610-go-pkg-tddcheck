// Package tddcheck provides AST-based checks for enforcing test-driven
// development conventions in Go projects.
//
// The package is designed for pre-commit and CI usage. It cannot prove that
// tests were written before implementation code, but it can enforce observable
// rules: changed production code should be accompanied by tests, exported API
// should have candidate tests, and committed tests should not be empty or
// skipped.
//
// Typical usage:
//
//	result, err := tddcheck.Check(ctx,
//	    tddcheck.WithRoot("."),
//	    tddcheck.WithStagedOnly(true),
//	    tddcheck.WithDefaultRules(),
//	)
//	if err != nil {
//	    return err
//	}
//	if !result.Passed {
//	    fmt.Println(result.Text())
//	}
package tddcheck
