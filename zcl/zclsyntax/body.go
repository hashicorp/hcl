package zclsyntax

import (
	"github.com/apparentlymart/go-zcl/zcl"
)

// Body is the implementation of zcl.Body for the zcl native syntax.
type Body struct {
	SrcRange zcl.Range
}

// Assert that *Body implements zcl.Body
var assertBodyImplBody zcl.Body = &Body{}

func (b *Body) walkChildNodes(w internalWalkFunc) {
	// Nothing to walk yet
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
	panic("Body.MissingItemRange not yet implemented")
}
