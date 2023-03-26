// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hclwrite

import (
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

type Body struct {
	inTree

	items nodeSet
}

func newBody() *Body {
	return &Body{
		inTree: newInTree(),
		items:  newNodeSet(),
	}
}

func (b *Body) appendItem(c nodeContent) *node {
	nn := b.children.Append(c)
	b.items.Add(nn)
	return nn
}

func (b *Body) appendItemNode(nn *node) *node {
	nn.assertUnattached()
	b.children.AppendNode(nn)
	b.items.Add(nn)
	return nn
}

// Clear removes all of the items from the body, making it empty.
func (b *Body) Clear() {
	b.children.Clear()
}

func (b *Body) AppendUnstructuredTokens(ts Tokens) {
	b.inTree.children.Append(ts)
}

// Attributes returns a new map of all of the attributes in the body, with
// the attribute names as the keys.
func (b *Body) Attributes() map[string]*Attribute {
	ret := make(map[string]*Attribute)
	for n := range b.items {
		if attr, isAttr := n.content.(*Attribute); isAttr {
			nameObj := attr.name.content.(*identifier)
			name := string(nameObj.token.Bytes)
			ret[name] = attr
		}
	}
	return ret
}

// Blocks returns a new slice of all the blocks in the body.
func (b *Body) Blocks() []*Block {
	ret := make([]*Block, 0, len(b.items))
	for _, n := range b.items.List() {
		if block, isBlock := n.content.(*Block); isBlock {
			ret = append(ret, block)
		}
	}
	return ret
}

// GetAttribute returns the attribute from the body that has the given name,
// or returns nil if there is currently no matching attribute.
func (b *Body) GetAttribute(name string) *Attribute {
	for n := range b.items {
		if attr, isAttr := n.content.(*Attribute); isAttr {
			nameObj := attr.name.content.(*identifier)
			if nameObj.hasName(name) {
				// We've found it!
				return attr
			}
		}
	}

	return nil
}

// getAttributeNode is like GetAttribute but it returns the node containing
// the selected attribute (if one is found) rather than the attribute itself.
func (b *Body) getAttributeNode(name string) *node {
	for n := range b.items {
		if attr, isAttr := n.content.(*Attribute); isAttr {
			nameObj := attr.name.content.(*identifier)
			if nameObj.hasName(name) {
				// We've found it!
				return n
			}
		}
	}

	return nil
}

// FirstMatchingBlock returns a first matching block from the body that has the
// given name and labels or returns nil if there is currently no matching
// block.
func (b *Body) FirstMatchingBlock(typeName string, labels []string) *Block {
	for _, block := range b.Blocks() {
		if typeName == block.Type() {
			labelNames := block.Labels()
			if len(labels) == 0 && len(labelNames) == 0 {
				return block
			}
			if reflect.DeepEqual(labels, labelNames) {
				return block
			}
		}
	}

	return nil
}

// RemoveBlock removes the given block from the body, if it's in that body.
// If it isn't present, this is a no-op.
//
// Returns true if it removed something, or false otherwise.
func (b *Body) RemoveBlock(block *Block) bool {
	for n := range b.items {
		if n.content == block {
			n.Detach()
			b.items.Remove(n)
			return true
		}
	}
	return false
}

// AttributeLeadComments returns the comments that annotate an attribute.
//
// The return value is nil if there was no such attribute, an empty slice if
// it had no lead comments, or a slice of strings representing the raw text of
// the comments, including leading and trailing comment markers.
func (b *Body) AttributeLeadComments(name string) []string {
	attr := b.GetAttribute(name)
	if attr == nil {
		return nil
	}

	tokens := attr.leadComments.content.(*comments).tokens
	attrComments := make([]string, len(tokens))
	for i := range tokens {
		attrComments[i] = string(tokens[i].Bytes)
	}
	return attrComments
}

// SetAttributeLeadComments replaces the comments that annotate an attribute.
//
// Each item should contain any leading and trailing markers, i.e. single-
// line comments should start with either “#” or “//” and end with a
// newline character, and block comments should start with “/*” and end with
// “*/”.
func (b *Body) SetAttributeLeadComments(name string, leadComments []string) *Attribute {
	attr := b.GetAttribute(name)
	if attr == nil {
		return nil
	}

	c := attr.leadComments.content.(*comments)
	c.tokens = make(Tokens, len(leadComments))

	for i := range leadComments {
		c.tokens[i] = &Token{
			Type:  hclsyntax.TokenComment,
			Bytes: []byte(leadComments[i]),
		}
	}

	return attr
}

// Single-comment alternative to [SetAttributeLeadComments]
func (b *Body) SetAttributeLeadComment(name string, leadComment string) *Attribute {
	return b.SetAttributeLeadComments(name, []string{leadComment})
}

// AttributeLineComments returns the comments that follow an attribute on
// the same line.
//
// The return value is nil if there was no such attribute, an empty slice if
// it had no line comments, or a slice of strings representing the raw text of
// the comments, including leading and trailing comment markers.
func (b *Body) AttributeLineComments(name string) []string {
	attr := b.GetAttribute(name)
	if attr == nil {
		return nil
	}

	tokens := attr.lineComments.content.(*comments).tokens
	attrComments := make([]string, len(tokens))
	for i := range tokens {
		attrComments[i] = string(tokens[i].Bytes)
	}
	return attrComments
}

// SetAttributeLineComments replaces the comments that follow an attribute on
// the same line.
//
// Each item should contain any leading and trailing markers, i.e. single-
// line comments should start with either “#” or “//” and end with a
// newline character, and block comments should start with “/*” and end with
// “*/”.
//
// While it is technically possible to set multiple single-line comments using
// this function, doing so is discouraged since the parser would never produce
// such constructs. For single-line comments, prefer to use
// [SetAttributeLineComment].
func (b *Body) SetAttributeLineComments(name string, lineComments []string) *Attribute {
	attr := b.GetAttribute(name)
	if attr == nil {
		return nil
	}

	c := attr.lineComments.content.(*comments)
	c.tokens = make(Tokens, len(lineComments))

	for i := range lineComments {
		c.tokens[i] = &Token{
			Type:  hclsyntax.TokenComment,
			Bytes: []byte(lineComments[i]),
		}
	}

	return attr
}

// Single-comment alternative to [SetAttributeLineComments]
func (b *Body) SetAttributeLineComment(name string, lineComment string) *Attribute {
	return b.SetAttributeLineComments(name, []string{lineComment})
}

// SetAttributeRaw either replaces the expression of an existing attribute
// of the given name or adds a new attribute definition to the end of the block,
// using the given tokens verbatim as the expression.
//
// The same caveats apply to this function as for NewExpressionRaw on which
// it is based. If possible, prefer to use SetAttributeValue or
// SetAttributeTraversal.
func (b *Body) SetAttributeRaw(name string, tokens Tokens) *Attribute {
	attr := b.GetAttribute(name)
	expr := NewExpressionRaw(tokens)
	if attr != nil {
		attr.expr = attr.expr.ReplaceWith(expr)
	} else {
		attr := newAttribute()
		attr.init(name, expr)
		b.appendItem(attr)
	}
	return attr
}

// SetAttributeValue either replaces the expression of an existing attribute
// of the given name or adds a new attribute definition to the end of the block.
//
// The value is given as a cty.Value, and must therefore be a literal. To set
// a variable reference or other traversal, use SetAttributeTraversal.
//
// The return value is the attribute that was either modified in-place or
// created.
func (b *Body) SetAttributeValue(name string, val cty.Value) *Attribute {
	attr := b.GetAttribute(name)
	expr := NewExpressionLiteral(val)
	if attr != nil {
		attr.expr = attr.expr.ReplaceWith(expr)
	} else {
		attr := newAttribute()
		attr.init(name, expr)
		b.appendItem(attr)
	}
	return attr
}

// SetAttributeTraversal either replaces the expression of an existing attribute
// of the given name or adds a new attribute definition to the end of the body.
//
// The new expression is given as a hcl.Traversal, which must be an absolute
// traversal. To set a literal value, use SetAttributeValue.
//
// The return value is the attribute that was either modified in-place or
// created.
func (b *Body) SetAttributeTraversal(name string, traversal hcl.Traversal) *Attribute {
	attr := b.GetAttribute(name)
	expr := NewExpressionAbsTraversal(traversal)
	if attr != nil {
		attr.expr = attr.expr.ReplaceWith(expr)
	} else {
		attr := newAttribute()
		attr.init(name, expr)
		b.appendItem(attr)
	}
	return attr
}

// RemoveAttribute removes the attribute with the given name from the body.
//
// The return value is the attribute that was removed, or nil if there was
// no such attribute (in which case the call was a no-op).
func (b *Body) RemoveAttribute(name string) *Attribute {
	node := b.getAttributeNode(name)
	if node == nil {
		return nil
	}
	node.Detach()
	b.items.Remove(node)
	return node.content.(*Attribute)
}

// AppendBlock appends an existing block (which must not be already attached
// to a body) to the end of the receiving body.
func (b *Body) AppendBlock(block *Block) *Block {
	b.appendItem(block)
	return block
}

// AppendNewBlock appends a new nested block to the end of the receiving body
// with the given type name and labels.
func (b *Body) AppendNewBlock(typeName string, labels []string) *Block {
	block := newBlock()
	block.init(typeName, labels)
	b.appendItem(block)
	return block
}

// AppendNewline appends a newline token to th end of the receiving body,
// which generally serves as a separator between different sets of body
// contents.
func (b *Body) AppendNewline() {
	b.AppendUnstructuredTokens(Tokens{
		{
			Type:  hclsyntax.TokenNewline,
			Bytes: []byte{'\n'},
		},
	})
}
