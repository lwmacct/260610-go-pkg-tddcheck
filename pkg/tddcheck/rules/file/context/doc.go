// Package context checks local boundaries around context helpers and
// context.WithValue.
//
// It reports context helper functions outside context.go, unrelated functions in
// context.go, and context.WithValue calls outside context.go. Context helper
// detection is based on names such as ContextWith*, *FromContext,
// *ContextFrom, and *Context with context.Context parameters.
package context
