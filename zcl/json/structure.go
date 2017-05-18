package json

import (
	"github.com/apparentlymart/go-zcl/zcl"
)

// body is the implementation of "Body" used for files processed with the JSON
// parser.
type body struct {
	obj *objectVal

	// If non-nil, the keys of this map cause the corresponding attributes to
	// be treated as non-existing. This is used when Body.PartialContent is
	// called, to produce the "remaining content" Body.
	hiddenAttrs map[string]struct{}
}

// expression is the implementation of "Expression" used for files processed
// with the JSON parser.
type expression struct {
	src node
}

func (b *body) Content(schema *zcl.BodySchema) (*zcl.BodyContent, zcl.Diagnostics) {
	// TODO: Implement
	return nil, nil
}

func (b *body) PartialContent(schema *zcl.BodySchema) (*zcl.BodyContent, zcl.Body, zcl.Diagnostics) {
	// TODO: Implement
	return nil, nil, nil
}
