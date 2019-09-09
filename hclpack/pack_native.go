package hclpack

import (
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
)

// PackNativeFile parses the given source code as HCL native syntax and packs
// it into a hclpack Body ready to be marshalled.
//
// If the given source code contains syntax errors then error diagnostics will
// be returned. A non-nil body might still be returned in this case, which
// allows a cautious caller to still do certain analyses on the result.
func PackNativeFile(src []byte, filename string, start hcl.Pos) (*Body, hcl.Diagnostics) {
	f, diags := hclsyntax.ParseConfig(src, filename, start)
	rootBody := f.Body.(*hclsyntax.Body)
	return packNativeBody(rootBody, src), diags
}

func packNativeBody(body *hclsyntax.Body, src []byte) *Body {
	ret := &Body{}
	for name, attr := range body.Attributes {
		exprRng := attr.Expr.Range()
		exprStartRng := attr.Expr.StartRange()
		exprSrc := exprRng.SliceBytes(src)
		ret.setAttribute(name, Attribute{
			Expr: Expression{
				Source:     exprSrc,
				SourceType: ExprNative,

				Range_:      exprRng,
				StartRange_: exprStartRng,
			},
			Range:     attr.Range(),
			NameRange: attr.NameRange,
		})
	}

	for _, block := range body.Blocks {
		childBody := packNativeBody(block.Body, src)
		defRange := block.TypeRange
		if len(block.LabelRanges) > 0 {
			defRange = hcl.RangeBetween(defRange, block.LabelRanges[len(block.LabelRanges)-1])
		}
		ret.appendBlock(Block{
			Type:        block.Type,
			Labels:      block.Labels,
			Body:        *childBody,
			TypeRange:   block.TypeRange,
			DefRange:    defRange,
			LabelRanges: block.LabelRanges,
		})
	}

	ret.MissingItemRange_ = body.EndRange

	return ret
}
