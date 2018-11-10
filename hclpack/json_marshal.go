package hclpack

// MarshalJSON is an implementation of Marshaler from encoding/json, allowing
// bodies to be included in other types that are JSON-marshalable.
//
// The result of MarshalJSON is optimized for compactness rather than easy
// human consumption/editing. Use UnmarshalJSON to decode it.
func (b *Body) MarshalJSON() ([]byte, error) {
	return nil, nil
}

// UnmarshalJSON is an implementation of Unmarshaler from encoding/json,
// allowing bodies to be included in other types that are JSON-unmarshalable.
func (b *Body) UnmarshalJSON([]byte) error {
	return nil
}

type jsonHeader struct {
	Body bodyJSON `json:"b"`

	Sources []string        `json:"s,omitempty"`
	Pos     positionsPacked `json:"p,omitempty"`
}

type bodyJSON struct {
	// Files are the source filenames that were involved in
	Attrs            map[string]attrJSON `json:"a,omitempty"`
	Blocks           []blockJSON         `json:"b,omitempty"`
	MissingItemRange rangePacked         `json:"r,omitempty"`
}

type attrJSON struct {
	Expr exprJSON `json:"e"`

	// Ranges contains the full range followed by the name range
	Ranges rangesPacked `json:"r,omitempty"`
}

type blockJSON struct {
	Type   string   `json:"t"`
	Labels []string `json:"l,omitempty"`
	Body   bodyJSON `json:"b,omitempty"`

	// Ranges contains the DefRange followed by the TypeRange
	Ranges rangesPacked `json:"r,omitempty"`
}

type exprJSON struct {
	Source string `json:"s"`
	Syntax string `json:"t"`

	// Ranges contains the Range followed by the StartRange
	Ranges rangesPacked `json:"r,omitempty"`
}
