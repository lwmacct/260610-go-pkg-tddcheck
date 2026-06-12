// Package tddcheck provides Go test helpers for enforcing project architecture
// boundaries in layered Go modules.
//
// The package is intentionally mechanical: it scans Go source files with the
// standard parser and reports naming, file ownership, and dependency direction
// violations. It is designed to run from ordinary tests and CI.
//
// Typical usage:
//
//	func TestArchitecture(t *testing.T) {
//	    tddcheck.ProjectRules{Root: "internal"}.Assert(t)
//	}
//
// Project-specific policy can be customized through Config while preserving the
// same rule APIs.
package tddcheck
