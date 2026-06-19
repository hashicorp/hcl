// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package hclwrite

import "strings"

type ObjectConsExpr struct {
	inTree

	items nodeSet
}

func newObjectConsExpr() *ObjectConsExpr {
	return &ObjectConsExpr{
		inTree: newInTree(),
		items:  newNodeSet(),
	}
}

func (o *ObjectConsExpr) ValueFor(key string) *ObjectConsValue {
	for _, n := range o.items.List() {
		if item, ok := n.content.(*ObjectConsItem); ok {
			k := item.key.content.(*ObjectConsKeyExpr)

			maybeKey := k.String()
			if maybeKey == key {
				return item.value.content.(*ObjectConsValue)
			}
		}
	}

	return nil
}

type ObjectConsItem struct {
	inTree
	key   *node
	value *node
}

func newObjectConsItem() *ObjectConsItem {
	return &ObjectConsItem{
		inTree: newInTree(),
	}
}

type ObjectConsKeyExpr struct {
	inTree

	literalName string
	wrapped     *node
}

func newObjectConsKeyExpr(wrapped *node) *ObjectConsKeyExpr {
	expr := &ObjectConsKeyExpr{
		inTree:  newInTree(),
		wrapped: wrapped,
	}
	expr.children.AppendNode(wrapped)

	return expr
}

func (k *ObjectConsKeyExpr) String() string {
	if k.wrapped == nil {
		return ""
	}

	if t, ok := k.wrapped.content.(*Traversal); ok && len(t.steps.List()) > 0 {
		var b strings.Builder
		tok := t.steps.List()[0].BuildTokens(nil)
		tok.WriteTo(&b)

		return b.String()
	}

	return ""
}

type ObjectConsValue struct {
	inTree
	expr *node
}

func newObjectConsValue() *ObjectConsValue {
	return &ObjectConsValue{
		inTree: newInTree(),
	}
}

func (o *ObjectConsValue) Expr() *Expression {
	return o.expr.content.(*Expression)
}
