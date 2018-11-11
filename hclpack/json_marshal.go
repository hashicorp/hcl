package hclpack

import (
	"encoding/json"

	"github.com/hashicorp/hcl2/hcl"
)

// MarshalJSON is an implementation of Marshaler from encoding/json, allowing
// bodies to be included in other types that are JSON-marshalable.
//
// The result of MarshalJSON is optimized for compactness rather than easy
// human consumption/editing. Use UnmarshalJSON to decode it.
func (b *Body) MarshalJSON() ([]byte, error) {
	rngs := make(map[hcl.Range]struct{})
	b.addRanges(rngs)

	fns, posList, posMap := packPositions(rngs)

	head := jsonHeader{
		Body:    b.forJSON(posMap),
		Sources: fns,
		Pos:     posList,
	}

	return json.Marshal(&head)
}

func (b *Body) forJSON(pos map[string]map[hcl.Pos]posOfs) bodyJSON {
	var ret bodyJSON

	if len(b.Attributes) > 0 {
		ret.Attrs = make(map[string]attrJSON, len(b.Attributes))
		for name, attr := range b.Attributes {
			ret.Attrs[name] = attr.forJSON(pos)
		}
	}
	if len(b.ChildBlocks) > 0 {
		ret.Blocks = make([]blockJSON, len(b.ChildBlocks))
		for i, block := range b.ChildBlocks {
			ret.Blocks[i] = block.forJSON(pos)
		}
	}
	ret.Ranges = make(rangesPacked, 1)
	ret.Ranges[0] = packRange(b.MissingItemRange_, pos)

	return ret
}

func (a *Attribute) forJSON(pos map[string]map[hcl.Pos]posOfs) attrJSON {
	var ret attrJSON

	ret.Source = string(a.Expr.Source)
	switch a.Expr.SourceType {
	case ExprNative:
		ret.Syntax = 0
	case ExprTemplate:
		ret.Syntax = 1
	case ExprLiteralJSON:
		ret.Syntax = 2
	}
	ret.Ranges = make(rangesPacked, 4)
	ret.Ranges[0] = packRange(a.Range, pos)
	ret.Ranges[1] = packRange(a.NameRange, pos)
	ret.Ranges[2] = packRange(a.Expr.Range_, pos)
	ret.Ranges[3] = packRange(a.Expr.StartRange_, pos)

	return ret
}

func (b *Block) forJSON(pos map[string]map[hcl.Pos]posOfs) blockJSON {
	var ret blockJSON

	ret.Header = make([]string, len(b.Labels)+1)
	ret.Header[0] = b.Type
	copy(ret.Header[1:], b.Labels)
	ret.Body = b.Body.forJSON(pos)
	ret.Ranges = make(rangesPacked, 2+len(b.LabelRanges))
	ret.Ranges[0] = packRange(b.DefRange, pos)
	ret.Ranges[1] = packRange(b.TypeRange, pos)
	for i, rng := range b.LabelRanges {
		ret.Ranges[i+2] = packRange(rng, pos)
	}

	return ret
}

// UnmarshalJSON is an implementation of Unmarshaler from encoding/json,
// allowing bodies to be included in other types that are JSON-unmarshalable.
func (b *Body) UnmarshalJSON(data []byte) error {
	var head jsonHeader
	err := json.Unmarshal(data, &head)
	if err != nil {
		return err
	}

	fns := head.Sources
	positions := head.Pos.Unpack()

	*b = head.Body.decode(fns, positions)

	return nil
}

type jsonHeader struct {
	Body bodyJSON `json:"r"`

	Sources []string        `json:"s,omitempty"`
	Pos     positionsPacked `json:"p,omitempty"`
}

type bodyJSON struct {
	// Files are the source filenames that were involved in
	Attrs  map[string]attrJSON `json:"a,omitempty"`
	Blocks []blockJSON         `json:"b,omitempty"`

	// Ranges contains the MissingItemRange
	Ranges rangesPacked `json:"r,omitempty"`
}

func (bj *bodyJSON) decode(fns []string, positions []position) Body {
	var ret Body

	if len(bj.Attrs) > 0 {
		ret.Attributes = make(map[string]Attribute, len(bj.Attrs))
		for name, aj := range bj.Attrs {
			ret.Attributes[name] = aj.decode(fns, positions)
		}
	}

	if len(bj.Blocks) > 0 {
		ret.ChildBlocks = make([]Block, len(bj.Blocks))
		for i, blj := range bj.Blocks {
			ret.ChildBlocks[i] = blj.decode(fns, positions)
		}
	}

	ret.MissingItemRange_ = bj.Ranges.UnpackIdx(fns, positions, 0)

	return ret
}

type attrJSON struct {
	// To keep things compact, in the JSON encoding we flatten the
	// expression down into the attribute object, since overhead
	// for attributes adds up in a complex config.
	Source string `json:"s"`
	Syntax int    `json:"t,omitempty"` // omitted for 0=native

	// Ranges contains the Range, NameRange, Expr.Range, Expr.StartRange
	Ranges rangesPacked `json:"r,omitempty"`
}

func (aj *attrJSON) decode(fns []string, positions []position) Attribute {
	var ret Attribute

	ret.Expr.Source = []byte(aj.Source)
	switch aj.Syntax {
	case 0:
		ret.Expr.SourceType = ExprNative
	case 1:
		ret.Expr.SourceType = ExprTemplate
	case 2:
		ret.Expr.SourceType = ExprLiteralJSON
	}

	ret.Range = aj.Ranges.UnpackIdx(fns, positions, 0)
	ret.NameRange = aj.Ranges.UnpackIdx(fns, positions, 1)
	ret.Expr.Range_ = aj.Ranges.UnpackIdx(fns, positions, 2)
	ret.Expr.StartRange_ = aj.Ranges.UnpackIdx(fns, positions, 3)
	if ret.Expr.StartRange_ == (hcl.Range{}) {
		// If the start range wasn't present then we'll just use the Range
		ret.Expr.StartRange_ = ret.Expr.Range_
	}

	return ret
}

type blockJSON struct {
	// Header is the type followed by any labels. We flatten this here
	// to keep the JSON encoding compact.
	Header []string `json:"h"`
	Body   bodyJSON `json:"b,omitempty"`

	// Ranges contains the DefRange followed by the TypeRange and then
	// each of the label ranges in turn.
	Ranges rangesPacked `json:"r,omitempty"`
}

func (blj *blockJSON) decode(fns []string, positions []position) Block {
	var ret Block

	if len(blj.Header) > 0 { // If the header is invalid then we'll end up with an empty type
		ret.Type = blj.Header[0]
	}
	if len(blj.Header) > 1 {
		ret.Labels = blj.Header[1:]
	}
	ret.Body = blj.Body.decode(fns, positions)

	ret.DefRange = blj.Ranges.UnpackIdx(fns, positions, 0)
	ret.TypeRange = blj.Ranges.UnpackIdx(fns, positions, 1)
	if len(ret.Labels) > 0 {
		ret.LabelRanges = make([]hcl.Range, len(ret.Labels))
		for i := range ret.Labels {
			ret.LabelRanges[i] = blj.Ranges.UnpackIdx(fns, positions, i+2)
		}
	}

	return ret
}
