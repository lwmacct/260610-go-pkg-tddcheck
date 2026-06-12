// Package validation checks validation helper ownership and naming.
//
// It reports validation helpers outside validation.go, exported vars or consts
// in validation.go, type declarations in validation.go, functions that do not
// start with validate or normalize, and receiver methods other than configured
// Resolve signatures.
package validation
