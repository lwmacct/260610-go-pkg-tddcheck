// Package repository 检查持久化边界规则。
//
// 它会报告 repository.go 之外的 repository 类型、构造函数和接收者方法，repository.go
// 中的无关声明，被禁止的协议导入，repository.go 中的 mapper 函数，返回 DTO 的
// repository 函数，以及应改用 OrderExpr 或 OrderBy 的不安全 Bun Order 调用。
package repository
