package hclwrite

import (
	"github.com/hashicorp/hcl2/hcl"
	"github.com/zclconf/go-cty/cty"
)

type Body struct {
	inTree

	items nodeSet

	// indentLevel is the number of spaces that should appear at the start
	// of lines added within this body.
	indentLevel int
}

func (b *Body) appendItem(n *node) {
	b.inTree.children.AppendNode(n)
	b.items.Add(n)
}

func (b *Body) AppendUnstructuredTokens(ts Tokens) {
	b.inTree.children.Append(ts)
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

// SetAttributeValue either replaces the expression of an existing attribute
// of the given name or adds a new attribute definition to the end of the block.
//
// The value is given as a cty.Value, and must therefore be a literal. To set
// a variable reference or other traversal, use SetAttributeTraversal.
//
// The return value is the attribute that was either modified in-place or
// created.
func (b *Body) SetAttributeValue(name string, val cty.Value) *Attribute {
	panic("Body.SetAttributeValue not yet implemented")
}

// SetAttributeTraversal either replaces the expression of an existing attribute
// of the given name or adds a new attribute definition to the end of the block.
//
// The new expression is given as a hcl.Traversal, which must be an absolute
// traversal. To set a literal value, use SetAttributeValue.
//
// The return value is the attribute that was either modified in-place or
// created.
func (b *Body) SetAttributeTraversal(name string, traversal hcl.Traversal) *Attribute {
	panic("Body.SetAttributeTraversal not yet implemented")
}

type Attribute struct {
	inTree

	leadComments *node
	name         *node
	expr         *node
	lineComments *node
}

type Block struct {
	inTree

	leadComments *node
	typeName     *node
	labels       nodeSet
	open         *node
	body         *node
	close        *node
}
