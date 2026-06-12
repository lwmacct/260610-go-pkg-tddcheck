// Package layerdeps 检查已配置架构层之间的导入方向。
//
// 它会报告违反已配置层依赖规则的导入。默认情况下，它检查 domain、usecase、adapter、
// runtime 和 infra 目录，并阻止内层或基础设施层导入被禁止的目标。
package layerdeps
