// Package publicapi 检查使用内部职责前缀的公共 API 名称。
//
// 它会报告名称以 Validate 或 Normalize 开头的导出函数。在此架构策略中，这些前缀保留给内部验证和规范化辅助函数使用。
package publicapi
