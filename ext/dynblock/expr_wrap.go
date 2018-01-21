package dynblock

import (
	"github.com/hashicorp/hcl2/hcl"
	"github.com/zclconf/go-cty/cty"
)

type exprWrap struct {
	hcl.Expression
	i *iteration
}

func (e exprWrap) Variables() []hcl.Traversal {
	raw := e.Expression.Variables()
	ret := make([]hcl.Traversal, 0, len(raw))

	// Filter out traversals that refer to our iterator name or any
	// iterator we've inherited; we're going to provide those in
	// our Value wrapper, so the caller doesn't need to know about them.
	for _, traversal := range raw {
		rootName := traversal.RootName()
		if rootName == e.i.IteratorName {
			continue
		}
		if _, inherited := e.i.Inherited[rootName]; inherited {
			continue
		}
		ret = append(ret, traversal)
	}
	return ret
}

func (e exprWrap) Value(ctx *hcl.EvalContext) (cty.Value, hcl.Diagnostics) {
	extCtx := e.i.EvalContext(ctx)
	return e.Expression.Value(extCtx)
}

// Passthrough implementation for hcl.ExprList
func (e exprWrap) ExprList() []hcl.Expression {
	type exprList interface {
		ExprList() []hcl.Expression
	}

	if el, supported := e.Expression.(exprList); supported {
		return el.ExprList()
	}
	return nil
}

// Passthrough implementation for hcl.AbsTraversalForExpr and hcl.RelTraversalForExpr
func (e exprWrap) AsTraversal() hcl.Traversal {
	type asTraversal interface {
		AsTraversal() hcl.Traversal
	}

	if at, supported := e.Expression.(asTraversal); supported {
		return at.AsTraversal()
	}
	return nil
}
