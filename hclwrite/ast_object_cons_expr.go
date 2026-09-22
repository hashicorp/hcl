// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package hclwrite

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
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
			k := item.KeyExpr()
			unwrapped := unwrapUntilType[*identifier](k.wrapped)
			if unwrapped == nil {
				continue
			}

			identifier := unwrapped.content.(*identifier)
			if identifier.hasName(key) {
				return item
			}
		}
	}

	return nil
}

// ValueExprFor finds an item that matches the given key and returns the
// expression assigned to that key.
//
// This is limited to items that have an identifier key. Items with a name
// taken from a variable will not be found.
//
// Example: { hat = "derby", (cat) = "calico" }
//
// ValueExprFor("hat") returns an object that represents the value `"derby"`;
// however, ValueExprFor cannot locate the cat.
func (o *ObjectConsExpr) ValueExprFor(key string) *ObjectConsValue {
	if item := o.ItemFor(key); item == nil {
		return nil
	} else {
		return item.ValueExpr()
	}
}

// SetItem either replaces the expression of an existing item of the given
// name or adds a new item definition to the end of the object, using the given
// expression.
func (object *ObjectConsExpr) SetItem(key string, expr *Expression) (*ObjectConsKeyExpr, *ObjectConsValue) {
	item := object.ItemFor(key)
	if item != nil {
		item.ValueExpr().expr.Detach()
		item.ValueExpr().expr = item.ValueExpr().children.Append(expr)
	} else {
		item = newObjectConsItem()
		item.init(key, expr)
		object.items.Add(object.children.Insert(object.children.last, item))
	}
	return item.kv()
}

// SetItemRaw either replaces the expression of an existing item of the given
// name or adds a new item definition to the end of the object, using the given
// tokens verbatim as the expression.
//
// The same caveats apply to this function as for NewExpressionRaw on which it
// is based. If possible, prefer to use SetItemValue or SetItemTraversal.
func (object *ObjectConsExpr) SetItemRaw(key string, tokens Tokens) (*ObjectConsKeyExpr, *ObjectConsValue) {
	expr := NewExpressionRaw(tokens)
	return object.SetItem(key, expr)
}

// SetItemTraversal either replaces the expression of an existing item of the given
// name or adds a new item definition to the end of the object.
//
// The return value is the item that was either modified in-place or created.
func (object *ObjectConsExpr) SetItemTraversal(key string, traversal hcl.Traversal) (*ObjectConsKeyExpr, *ObjectConsValue) {
	expr := NewExpressionAbsTraversal(traversal)
	return object.SetItem(key, expr)
}

// SetItemValue either replaces the expression of an existing item of the given
// name or adds a new item definition to the end of the object.
//
// The value is given as a cty.Value, and must therefore be a literal. To set a
// variable reference or other traversal, use SetItemTraversal.
//
// The return value is the item that was either modified in-place or created.
func (object *ObjectConsExpr) SetItemValue(key string, val cty.Value) (*ObjectConsKeyExpr, *ObjectConsValue) {
	expr := NewExpressionLiteral(val)
	return object.SetItem(key, expr)
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

func (item *ObjectConsItem) KeyExpr() *ObjectConsKeyExpr {
	return item.key.content.(*ObjectConsKeyExpr)
}

func (item *ObjectConsItem) ValueExpr() *ObjectConsValue {
	return item.value.content.(*ObjectConsValue)
}

func (item *ObjectConsItem) init(key string, value *Expression) {
	value.assertUnattached()

	item.children.AppendUnstructuredTokens(Tokens{
		{
			Type:  hclsyntax.TokenNewline,
			Bytes: []byte{'\n'},
		},
	})
	identifier := newIdentifier(newIdentToken(key))
	keyExpr := newObjectConsKeyExpr(newNode(identifier))
	item.key = item.children.Append(keyExpr)

	item.children.AppendUnstructuredTokens(Tokens{
		{
			Type:  hclsyntax.TokenEqual,
			Bytes: []byte{'='},
		},
	})

	item.value = item.children.Append(newObjectConsValue())
	item.ValueExpr().expr = item.ValueExpr().children.Append(value)
}

func (item *ObjectConsItem) kv() (*ObjectConsKeyExpr, *ObjectConsValue) {
	return item.KeyExpr(), item.ValueExpr()
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

// AsIdentifier returns the name of the object item key, if it is specified by
// an identifier. Otherwise, it returns an empty string.
func (k *ObjectConsKeyExpr) asIdentifier() *identifier {
	if k.wrapped == nil {
		return nil
	}

	unwrapped := unwrapUntilType[*identifier](k.wrapped)
	if unwrapped == nil {
		return nil
	}

	return unwrapped.content.(*identifier)
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
