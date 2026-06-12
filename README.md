# tddcheck

[![License](https://img.shields.io/github/license/lwmacct/260610-go-pkg-tddcheck)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/lwmacct/260610-go-pkg-tddcheck.svg)](https://pkg.go.dev/github.com/lwmacct/260610-go-pkg-tddcheck)
[![Go CI](https://github.com/lwmacct/260610-go-pkg-tddcheck/actions/workflows/go-ci.yml/badge.svg)](https://github.com/lwmacct/260610-go-pkg-tddcheck/actions/workflows/go-ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/lwmacct/260610-go-pkg-tddcheck)](https://goreportcard.com/report/github.com/lwmacct/260610-go-pkg-tddcheck)
[![GitHub Tag](https://img.shields.io/github/v/tag/lwmacct/260610-go-pkg-tddcheck?sort=semver)](https://github.com/lwmacct/260610-go-pkg-tddcheck/tags)

`tddcheck` is a Go AST based unit-test helper for enforcing observable TDD
conventions.

It cannot prove that a test was written before implementation code. It enforces
rules that make that workflow visible: public APIs need candidate tests, test
functions must not be empty, and optional changed-code checks can be used from
pre-commit.

## Install

```bash
go get github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck
```

## Unit Test Usage

Create a normal Go test in the project that owns the policy:

```go
package project_test

import (
    "testing"

    "github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck"
)

func TestTDDPolicy(t *testing.T) {
    tddcheck.Assert(t)
}
```

For explicit policy configuration:

```go
var policy = tddcheck.Policy{
    Ignore: []string{
        "gen/**",
        "mocks/**",
    },
}

func TestTDDPolicy(t *testing.T) {
    policy.Assert(t)
}
```

The root directory is auto-detected by walking up from the calling test file to
the nearest `go.mod`. Use `WithRoot` or `Policy.Root` when a test lives outside
the module it checks.

## Rules

Default unit-test rules:

- `PublicAPIsHaveTests`: exported production functions and methods on exported
  receiver types require candidate tests.
- `TestsAreNotEmpty`: test functions must not be empty.

Changed-code rule:

- `ChangedCodeHasTests`: staged production Go changes require a staged test file
  in the same package directory.

Enable changed-code checks explicitly:

```go
func TestChangedCodeHasTests(t *testing.T) {
    tddcheck.Assert(t, tddcheck.WithChanged(true))
}
```

## pre-commit

The preferred hook runs the policy as a normal unit test:

```yaml
repos:
  - repo: local
    hooks:
      - id: tddcheck
        name: tddcheck
        entry: go test ./... -run TestTDDPolicy
        language: system
        pass_filenames: false
        types: [go]
```

The repository also publishes a thin CLI hook for projects that do not want to
add a policy test:

```yaml
repos:
  - repo: https://github.com/lwmacct/260610-go-pkg-tddcheck
    rev: v0.1.0
    hooks:
      - id: tddcheck
```

## CLI

```bash
go install github.com/lwmacct/260610-go-pkg-tddcheck/cmd/tddcheck@latest
tddcheck --root .
tddcheck --staged
```

## License

MIT
