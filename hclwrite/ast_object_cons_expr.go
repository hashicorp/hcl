// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package hclwrite

type ObjectConsExpr struct {
	inTree
}

func newObjectConsExpr() *ObjectConsExpr {
	return &ObjectConsExpr{
		inTree: newInTree(),
	}
}

func (o *ObjectConsExpr) ValueFor(key string) *ObjectConsValue {
	var found *ObjectConsValue
	o.walkChildNodes(func(n *node) {
		if item, ok := n.content.(*ObjectConsItem); ok {
			k := item.key.content.(*ObjectConsKey)
			name := k.name.content.(*Expression)

			// A temporary workaround for parsing expressions as names at ...
			// parse-time
			maybeKey := ""
			for _, token := range name.BuildTokens(nil) {
				maybeKey += string(token.Bytes) // no spaces before
			}

			if k.literal && (maybeKey == key || maybeKey == `"`+key+`"`) {
				found = item.value.content.(*ObjectConsValue)
				return
			}
		}
	})
	return found
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

type ObjectConsKey struct {
	inTree

	literal bool
	name    *node
}

func newObjectConsKey() *ObjectConsKey {
	return &ObjectConsKey{
		inTree: newInTree(),
	}
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
