// Package validation 检查验证辅助函数的归属和命名。
//
// 它会报告 validation.go 之外的验证辅助函数、validation.go 中的导出变量或常量、
// validation.go 中的类型声明、未以 validate 或 normalize 开头的函数，以及不符合配置的
// Resolve 签名的接收者方法。
package validation
