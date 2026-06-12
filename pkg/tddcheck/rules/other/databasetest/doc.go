// Package databasetest 检查可选数据库测试辅助工具的边界。
//
// 它会扫描 _test.go 文件，并报告应改用共享数据库测试辅助工具的项目专属 SQLite 设置模式。
// 可通过 rulekit.Config 配置允许的路径和检测字符串。
package databasetest
