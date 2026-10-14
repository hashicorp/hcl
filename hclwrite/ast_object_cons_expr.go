// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package hclwrite

import (
	"bytes"

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
// This is limited to items that have an identifier key or a quoted string
// literal key, which are equivalent. Items with a name taken from a variable
// or other expression will not be found.
//
// Example: { hat = "derby", "bat" = "cricket", (cat) = "calico" }
//
// ItemFor("hat") returns an object that represents the item `hat = "derby"`,
// and ItemFor("bat") returns the item `"bat" = "cricket"`; however, ItemFor
// cannot locate the cat.
func (o *ObjectConsExpr) ItemFor(key string) *ObjectConsItem {
	for _, n := range o.items.List() {
		if item, ok := n.content.(*ObjectConsItem); ok {
			if item.KeyExpr().hasName(key) {
				return item
			}
		}
	}

	return nil
}

func (o *ObjectConsExpr) variables() []*Traversal {
	var ret []*Traversal
	for _, n := range o.items.List() {
		item, ok := n.content.(*ObjectConsItem)
		if !ok {
			continue
		}
		key := item.key.content.(*ObjectConsKeyExpr)
		if expr, ok := key.wrapped.content.(*Expression); ok {
			ret = append(ret, expr.Variables()...)
		}
		ret = append(ret, item.value.content.(*ObjectConsValue).Expr().Variables()...)
	}
	return ret
}

// ValueExprFor finds an item that matches the given key and returns the
// expression assigned to that key.
//
// The same restrictions on keys apply as for ItemFor.
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
//
// When adding a new item, the key is written as an identifier if it is a valid
// identifier, or as a quoted string literal otherwise. The new item follows
// the existing layout: in an object written on a single line it is added on
// the same line after a comma, and otherwise it is added on its own line
// before the closing brace.
//
// It asserts that the expression is not already attached to a node, and it
// panics otherwise.
func (object *ObjectConsExpr) SetItem(key string, expr *Expression) (*ObjectConsKeyExpr, *ObjectConsValue) {
	expr.assertUnattached()
	item := object.ItemFor(key)
	if item != nil {
		item.ValueExpr().expr.Detach()
		item.ValueExpr().expr = item.ValueExpr().children.Append(expr)
	} else {
		item = newObjectConsItem()
		item.init(key, expr)
		object.items.Add(object.appendItem(item))
	}
	return item.kv()
}

// appendItem inserts a new item at the end of the object, adding whatever
// separator the object's existing layout calls for, and returns the new node.
//
// Separators and comments that follow the last existing item on its line, such
// as a trailing comma or a line comment, stay with that item. In a multi-line
// object, anything on the following lines, such as a comment before the
// closing brace, stays with the closing brace.
func (object *ObjectConsExpr) appendItem(item *ObjectConsItem) *node {
	tail := object.children.last
	toks, ok := tail.content.(Tokens)
	if !ok || len(toks) == 0 || toks[len(toks)-1].Type != hclsyntax.TokenCBrace {
		// Not the shape the parser produces, so make no layout decisions.
		return object.children.Insert(tail, item)
	}

	// Split the tokens after the last item into those that stay with it and
	// those that stay with the closing brace, so the new item can go between.
	head, brace := toks[:len(toks)-1], toks[len(toks)-1:]

	if object.isSingleLine() {
		prev := head
		if len(prev) == 0 {
			prev = tail.before.content.BuildTokens(nil)
		}
		if len(object.items) > 0 && !endsWith(prev, hclsyntax.TokenComma) {
			head = withToken(head, &Token{Type: hclsyntax.TokenComma, Bytes: []byte{','}})
		}
	} else {
		if endsLine(tail.before.content.BuildTokens(nil)) {
			// Already at the start of a line, such as after an item added
			// earlier, so everything remaining belongs to the closing brace.
			head, brace = nil, toks
		} else if i := firstLineBreak(head); i >= 0 {
			head, brace = toks[:i+1], toks[i+1:]
		} else {
			// The closing brace shares a line with the last item or the
			// opening brace, so the new item needs a line of its own.
			head = withToken(head, newlineToken())
		}
		item.children.AppendUnstructuredTokens(Tokens{newlineToken()})
	}

	braceNode := tail
	if len(head) > 0 {
		tail.content = head
		braceNode = object.children.Append(brace)
	}
	return object.children.Insert(braceNode, item)
}

// isSingleLine returns true if none of the separators directly within the
// object's braces break the line. Line breaks within item keys and values,
// such as in a nested multi-line object, don't count.
func (object *ObjectConsExpr) isSingleLine() bool {
	for n := object.children.first; n != nil; n = n.after {
		switch c := n.content.(type) {
		case Tokens:
			if hasLineBreak(c) {
				return false
			}
		case *ObjectConsItem:
			for in := c.children.first; in != nil; in = in.after {
				if lead, ok := in.content.(Tokens); ok && hasLineBreak(lead) {
					return false
				}
			}
		}
	}
	return true
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

	keyExpr := newObjectConsKeyExpr(newObjectConsKeyNode(key))
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

func newlineToken() *Token {
	return &Token{
		Type:  hclsyntax.TokenNewline,
		Bytes: []byte{'\n'},
	}
}

// withToken returns a copy of ts with tok appended. It copies because parsed
// token slices may share a backing array with neighboring nodes.
func withToken(ts Tokens, tok *Token) Tokens {
	ret := make(Tokens, len(ts), len(ts)+1)
	copy(ret, ts)
	return append(ret, tok)
}

func endsWith(ts Tokens, ty hclsyntax.TokenType) bool {
	return len(ts) > 0 && ts[len(ts)-1].Type == ty
}

// isLineBreak returns true if the token ends a line. Line comments include
// their terminating newline.
func isLineBreak(tok *Token) bool {
	switch tok.Type {
	case hclsyntax.TokenNewline:
		return true
	case hclsyntax.TokenComment:
		return bytes.HasSuffix(tok.Bytes, []byte{'\n'})
	}
	return false
}

func endsLine(ts Tokens) bool {
	return len(ts) > 0 && isLineBreak(ts[len(ts)-1])
}

func hasLineBreak(ts Tokens) bool {
	return firstLineBreak(ts) >= 0
}

// firstLineBreak returns the index of the first token that ends a line, or -1
// if there is none.
func firstLineBreak(ts Tokens) int {
	for i, tok := range ts {
		if isLineBreak(tok) {
			return i
		}
	}
	return -1
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

// hasName returns true if the key is an identifier or a quoted string literal
// with the given name.
func (k *ObjectConsKeyExpr) hasName(name string) bool {
	if n := unwrapUntilType[*identifier](k.wrapped); n != nil {
		return n.content.(*identifier).hasName(name)
	}
	if n := unwrapUntilType[*quoted](k.wrapped); n != nil {
		return n.content.(*quoted).hasValue(name)
	}
	return false
}

func newObjectConsKeyExpr(wrapped *node) *ObjectConsKeyExpr {
	expr := &ObjectConsKeyExpr{
		inTree:  newInTree(),
		wrapped: wrapped,
	}
	expr.children.AppendNode(wrapped)

	return expr
}

// newObjectConsKeyNode returns the node for a new item's key, which is written
// as an identifier when the key is a valid identifier, or as a quoted string
// literal otherwise, matching how the parser represents each form.
func newObjectConsKeyNode(key string) *node {
	if hclsyntax.ValidIdentifier(key) {
		return newNode(newIdentifier(newIdentToken(key)))
	}

	expr := newExpression()
	expr.wrap(newQuoted(TokensForValue(cty.StringVal(key))))
	return newNode(expr)
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
