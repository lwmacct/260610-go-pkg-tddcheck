# tddcheck

[![License](https://img.shields.io/github/license/lwmacct/260610-go-pkg-tddcheck)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/lwmacct/260610-go-pkg-tddcheck.svg)](https://pkg.go.dev/github.com/lwmacct/260610-go-pkg-tddcheck)
[![Go CI](https://github.com/lwmacct/260610-go-pkg-tddcheck/actions/workflows/go-ci.yml/badge.svg)](https://github.com/lwmacct/260610-go-pkg-tddcheck/actions/workflows/go-ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/lwmacct/260610-go-pkg-tddcheck)](https://goreportcard.com/report/github.com/lwmacct/260610-go-pkg-tddcheck)
[![GitHub Tag](https://img.shields.io/github/v/tag/lwmacct/260610-go-pkg-tddcheck?sort=semver)](https://github.com/lwmacct/260610-go-pkg-tddcheck/tags)

`tddcheck` 是一个 Go test 辅助包，用于在分层 Go 项目中执行可机械检测的架构边界规则。它使用标准 Go parser 检查文件职责、命名约定和 import 方向。

当前版本故意不兼容旧的公共 API 测试检查器。这个包现在专注于项目架构规则。

## 安装

```bash
go get github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck
```

## Go Test 用法

在拥有架构策略的项目中创建一个普通 Go 测试：

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

`Root` 可以是绝对路径，也可以是相对最近 `go.mod` 的路径。省略时默认检查 `internal`。

推荐入口是项目级检查。需要只运行某条检测时，可以直接使用 `pkg/tddcheck/rules/<files|other>/<rule>` 下的子包；共享的扫描和配置基础设施位于 `pkg/tddcheck/rulekit`。

规则包按目录分为 `rules/files` 和 `rules/other`。目录用于浏览，规则自身的机器可读分类由每个规则包的 `Meta.Kind` 提供，例如 `file`、`name`、`dependency` 和 `test`。

```go
import "github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rules/other/layerdeps"

func TestLayerDependencies(t *testing.T) {
    layerdeps.New("internal").Assert(t)
}
```

目标项目中更推荐把需要单独运行的检测整理成 `TestRules/<rule>` 子测试，例如 `internal/testutil/tddcheck/tddcheck_test.go`：

表格里保存规则对象，不保存 `func(*testing.T)` 匿名函数；这样可以避开 `golangci-lint` 的 `thelper` 规则误判。

```go
package tddcheck_test

import (
    "testing"

    "github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rules/other/databasetest"
    "github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rules/files/dto"
    "github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rules/other/layerdeps"
)

type assertRule interface {
    Assert(*testing.T)
}

func TestRules(t *testing.T) {
    tests := []struct {
        name string
        rule assertRule
    }{
        {"dependency-layerdeps", layerdeps.New("internal")},
        {"file-dto", dto.New("internal")},
        {"test-database", databasetest.New(".")},
    }

    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            test.rule.Assert(t)
        })
    }
}
```

这样可以运行全部检测，也可以只运行单条检测：

```bash
go test ./internal/testutil/tddcheck
go test ./internal/testutil/tddcheck -run 'TestRules/file-dto'
```

## 默认项目规则

`ProjectRules` 默认运行以下检查：

- `dependency-layerdeps`：`domain`、`usecase`、`adapter`、`runtime`、`infra` 的 import 遵守依赖方向。
- `name-package`：包名与目录名一致。
- `file-constants`：包级常量必须放在 `constants.go`。
- `file-entity`：具体实体和值对象类型必须放在 `entity.go`。
- `file-errors` 和 `name-error-prefix`：包级错误必须放在 `errors.go`，并使用 `Err*` 命名。
- `file-context`：context helper 和 `context.WithValue` 使用必须放在 `context.go`。
- `file-cqrs`：CQRS 结构体和接口使用明确后缀。
- `file-dto`：DTO 结构体必须放在 `dto.go`，使用 `DTO` 或 `DTOs` 后缀，并且 `dto.go` 不能声明函数。
- `file-mapper`：mapper 函数必须是 `mapper.go` 中的纯 `To*` 转换。
- `name-public-api`：导出 API 避免使用 `Validate*`、`Normalize*` 等内部职责前缀。
- `file-service`：`Service`、`NewService` 和 `Service` 方法留在 service 文件中。
- `file-validation`：校验 helper 必须放在 `validation.go`，并使用私有 `validate*` 或 `normalize*` 命名。
- `file-handler`：协议 handler 不能承载持久化职责。
- `file-repository`：repository 不能承载协议职责或 DTO 映射职责。
- `file-schema`：schema 文件专用于存储 schema 定义。
- `file-utils`：私有 `util*` helper 必须放在 `utils.go`。

数据库测试边界检查是可选项：

```go
func TestArchitecture(t *testing.T) {
    tddcheck.ProjectRules{
        Root:                 ".",
        IncludeDatabaseTests: true,
    }.Assert(t)
}
```

## 配置

默认策略保留本包最初服务的分层项目约定。其他项目可以通过 `Config` 覆盖项目相关的策略部分。

slice 字段为 `nil` 时使用默认值。非 nil 的空 slice 表示有意禁用该部分默认策略。

```go
func TestArchitecture(t *testing.T) {
    tddcheck.ProjectRules{
        Root: "internal",
        Config: tddcheck.Config{
            LayerDirs: []string{"core", "app", "adapter"},
            LayerRules: []tddcheck.LayerDependencyRule{
                {
                    SourceLayer: "core",
                    TargetLayer: "adapter",
                    Message:     "core must not import adapter",
                },
            },
            MapperForbiddenImports: []string{
                "context",
                "database/sql",
                "net/http",
            },
            HandlerForbiddenImports: []string{},
        },
    }.Assert(t)
}
```

当前可配置的策略区域包括：

- 分层目录名和层依赖规则
- 跳过扫描的目录名
- mapper 禁止 import 列表
- handler 禁止 import 列表和持久化风格调用
- repository 禁止 import 列表和实现类型名
- validation `Resolve` 签名使用的框架类型
- 数据库测试 helper 的匹配字符串和允许路径

## CLI

```bash
go install github.com/lwmacct/260610-go-pkg-tddcheck/cmd/tddcheck@latest
tddcheck --root internal
tddcheck --root . --database-tests
```

发现违规时命令退出码为 `1`，执行错误时退出码为 `2`。

## CI 和 pre-commit

推荐把策略作为普通测试运行：

```yaml
repos:
  - repo: local
    hooks:
      - id: tddcheck
        name: tddcheck
        entry: go test ./internal/testutil/tddcheck -run TestRules
        language: system
        pass_filenames: false
        types: [go]
```

如果不想在目标项目中添加策略测试，也可以使用 CLI：

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

## 许可证

MIT
