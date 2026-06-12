// Package constants checks package-level constant ownership in layered Go
// modules.
//
// It reports const declarations outside constants.go. It also reports
// constants.go declarations that are not imports or const declarations.
package constants
