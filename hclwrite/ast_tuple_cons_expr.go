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

func (tuple *TupleConsExpr) Items() []*Expression {
	list := tuple.exprs.List()
	items := make([]*Expression, 0, len(list))

	for _, n := range list {
		items = append(items, n.content.(*Expression))
	}

	return items
}
