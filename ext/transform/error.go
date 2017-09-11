package transform

import (
	"github.com/hashicorp/hcl2/zcl"
)

// NewErrorBody returns a zcl.Body that returns the given diagnostics whenever
// any of its content-access methods are called.
//
// The given diagnostics must have at least one diagnostic of severity
// zcl.DiagError, or this function will panic.
//
// This can be used to prepare a return value for a Transformer that
// can't complete due to an error. While the transform itself will succeed,
// the error will be returned as soon as a caller attempts to extract content
// from the resulting body.
func NewErrorBody(diags zcl.Diagnostics) zcl.Body {
	if !diags.HasErrors() {
		panic("NewErrorBody called without any error diagnostics")
	}
	return diagBody{
		Diags: diags,
	}
}

// BodyWithDiagnostics returns a zcl.Body that wraps another zcl.Body
// and emits the given diagnostics for any content-extraction method.
//
// Unlike the result of NewErrorBody, a body with diagnostics still runs
// the extraction actions on the underlying body if (and only if) the given
// diagnostics do not contain errors, but prepends the given diagnostics with
// any diagnostics produced by the action.
//
// If the given diagnostics is empty, the given body is returned verbatim.
//
// This function is intended for conveniently reporting errors and/or warnings
// produced during a transform, ensuring that they will be seen when the
// caller eventually extracts content from the returned body.
func BodyWithDiagnostics(body zcl.Body, diags zcl.Diagnostics) zcl.Body {
	if len(diags) == 0 {
		// nothing to do!
		return body
	}

	return diagBody{
		Diags:   diags,
		Wrapped: body,
	}
}

type diagBody struct {
	Diags   zcl.Diagnostics
	Wrapped zcl.Body
}

func (b diagBody) Content(schema *zcl.BodySchema) (*zcl.BodyContent, zcl.Diagnostics) {
	if b.Diags.HasErrors() {
		return b.emptyContent(), b.Diags
	}

	content, wrappedDiags := b.Wrapped.Content(schema)
	diags := make(zcl.Diagnostics, 0, len(b.Diags)+len(wrappedDiags))
	diags = append(diags, b.Diags...)
	diags = append(diags, wrappedDiags...)
	return content, diags
}

func (b diagBody) PartialContent(schema *zcl.BodySchema) (*zcl.BodyContent, zcl.Body, zcl.Diagnostics) {
	if b.Diags.HasErrors() {
		return b.emptyContent(), b.Wrapped, b.Diags
	}

	content, remain, wrappedDiags := b.Wrapped.PartialContent(schema)
	diags := make(zcl.Diagnostics, 0, len(b.Diags)+len(wrappedDiags))
	diags = append(diags, b.Diags...)
	diags = append(diags, wrappedDiags...)
	return content, remain, diags
}

func (b diagBody) JustAttributes() (zcl.Attributes, zcl.Diagnostics) {
	if b.Diags.HasErrors() {
		return nil, b.Diags
	}

	attributes, wrappedDiags := b.Wrapped.JustAttributes()
	diags := make(zcl.Diagnostics, 0, len(b.Diags)+len(wrappedDiags))
	diags = append(diags, b.Diags...)
	diags = append(diags, wrappedDiags...)
	return attributes, diags
}

func (b diagBody) MissingItemRange() zcl.Range {
	if b.Wrapped != nil {
		return b.Wrapped.MissingItemRange()
	}

	// Placeholder. This should never be seen in practice because decoding
	// a diagBody without a wrapped body should always produce an error.
	return zcl.Range{
		Filename: "<empty>",
	}
}

func (b diagBody) emptyContent() *zcl.BodyContent {
	return &zcl.BodyContent{
		MissingItemRange: b.MissingItemRange(),
	}
}
