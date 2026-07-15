// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package hclwrite

import (
	"reflect"
	"strings"

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
			k := item.KeyObj()

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
		return item.ValueObj()
	}
}

func (object *ObjectConsExpr) RemoveItem(key string) bool {
	node := object.nodeFor(key)
	if node == nil {
		return false
	}

	node.Detach()
	object.items.Remove(node)
	return true
}

// SetItemRaw either replaces the expression of an existing item of the given
// name or adds a new item definition to the end of the object, using the given
// tokens verbatim as the expression.
//
// The same caveats apply to this function as for NewExpressionRaw on which it
// is based. If possible, prefer to use SetItemValue or SetItemTraversal.
func (object *ObjectConsExpr) SetItemRaw(key string, tokens Tokens) (*ObjectConsKeyExpr, *ObjectConsValue) {
	item := object.ItemFor(key)
	expr := NewExpressionRaw(tokens)
	if item != nil {
		item.ValueObj().expr.Detach()
		item.ValueObj().expr = item.ValueObj().children.Append(expr)
	} else {
		item = newObjectConsItem()
		item.init(key, expr)
		if firstItemNode := object.firstItemNode(); firstItemNode == nil {
			return nil, nil
		} else {
			object.items.Add(object.children.Insert(firstItemNode, item))
		}
	}
	return item.kv()
}

// SetItemTraversal either replaces the expression of an existing item of the given
// name or adds a new item definition to the end of the object.
//
// The return value is the item that was either modified in-place or created.
func (object *ObjectConsExpr) SetItemTraversal(key string, traversal hcl.Traversal) (*ObjectConsKeyExpr, *ObjectConsValue) {
	item := object.ItemFor(key)
	expr := NewExpressionAbsTraversal(traversal)
	if item != nil {
		item.ValueObj().expr.Detach()
		item.ValueObj().expr = item.ValueObj().children.Append(expr)
	} else {
		item = newObjectConsItem()
		item.init(key, expr)
		if firstItemNode := object.firstItemNode(); firstItemNode == nil {
			return nil, nil
		} else {
			object.items.Add(object.children.Insert(firstItemNode, item))
		}
	}
	return item.kv()
}

// SetItemValue either replaces the expression of an existing item of the given
// name or adds a new item definition to the end of the object.
//
// The value is given as a cty.Value, and must therefore be a literal. To set a
// variable reference or other traversal, use SetItemTraversal.
//
// The return value is the item that was either modified in-place or created.
func (object *ObjectConsExpr) SetItemValue(key string, val cty.Value) (*ObjectConsKeyExpr, *ObjectConsValue) {
	item := object.ItemFor(key)
	expr := NewExpressionLiteral(val)
	if item != nil {
		item.ValueObj().expr.Detach()
		item.ValueObj().expr = item.ValueObj().children.Append(expr)
	} else {
		item = newObjectConsItem()
		item.init(key, expr)
		if firstItemNode := object.firstItemNode(); firstItemNode == nil {
			return nil, nil
		} else {
			object.items.Add(object.children.Insert(firstItemNode, item))
		}
	}
	return item.kv()
}

func (object *ObjectConsExpr) firstItemNode() *node {
	return object.items.List()[0]
}

func (object *ObjectConsExpr) nodeFor(key string) *node {
	for _, n := range object.items.List() {
		if item, ok := n.content.(*ObjectConsItem); ok {
			k := item.KeyObj()

			maybeKey := k.String()
			if maybeKey == key {
				return n
			}
		}
	}

	return nil
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

func (item *ObjectConsItem) KeyObj() *ObjectConsKeyExpr {
	return item.key.content.(*ObjectConsKeyExpr)
}

func (item *ObjectConsItem) ValueObj() *ObjectConsValue {
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
	// item.KeyObj().children.Append(newIdentifier(newIdentToken(key)))

	item.children.AppendUnstructuredTokens(Tokens{
		{
			Type:  hclsyntax.TokenEqual,
			Bytes: []byte{'='},
		},
	})

	item.value = item.children.Append(newObjectConsValue())
	item.ValueObj().children.Append(value)
}

func (item *ObjectConsItem) kv() (*ObjectConsKeyExpr, *ObjectConsValue) {
	return item.KeyObj(), item.ValueObj()
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
