package hclhil

import (
	"fmt"

	"github.com/apparentlymart/go-cty/cty"
	"github.com/apparentlymart/go-zcl/zcl"
	hclast "github.com/hashicorp/hcl/hcl/ast"
)

// body is our implementation of zcl.Body in terms of an HCL ObjectList
type body struct {
	oli *hclast.ObjectList
}

func (b *body) Content(schema *zcl.BodySchema) (*zcl.BodyContent, zcl.Diagnostics) {
	return nil, nil
}

func (b *body) PartialContent(schema *zcl.BodySchema) (*zcl.BodyContent, zcl.Body, zcl.Diagnostics) {
	return nil, nil, nil
}

func (b *body) JustAttributes() (zcl.Attributes, zcl.Diagnostics) {
	items := b.oli.Items
	attrs := make(zcl.Attributes)
	var diags zcl.Diagnostics

	for _, item := range items {
		if len(item.Keys) == 0 {
			// Should never happen, since we don't use b.oli.Filter
			diags = append(diags, &zcl.Diagnostic{
				Severity: zcl.DiagError,
				Summary:  "Invalid item",
				Detail:   "Somehow we have an HCL item with no keys. This should never happen.",
				Context:  rangeFromHCLPos(item.Pos()).Ptr(),
			})
			continue
		}
		if len(item.Keys) > 1 {
			name := item.Keys[0].Token.Value().(string)
			diags = append(diags, &zcl.Diagnostic{
				Severity: zcl.DiagError,
				Summary:  fmt.Sprintf("Unexpected %s block", name),
				Detail:   "Blocks are not allowed here.",
				Context:  rangeFromHCLPos(item.Pos()).Ptr(),
			})
			continue
		}

		name := item.Keys[0].Token.Value().(string)

		if item.Assign.Line == 0 {
			diags = append(diags, &zcl.Diagnostic{
				Severity: zcl.DiagWarning,
				Summary:  "Block syntax used for attribute",
				Detail:   fmt.Sprintf("Attribute %q is defined using block syntax, which is deprecated. Use an equals sign after the attribute name instead.", name),
				Context:  rangeFromHCLPos(item.Pos()).Ptr(),
			})
		}

		if attrs[name] != nil {
			diags = append(diags, &zcl.Diagnostic{
				Severity: zcl.DiagError,
				Summary:  "Duplicate attribute definition",
				Detail: fmt.Sprintf(
					"Attribute %q was previously defined at %s",
					name, attrs[name].NameRange.String(),
				),
				Context: rangeFromHCLPos(item.Pos()).Ptr(),
			})
			continue
		}

		attrs[name] = &zcl.Attribute{
			Name:      name,
			Expr:      &expression{src: item.Val},
			Range:     rangeFromHCLPos(item.Pos()),
			NameRange: rangeFromHCLPos(item.Keys[0].Pos()),
		}
	}

	return attrs, diags
}

func (b *body) MissingItemRange() zcl.Range {
	return rangeFromHCLPos(b.oli.Pos())
}

// body is our implementation of zcl.Body in terms of an HCL node, which may
// internally have strings to be interpreted as HIL templates.
type expression struct {
	src hclast.Node
}

func (e *expression) Value(ctx *zcl.EvalContext) (cty.Value, zcl.Diagnostics) {
	// TODO: Implement
	return cty.NilVal, nil
}

func (e *expression) Range() zcl.Range {
	return rangeFromHCLPos(e.src.Pos())
}
func (e *expression) StartRange() zcl.Range {
	return rangeFromHCLPos(e.src.Pos())
}
