package hcl

// AbsTraversalForExpr attempts to interpret the given expression as
// an absolute traversal, or returns error diagnostic(s) if that is
// not possible for the given expression.
//
// A particular Expression implementation can support this function by
// offering a method called AsTraversal that takes no arguments and
// returns either a valid absolute traversal or nil to indicate that
// no traversal is possible. Alternatively, an implementation can support
// UnwrapExpression to delegate handling of this function to a wrapped
// Expression object.
//
// In most cases the calling application is interested in the value
// that results from an expression, but in rarer cases the application
// needs to see the the name of the variable and subsequent
// attributes/indexes itself, for example to allow users to give references
// to the variables themselves rather than to their values. An implementer
// of this function should at least support attribute and index steps.
func AbsTraversalForExpr(expr Expression) (Traversal, Diagnostics) {
	type asTraversal interface {
		AsTraversal() Traversal
	}

	physExpr := UnwrapExpressionUntil(expr, func(expr Expression) bool {
		_, supported := expr.(asTraversal)
		return supported
	})

	if asT, supported := physExpr.(asTraversal); supported {
		if traversal := asT.AsTraversal(); traversal != nil {
			return traversal, nil
		}
	}
	return nil, Diagnostics{
		&Diagnostic{
			Severity: DiagError,
			Summary:  "Invalid expression",
			Detail:   "A static variable reference is required.",
			Subject:  expr.Range().Ptr(),
		},
	}
}

// RelTraversalForExpr is similar to AbsTraversalForExpr but it returns
// a relative traversal instead. Due to the nature of ZCL expressions, the
// first element of the returned traversal is always a TraverseAttr, and
// then it will be followed by zero or more other expressions.
//
// Any expression accepted by AbsTraversalForExpr is also accepted by
// RelTraversalForExpr.
func RelTraversalForExpr(expr Expression) (Traversal, Diagnostics) {
	traversal, diags := AbsTraversalForExpr(expr)
	if len(traversal) > 0 {
		root := traversal[0].(TraverseRoot)
		traversal[0] = TraverseAttr{
			Name:     root.Name,
			SrcRange: root.SrcRange,
		}
	}
	return traversal, diags
}
