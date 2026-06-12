# tddcheck

[![License](https://img.shields.io/github/license/lwmacct/260610-go-pkg-tddcheck)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/lwmacct/260610-go-pkg-tddcheck.svg)](https://pkg.go.dev/github.com/lwmacct/260610-go-pkg-tddcheck)
[![Go CI](https://github.com/lwmacct/260610-go-pkg-tddcheck/actions/workflows/go-ci.yml/badge.svg)](https://github.com/lwmacct/260610-go-pkg-tddcheck/actions/workflows/go-ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/lwmacct/260610-go-pkg-tddcheck)](https://goreportcard.com/report/github.com/lwmacct/260610-go-pkg-tddcheck)
[![GitHub Tag](https://img.shields.io/github/v/tag/lwmacct/260610-go-pkg-tddcheck?sort=semver)](https://github.com/lwmacct/260610-go-pkg-tddcheck/tags)

`tddcheck` is a Go test helper package for enforcing mechanical architecture
boundaries in layered Go projects. It checks file ownership, naming conventions,
and import direction with the standard Go parser.

This version is intentionally not compatible with the old public API test
checker. The package is now focused on project architecture rules.

## Install

```bash
go get github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck
```

## Go Test Usage

Create a normal Go test in the project that owns the architecture policy:

```go
package project_test

import (
    "testing"

    "github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck"
)

func TestArchitecture(t *testing.T) {
    tddcheck.ProjectRules{Root: "internal"}.Assert(t)
}
```

`Root` may be absolute or relative to the nearest `go.mod`. If omitted, the
default root is `internal`.

For fine-grained checks, use the individual rule types:

```go
func TestLayerDependencies(t *testing.T) {
    tddcheck.ModuleLayerRules{Root: "internal"}.AssertLayerDependencies(t)
}

func TestServiceBoundaries(t *testing.T) {
    tddcheck.ModuleServiceRules{Root: "internal"}.AssertServiceBoundaries(t)
}
```

## Default Project Rules

`ProjectRules` runs these checks:

- `layer`: `domain`, `usecase`, `adapter`, `runtime`, and `infra` imports obey dependency direction.
- `package-name`: package names match their directory names.
- `constants`: package constants live in `constants.go`.
- `entity`: concrete entity and value object types live in `entity.go`.
- `errors` and `error-prefix`: package errors live in `errors.go` and use `Err*` names.
- `context`: context helpers and `context.WithValue` usage live in `context.go`.
- `cqrs`: CQRS structs and interfaces use explicit suffixes.
- `dto`: DTO structs live in `dto.go`, use `DTO` or `DTOs`, and `dto.go` has no functions.
- `mapper`: mapper functions are pure `To*` conversions in `mapper.go`.
- `public-api`: exported APIs avoid internal responsibility prefixes such as `Validate*` and `Normalize*`.
- `service`: `Service`, `NewService`, and `Service` methods stay in service files.
- `validation`: validation helpers live in `validation.go` and use private `validate*` or `normalize*` names.
- `handler`: protocol handlers do not carry persistence responsibilities.
- `repository`: repositories do not carry protocol or DTO mapping responsibilities.
- `schema`: schema files are reserved for storage schema definitions.
- `utils`: private `util*` helpers live in `utils.go`.

Database test boundary checks are opt-in:

```go
func TestArchitecture(t *testing.T) {
    tddcheck.ProjectRules{
        Root:                 ".",
        IncludeDatabaseTests: true,
    }.Assert(t)
}
```

## CLI

```bash
go install github.com/lwmacct/260610-go-pkg-tddcheck/cmd/tddcheck@latest
tddcheck --root internal
tddcheck --root . --database-tests
```

The command exits with code `1` when violations are found and `2` for execution
errors.

## CI and pre-commit

Prefer running the policy as a normal test:

```yaml
repos:
  - repo: local
    hooks:
      - id: tddcheck
        name: tddcheck
        entry: go test ./... -run TestArchitecture
        language: system
        pass_filenames: false
        types: [go]
```

The CLI is useful when you do not want to add a policy test to the target
project:

```yaml
repos:
  - repo: local
    hooks:
      - id: tddcheck
        name: tddcheck
        entry: tddcheck --root internal
        language: system
        pass_filenames: false
        types: [go]
```

## License

MIT
