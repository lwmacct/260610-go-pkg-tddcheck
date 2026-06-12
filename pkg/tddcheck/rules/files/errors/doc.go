// Package errors 检查包级错误 API 的归属和形态。
//
// 它会报告 errors.go 之外的包级错误变量、errors.go 中的非错误声明、不是非别名 *Error
// 类型的错误类型，以及 errors.go 中不被接受的错误辅助函数或方法。
package errors
