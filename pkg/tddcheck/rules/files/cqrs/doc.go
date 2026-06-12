// Package cqrs 检查命令/查询职责分离类型的命名规则。
//
// 它会扫描 cqrs.go 文件，并报告未以 Query、Result 或 Command 结尾的结构体。它还会报告未表达
// UseCase、CommandHandler、QueryHandler、Access、Policy 或 Authorizer 等可接受用例契约的接口。
package cqrs
