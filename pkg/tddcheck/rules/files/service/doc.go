// Package service 检查 service 文件归属。
//
// 它会报告 service.go 之外的 Service 和 NewService 声明、service.go 或 service.*.go
// 之外的 Service 接收者方法、service 文件中的无关声明，以及 service.*.go 中不是 Service
// 接收者方法的声明。
package service
