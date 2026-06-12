// Package mapper checks boundaries for pure conversion functions.
//
// It reports forbidden imports in mapper.go, mapper functions with receivers,
// mapper.go functions that do not start with To, and mapper-like conversion
// functions declared outside mapper.go.
package mapper
