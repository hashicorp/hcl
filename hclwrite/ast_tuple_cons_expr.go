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

func (tuple *TupleConsExpr) ExpandFinal() bool {
	return tuple.expandFinal
}

func (tuple *TupleConsExpr) Items() []*Expression {
	list := tuple.exprs.List()
	items := make([]*Expression, 0, len(list))

	for _, n := range list {
		items = append(items, n.content.(*Expression))
	}

	return items
}
