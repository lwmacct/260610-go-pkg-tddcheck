// Package tddcheck provides AST-based checks for enforcing test-driven
// development conventions in Go projects.
//
// The package is designed to be used from ordinary unit tests. It cannot prove
// that tests were written before implementation code, but it can enforce
// observable rules: exported API should have candidate tests and committed
// tests should not be empty.
//
// Typical usage:
//
//	func TestTDDPolicy(t *testing.T) {
//	    tddcheck.Assert(t)
//	}
package tddcheck
