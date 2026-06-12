// Package repository checks persistence boundary rules.
//
// It reports repository types, constructors, and receiver methods outside
// repository.go, unrelated declarations inside repository.go, forbidden protocol
// imports, mapper functions in repository.go, repository functions returning
// DTOs, and unsafe Bun Order calls that should use OrderExpr or OrderBy.
package repository
