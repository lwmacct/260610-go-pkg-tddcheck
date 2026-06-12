// Package entity 检查具体实体和值对象类型的归属。
//
// 它会报告 entity.go 中的非具体类型或别名类型、entity.go 中的包级函数、接收者类型未在
// entity.go 中声明的实体方法，以及在 entity.go 之外声明的实体接收者方法。
package entity
