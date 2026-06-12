// Package entity checks ownership of concrete entity and value object types.
//
// It reports non-concrete or alias types in entity.go, package-level functions
// in entity.go, entity methods whose receiver type is not declared in entity.go,
// and entity receiver methods declared outside entity.go.
package entity
