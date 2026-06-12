// Package database checks optional database test helper boundaries.
//
// It scans _test.go files and reports project-specific SQLite setup patterns
// that should use the shared database test helper. Allowed paths and detection
// strings are configurable through rulekit.Config.
package database
