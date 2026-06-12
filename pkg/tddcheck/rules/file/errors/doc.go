// Package errors checks ownership and shape of package error APIs.
//
// It reports package-level error vars outside errors.go, non-error declarations
// in errors.go, error types that are not non-alias *Error types, and functions
// or methods in errors.go that are not accepted error helpers.
package errors
