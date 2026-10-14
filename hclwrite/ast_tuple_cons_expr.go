// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package hclwrite

func newTupleConsExpr() *TupleConsExpr {
	return &TupleConsExpr{
		exprs:  newNodeSet(),
		inTree: newInTree(),
	}
}

type TupleConsExpr struct {
	inTree

	exprs nodeSet
}

func (tuple *TupleConsExpr) variables() []*Traversal {
	var ret []*Traversal
	for _, n := range tuple.exprs.List() {
		expr, ok := n.content.(*Expression)
		if !ok {
			continue
		}
		ret = append(ret, expr.Variables()...)
	}
	return ret
}
