package zclsyntax

import (
	"github.com/zclconf/go-zcl/zcl"
)

// AsZCLBlock returns the block data expressed as a *zcl.Block.
func (b *Block) AsZCLBlock() *zcl.Block {
	lastHeaderRange := b.TypeRange
	if len(b.LabelRanges) > 0 {
		lastHeaderRange = b.LabelRanges[len(b.LabelRanges)-1]
	}

	return &zcl.Block{
		Type:   b.Type,
		Labels: b.Labels,
		Body:   b.Body,

		DefRange:    zcl.RangeBetween(b.TypeRange, lastHeaderRange),
		TypeRange:   b.TypeRange,
		LabelRanges: b.LabelRanges,
	}
}

// Body is the implementation of zcl.Body for the zcl native syntax.
type Body struct {
	Attributes Attributes
	Blocks     Blocks

	SrcRange zcl.Range
	EndRange zcl.Range // Final token of the body, for reporting missing items
}

// Assert that *Body implements zcl.Body
var assertBodyImplBody zcl.Body = &Body{}

func (b *Body) walkChildNodes(w internalWalkFunc) {
	b.Attributes = w(b.Attributes).(Attributes)
	b.Blocks = w(b.Blocks).(Blocks)
}

func (b *Body) Range() zcl.Range {
	return b.SrcRange
}

func (b *Body) Content(schema *zcl.BodySchema) (*zcl.BodyContent, zcl.Diagnostics) {
	panic("Body.Content not yet implemented")
}

func (b *Body) PartialContent(schema *zcl.BodySchema) (*zcl.BodyContent, zcl.Body, zcl.Diagnostics) {
	panic("Body.PartialContent not yet implemented")
}

func (b *Body) JustAttributes() (zcl.Attributes, zcl.Diagnostics) {
	panic("Body.JustAttributes not yet implemented")
}

func (b *Body) MissingItemRange() zcl.Range {
	return b.EndRange
}

// Attributes is the collection of attribute definitions within a body.
type Attributes map[string]*Attribute

func (a Attributes) walkChildNodes(w internalWalkFunc) {
	for k, attr := range a {
		a[k] = w(attr).(*Attribute)
	}
}

// Range returns the range of some arbitrary point within the set of
// attributes, or an invalid range if there are no attributes.
//
// This is provided only to complete the Node interface, but has no practical
// use.
func (a Attributes) Range() zcl.Range {
	// An attributes doesn't really have a useful range to report, since
	// it's just a grouping construct. So we'll arbitrarily take the
	// range of one of the attributes, or produce an invalid range if we have
	// none. In practice, there's little reason to ask for the range of
	// an Attributes.
	for _, attr := range a {
		return attr.Range()
	}
	return zcl.Range{
		Filename: "<unknown>",
	}
}

// Attribute represents a single attribute definition within a body.
type Attribute struct {
	Name string
	Expr Expression

	NameRange   zcl.Range
	EqualsRange zcl.Range
}

func (a *Attribute) walkChildNodes(w internalWalkFunc) {
	a.Expr = w(a.Expr).(Expression)
}

func (a *Attribute) Range() zcl.Range {
	return zcl.RangeBetween(a.NameRange, a.Expr.Range())
}

// Blocks is the list of nested blocks within a body.
type Blocks []*Block

func (bs Blocks) walkChildNodes(w internalWalkFunc) {
	for i, block := range bs {
		bs[i] = w(block).(*Block)
	}
}

// Range returns the range of some arbitrary point within the list of
// blocks, or an invalid range if there are no blocks.
//
// This is provided only to complete the Node interface, but has no practical
// use.
func (bs Blocks) Range() zcl.Range {
	if len(bs) > 0 {
		return bs[0].Range()
	}
	return zcl.Range{
		Filename: "<unknown>",
	}
}

// Block represents a nested block structure
type Block struct {
	Type   string
	Labels []string
	Body   *Body

	TypeRange       zcl.Range
	LabelRanges     []zcl.Range
	OpenBraceRange  zcl.Range
	CloseBraceRange zcl.Range
}

func (b *Block) walkChildNodes(w internalWalkFunc) {
	b.Body = w(b.Body).(*Body)
}

func (b *Block) Range() zcl.Range {
	return zcl.RangeBetween(b.TypeRange, b.CloseBraceRange)
}
