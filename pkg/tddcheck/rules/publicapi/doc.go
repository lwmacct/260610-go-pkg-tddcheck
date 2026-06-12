// Package publicapi checks public API names that use internal
// responsibility prefixes.
//
// It reports exported functions whose names start with Validate or Normalize.
// Those prefixes are reserved for internal validation and normalization helpers
// in this architecture policy.
package publicapi
