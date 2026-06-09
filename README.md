# tddcheck

[![License](https://img.shields.io/github/license/lwmacct/260610-tddcheck)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/lwmacct/260610-tddcheck.svg)](https://pkg.go.dev/github.com/lwmacct/260610-tddcheck)
[![Go CI](https://github.com/lwmacct/260610-tddcheck/actions/workflows/go-ci.yml/badge.svg)](https://github.com/lwmacct/260610-tddcheck/actions/workflows/go-ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/lwmacct/260610-tddcheck)](https://goreportcard.com/report/github.com/lwmacct/260610-tddcheck)
[![GitHub Tag](https://img.shields.io/github/v/tag/lwmacct/260610-tddcheck?sort=semver)](https://github.com/lwmacct/260610-tddcheck/tags)

`tddcheck` is a Go AST based checker for enforcing observable TDD conventions
in local development and CI.

It cannot prove that a test was written before implementation code. It does
enforce rules that make that workflow visible: production code changes need
nearby test changes, exported APIs need candidate tests, and committed tests
must not be empty or skipped.

## Install

```bash
go install github.com/lwmacct/260610-tddcheck/cmd/tddcheck@latest
```

Library usage:

```bash
go get github.com/lwmacct/260610-tddcheck/pkg/tddcheck
```

## CLI

```bash
tddcheck --root .
tddcheck --staged
```

`--staged` reads `git diff --cached --name-only --diff-filter=ACMR` and is
intended for pre-commit usage.

## Library

```go
result, err := tddcheck.Check(ctx,
    tddcheck.WithRoot("."),
    tddcheck.WithStagedOnly(true),
)
if err != nil {
    return err
}
if !result.Passed {
    fmt.Println(result.Text())
}
```

## Rules

- `changed-code-needs-test`: staged production Go changes require a staged test
  file in the same package directory.
- `exported-decls-need-tests`: exported production functions and methods on
  exported receiver types require candidate tests.
- `no-skipped-or-empty-tests`: test functions must not be empty or contain
  `t.Skip`, `t.Skipf`, or `t.SkipNow`.

## pre-commit

```yaml
repos:
  - repo: https://github.com/lwmacct/260610-tddcheck
    rev: v0.1.0
    hooks:
      - id: tddcheck
```

## License

MIT
