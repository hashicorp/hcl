package zclsyntax

import (
	"github.com/apparentlymart/go-zcl/zcl"
)

// Variables returns all of the variables referenced within a given experssion.
//
// This is the implementation of the "Variables" method on every native
// expression.
func Variables(expr Expression) []zcl.Traversal {
	var vars []zcl.Traversal
	VisitAll(expr, func(n Node) zcl.Diagnostics {
		if ste, ok := n.(*ScopeTraversalExpr); ok {
			vars = append(vars, ste.Traversal)
		}
		return nil
	})
	return vars
}
