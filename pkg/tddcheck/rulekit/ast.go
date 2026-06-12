package rulekit

import "go/ast"

func ReceiverTypeName(receiver *ast.FieldList) string {
	if receiver == nil || len(receiver.List) == 0 {
		return ""
	}
	return ExprTypeName(receiver.List[0].Type)
}

func ExprTypeName(expr ast.Expr) string {
	switch typed := expr.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.StarExpr:
		return ExprTypeName(typed.X)
	}
	return ""
}
