// Package errorprefix 检查包级错误变量是否带有 Err 前缀。
//
// 它会扫描 errors.go 文件，并报告未以 Err 开头的包级错误变量。错误变量通过显式 error
// 类型或 errors.New、fmt.Errorf 等常见构造函数识别。
package errorprefix
