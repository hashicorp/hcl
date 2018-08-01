package hclwrite

import (
	"github.com/hashicorp/hcl2/hcl"
	"github.com/zclconf/go-cty/cty"
)

type Expression struct {
	inTree

	absTraversals nodeSet
}

func newExpression() *Expression {
	return &Expression{
		inTree:        newInTree(),
		absTraversals: newNodeSet(),
	}
}

// NewExpressionLiteral constructs an an expression that represents the given
// literal value.
func NewExpressionLiteral(val cty.Value) *Expression {
	panic("NewExpressionLiteral not yet implemented")
}

// NewExpressionAbsTraversal constructs an expression that represents the
// given traversal, which must be absolute or this function will panic.
func NewExpressionAbsTraversal(traversal hcl.Traversal) {
	panic("NewExpressionAbsTraversal not yet implemented")
}

type Traversal struct {
	inTree

	steps nodeSet
}

func newTraversal() *Traversal {
	return &Traversal{
		inTree: newInTree(),
		steps:  newNodeSet(),
	}
}

type TraverseName struct {
	inTree

	name *node
}

func newTraverseName() *TraverseName {
	return &TraverseName{
		inTree: newInTree(),
	}
}

type TraverseIndex struct {
	inTree

	key *node
}

func newTraverseIndex() *TraverseIndex {
	return &TraverseIndex{
		inTree: newInTree(),
	}
}
