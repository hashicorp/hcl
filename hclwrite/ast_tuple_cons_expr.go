package hclwrite

import (
	"bytes"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

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

func (tuple *TupleConsExpr) AppendRaw(tokens Tokens) {
	var cbrackNode *node
	tuple.walkChildNodes(func(n *node) {
		if tokens, ok := n.content.(Tokens); ok {
			if tokens[0].Type == hclsyntax.TokenCBrack {
				cbrackNode = n
				return
			}
		}
	})

	if cbrackNode == nil {
		// TODO: this is a Problem.
		return
	}

	expr := NewExpressionRaw(tokens)

	length := len(tuple.exprs.List())
	if length > 0 {
		tuple.children.Insert(cbrackNode, Tokens{
			&Token{Type: hclsyntax.TokenComma, Bytes: []byte{','}},
		})
	}

	exprNode := tuple.children.Insert(cbrackNode, expr)
	tuple.exprs.Add(exprNode)
}

func (tuple *TupleConsExpr) AppendValue(value cty.Value) {
	var cbrackNode *node
	tuple.walkChildNodes(func(n *node) {
		if tokens, ok := n.content.(Tokens); ok {
			if tokens[0].Type == hclsyntax.TokenCBrack {
				cbrackNode = n
				return
			}
		}
	})

	if cbrackNode == nil {
		// TODO: this is a Problem.
		return
	}

	expr := NewExpressionLiteral(value)

	length := len(tuple.exprs.List())
	if length > 0 {
		tuple.children.Insert(cbrackNode, Tokens{
			&Token{Type: hclsyntax.TokenComma, Bytes: []byte{','}},
		})
	}
	exprNode := tuple.children.Insert(cbrackNode, expr)
	tuple.exprs.Add(exprNode)
}

func (tuple *TupleConsExpr) AddRaw(tokens Tokens) {
	formatted := Format(tokens.Bytes())
	for _, item := range tuple.Items() {
		formattedItem := Format(item.BuildTokens(nil).Bytes())
		if bytes.Equal(formatted, formattedItem) {
			return
		}
	}
	tuple.AppendRaw(tokens)
}

func (tuple *TupleConsExpr) AddValue(value cty.Value) {
	formatted := Format(TokensForValue(value).Bytes())
	for _, item := range tuple.Items() {
		formattedItem := Format(item.BuildTokens(nil).Bytes())
		if bytes.Equal(formatted, formattedItem) {
			return
		}
	}
	tuple.AppendValue(value)
}

func (tuple *TupleConsExpr) Clear() {
	tuple.children.Clear()
	tuple.children.Append(TokensForTuple(nil))

	tuple.exprs.Clear()
}

func (tuple *TupleConsExpr) Items() []*Expression {
	list := tuple.exprs.List()
	items := make([]*Expression, 0, len(list))

	for _, n := range list {
		items = append(items, n.content.(*Expression))
	}

	return items
}
