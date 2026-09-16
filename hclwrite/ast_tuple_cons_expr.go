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
