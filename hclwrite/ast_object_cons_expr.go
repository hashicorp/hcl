// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package hclwrite

import (
	"reflect"
	"strings"
)

// ObjectConsExpr represents the content of  an object-construct expression
// such as { hat = "derby", (cat) = "calico"}
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

// Items returns the content of all items in the object-construct expression.
func (o *ObjectConsExpr) Items() []*ObjectConsItem {
	list := o.items.List()
	items := make([]*ObjectConsItem, 0, len(list))

	for _, n := range list {
		items = append(items, n.content.(*ObjectConsItem))
	}

	return items
}


// ItemFor finds an item that matches the given key and returns the item.
//
// This is limited to items that have an identifier key. Items with a name
// taken from a variable will not be found.
//
// Example: { hat = "derby", (cat) = "calico" }
//
// ItemFor("hat") returns an object that represents the item `hat = "derby"`;
// however, ItemFor cannot locate the cat.
func (o *ObjectConsExpr) ItemFor(key string) *ObjectConsItem {
	for _, n := range o.items.List() {
		if item, ok := n.content.(*ObjectConsItem); ok {
			k := item.key.content.(*ObjectConsKeyExpr)

			maybeKey := k.String()
			if maybeKey == key {
				return item
			}
		}
	}

	return nil
}

// ItemFor finds an item that matches the given key and returns the value.
//
// This is limited to items that have an identifier key. Items with a name
// taken from a variable will not be found.
//
// Example: { hat = "derby", (cat) = "calico" }
//
// ItemFor("hat") returns an object that represents the value `"derby"`;
// however, ItemFor cannot locate the cat.
func (o *ObjectConsExpr) ValueFor(key string) *ObjectConsValue {
	if item := o.ItemFor(key); item == nil {
		return nil
	} else {
		return item.value.content.(*ObjectConsValue)
	}
}

// SetItemRaw either replaces the expression of an existing item of the given
// name or adds a new item definition to the end of the object, using the given
// tokens verbatim as the expression.
//
// The same caveats apply to this function as for NewExpressionRaw on which it
// is based. If possible, prefer to use SetItemValue or SetItemTraversal.
func (o *ObjectConsExpr) SetItemRaw(key string, tokens Tokens) (k *ObjectConsKeyExpr, v *ObjectConsValue) {
	item := o.ItemFor(key)
	expr := NewExpressionRaw(tokens)
	if item != nil {
		k = item.key.content.(*ObjectConsKeyExpr)
		v = item.value.content.(*ObjectConsValue)

		v.expr.list.Clear()
		v.expr = v.expr.ReplaceWith(expr)
		v.children.AppendNode(v.expr)

	} else {
		item = newObjectConsItem()

		ident := newIdentifier(TokensForIdentifier(key)[0])
		k = newObjectConsKeyExpr(newNode(ident))
		item.key = item.children.Append(k)

		v = newObjectConsValue() // TODO: expr in constructor
		v.expr = v.children.Append(expr)
		item.value = item.children.Append(v)

		node := newNode(item)
		o.children.AppendNode(node)
		o.items.Add(node)
	}
	return
}

// ObjectConsItem represents the content of a single item in an object-construct expression.
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

// ObjectConsKeyExpr represents the content that defines the name of an
// object-construct item. This content wraps an Expression.
type ObjectConsKeyExpr struct {
	inTree

	wrapped *node
}

// unwrap implements the unwrapNode interface.
func (k *ObjectConsKeyExpr) unwrap() *node {
	return k.wrapped
}

func newObjectConsKeyExpr(wrapped *node) *ObjectConsKeyExpr {
	expr := &ObjectConsKeyExpr{
		inTree:  newInTree(),
		wrapped: wrapped,
	}
	expr.children.AppendNode(wrapped)

	return expr
}

// String returns the name of the object item key, if it is specified by
// an identifier. Otherwise, it returns an empty string.
func (k *ObjectConsKeyExpr) String() string {
	if k.wrapped == nil {
		return ""
	}

	unwrapped := unwrapUntilType(k.wrapped, reflect.TypeFor[*identifier]())
	if unwrapped == nil {
		return ""
	}

	if ident, ok := unwrapped.content.(*identifier); ok {
		var b strings.Builder
		tok := ident.BuildTokens(nil)
		format(tok) // removes any SpacesBefore

		//nolint:errcheck // strings.Builder returns no error
		tok.WriteTo(&b)

		return b.String()
	}

	return ""
}

// ObjectConsValue represents the expression that is assigned to an
// object-construct item.
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
