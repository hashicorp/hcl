package hclwrite

func newTupleConsExpr() *TupleConsExpr {
	return &TupleConsExpr{
		exprs:  newNodeSet(),
		inTree: newInTree(),
	}
}

type TupleConsExpr struct {
	inTree

	exprs       nodeSet
	expandFinal bool
}
