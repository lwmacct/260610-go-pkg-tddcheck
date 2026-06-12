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

推荐入口是项目级检查。需要只运行某条检测时，可以直接使用 `pkg/tddcheck/rules/<rule>` 下的子包；共享的扫描和配置基础设施位于 `pkg/tddcheck/rulekit`。

```go
import "github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rules/layer"

func TestLayerDependencies(t *testing.T) {
    layer.New("internal").AssertLayerDependencies(t)
}
```

## 默认项目规则

`ProjectRules` 默认运行以下检查：

- `layer`：`domain`、`usecase`、`adapter`、`runtime`、`infra` 的 import 遵守依赖方向。
- `package-name`：包名与目录名一致。
- `constants`：包级常量必须放在 `constants.go`。
- `entity`：具体实体和值对象类型必须放在 `entity.go`。
- `errors` 和 `error-prefix`：包级错误必须放在 `errors.go`，并使用 `Err*` 命名。
- `context`：context helper 和 `context.WithValue` 使用必须放在 `context.go`。
- `cqrs`：CQRS 结构体和接口使用明确后缀。
- `dto`：DTO 结构体必须放在 `dto.go`，使用 `DTO` 或 `DTOs` 后缀，并且 `dto.go` 不能声明函数。
- `mapper`：mapper 函数必须是 `mapper.go` 中的纯 `To*` 转换。
- `public-api`：导出 API 避免使用 `Validate*`、`Normalize*` 等内部职责前缀。
- `service`：`Service`、`NewService` 和 `Service` 方法留在 service 文件中。
- `validation`：校验 helper 必须放在 `validation.go`，并使用私有 `validate*` 或 `normalize*` 命名。
- `handler`：协议 handler 不能承载持久化职责。
- `repository`：repository 不能承载协议职责或 DTO 映射职责。
- `schema`：schema 文件专用于存储 schema 定义。
- `utils`：私有 `util*` helper 必须放在 `utils.go`。

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
        entry: go test ./... -run TestArchitecture
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
