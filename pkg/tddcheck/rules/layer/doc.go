// Package layer checks import direction between configured architecture
// layers.
//
// It reports imports that violate configured layer dependency rules. By default
// it checks domain, usecase, adapter, runtime, and infra directories and
// prevents inner or infrastructure layers from importing forbidden targets.
package layer
