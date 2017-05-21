package hclhil

import (
	"github.com/apparentlymart/go-zcl/zcl"
	hclparser "github.com/hashicorp/hcl/hcl/parser"
	hcltoken "github.com/hashicorp/hcl/hcl/token"
)

// errorRange attempts to extract a source range from the given error,
// returning a pointer to the range if possible or nil if not.
//
// errorRange understands HCL's "PosError" type, which wraps an error
// with a source position.
func errorRange(err error) *zcl.Range {
	if perr, ok := err.(*hclparser.PosError); ok {
		rng := rangeFromHCLPos(perr.Pos)
		return &rng
	}

	return nil
}

func rangeFromHCLPos(pos hcltoken.Pos) zcl.Range {
	// HCL only marks single positions rather than ranges, so we adapt this
	// by creating a single-character range at the given position.
	return zcl.Range{
		Filename: pos.Filename,
		Start: zcl.Pos{
			Byte:   pos.Offset,
			Line:   pos.Line,
			Column: pos.Column,
		},
		End: zcl.Pos{
			Byte:   pos.Offset + 1,
			Line:   pos.Line,
			Column: pos.Column + 1,
		},
	}
}
