# tddcheck

[![License](https://img.shields.io/github/license/lwmacct/260610-go-pkg-tddcheck)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/lwmacct/260610-go-pkg-tddcheck.svg)](https://pkg.go.dev/github.com/lwmacct/260610-go-pkg-tddcheck)
[![Go CI](https://github.com/lwmacct/260610-go-pkg-tddcheck/actions/workflows/go-ci.yml/badge.svg)](https://github.com/lwmacct/260610-go-pkg-tddcheck/actions/workflows/go-ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/lwmacct/260610-go-pkg-tddcheck)](https://goreportcard.com/report/github.com/lwmacct/260610-go-pkg-tddcheck)
[![GitHub Tag](https://img.shields.io/github/v/tag/lwmacct/260610-go-pkg-tddcheck?sort=semver)](https://github.com/lwmacct/260610-go-pkg-tddcheck/tags)

`tddcheck` 是一个基于 Go AST 的单元测试辅助工具，用于强制执行可观察的
TDD 约定。

它无法证明测试是在实现代码之前编写的。它会执行一些规则，让这种工作流变得
可见：公共 API 需要候选测试，测试函数不能为空，并且可以在 pre-commit 中
使用可选的变更代码检查。

## 安装

```bash
go get github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck
```

## 单元测试用法

在负责该策略的项目中创建一个普通的 Go 测试：

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

如需显式配置策略：

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

根目录会从调用测试文件开始向上查找最近的 `go.mod` 来自动检测。当测试位于
它所检查的模块之外时，请使用 `WithRoot` 或 `Policy.Root`。

## 规则

默认单元测试规则：

- `PublicAPIsHaveTests`：导出的生产函数，以及导出接收者类型上的方法，
  都需要候选测试。
- `TestsAreNotEmpty`：测试函数不能为空。

变更代码规则：

- `ChangedCodeHasTests`：已暂存的生产 Go 代码变更，需要在同一包目录下有
  已暂存的测试文件。

显式启用变更代码检查：

```go
func TestChangedCodeHasTests(t *testing.T) {
    tddcheck.Assert(t, tddcheck.WithChanged(true))
}
```

## pre-commit

推荐的钩子会将该策略作为普通单元测试运行：

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

对于不想添加策略测试的项目，本仓库也发布了一个轻量 CLI 钩子：

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

## 许可证

MIT
