// Package service checks service file ownership.
//
// It reports Service and NewService declarations outside service.go, Service
// receiver methods outside service.go or service.*.go, unrelated declarations
// in service files, and service.*.go declarations that are not Service receiver
// methods.
package service
