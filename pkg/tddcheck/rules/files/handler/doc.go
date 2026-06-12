// Package handler 检查协议处理器边界。
//
// 它会报告 handler 文件中的持久化导入或调用、handler.go 或 handler.*.go 之外的 Handler
// 接收者方法、handler 文件中的无关声明、非 DTO 的 HTTP body schema 类型、应使用 httpapi
// 辅助工具的纯 body 包装结构体、应使用验证标签或 resolver 的必填字段 Huma bad request，
// 以及使用 *struct{} 输入的 Huma handler。
package handler
