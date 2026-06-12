// Package handler checks protocol handler boundaries.
//
// It reports persistence imports or calls in handler files, Handler receiver
// methods outside handler.go or handler.*.go, unrelated declarations in handler
// files, non-DTO HTTP body schema types, body-only wrapper structs that should
// use httpapi helpers, required-field Huma bad requests that should use
// validation tags or resolvers, and Huma handlers using *struct{} input.
package handler
