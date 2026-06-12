// Package errorprefix checks Err-prefixed package error variables.
//
// It scans errors.go files and reports package-level error variables that do
// not start with Err. Error variables are detected from explicit error types or
// common constructors such as errors.New and fmt.Errorf.
package errorprefix
