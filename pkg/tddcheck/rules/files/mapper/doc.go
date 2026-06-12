// Package mapper 检查纯转换函数的边界。
//
// 它会报告 mapper.go 中被禁止的导入、带接收者的 mapper 函数、mapper.go 中未以 To
// 开头的函数，以及在 mapper.go 之外声明的 mapper 类转换函数。
package mapper
