// Package tddcheck 提供 Go 测试辅助工具，用于在分层 Go 模块中强制检查项目架构边界。
//
// 该包刻意保持机械化：它使用标准解析器扫描 Go 源文件，并报告命名、文件归属和依赖方向违规。
// 它设计为可在普通测试和 CI 中运行。
//
// 典型用法：
//
//	func TestArchitecture(t *testing.T) {
//	    tddcheck.ProjectRules{Root: "internal"}.Assert(t)
//	}
//
// 可通过 Config 定制项目专属策略。ProjectRules 是推荐的公共入口；单独规则位于
// pkg/tddcheck/rules/<files|other>/<rule> 下，可用于聚焦检查。规则包会暴露 Meta
// 值，用于描述其稳定 ID 和类型。
package tddcheck
