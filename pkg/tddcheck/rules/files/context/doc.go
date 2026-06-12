// Package context 检查 context 辅助函数和 context.WithValue 周围的本地边界。
//
// 它会报告 context.go 之外的 context 辅助函数、context.go 中的无关函数，以及 context.go
// 之外的 context.WithValue 调用。context 辅助函数的检测基于 ContextWith*、*FromContext、
// *ContextFrom 和带 context.Context 参数的 *Context 等名称。
package context
