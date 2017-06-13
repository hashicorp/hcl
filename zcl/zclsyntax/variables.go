package zclsyntax

import (
	"github.com/zclconf/go-zcl/zcl"
)

// Variables returns all of the variables referenced within a given experssion.
//
// This is the implementation of the "Variables" method on every native
// expression.
func Variables(expr Expression) []zcl.Traversal {
	var vars []zcl.Traversal

	// TODO: When traversing into ForExpr, filter out references to
	// the iterator variables, since they are references into the child
	// scope, and thus not interesting to the caller.

	VisitAll(expr, func(n Node) zcl.Diagnostics {
		if ste, ok := n.(*ScopeTraversalExpr); ok {
			vars = append(vars, ste.Traversal)
		}
		return nil
	})
	return vars
}
