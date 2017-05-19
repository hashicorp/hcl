package json

import (
	"fmt"

	"github.com/apparentlymart/go-cty/cty"
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
	content, _, diags := b.PartialContent(schema)

	// TODO: generate errors for the stuff we didn't use in PartialContent

	return content, diags
}

func (b *body) PartialContent(schema *zcl.BodySchema) (*zcl.BodyContent, zcl.Body, zcl.Diagnostics) {

	obj := b.obj
	jsonAttrs := obj.Attrs
	usedNames := map[string]struct{}{}
	if b.hiddenAttrs != nil {
		for k := range b.hiddenAttrs {
			usedNames[k] = struct{}{}
		}
	}
	var diags zcl.Diagnostics

	content := &zcl.BodyContent{
		Attributes: map[string]*zcl.Attribute{},
		Blocks:     nil,
	}

	for _, attrS := range schema.Attributes {
		jsonAttr, exists := jsonAttrs[attrS.Name]
		_, used := usedNames[attrS.Name]
		if used || !exists {
			if attrS.Required {
				diags = diags.Append(&zcl.Diagnostic{
					Severity: zcl.DiagError,
					Summary:  "Missing required attribute",
					Detail:   fmt.Sprintf("The attribute %q is required, so a JSON object property must be present with this name.", attrS.Name),
					Subject:  &obj.OpenRange,
				})
			}
			usedNames[attrS.Name] = struct{}{}
			continue
		}
		content.Attributes[attrS.Name] = &zcl.Attribute{
			Name:  attrS.Name,
			Expr:  &expression{src: jsonAttr.Value},
			Range: zcl.RangeBetween(jsonAttr.NameRange, jsonAttr.Value.Range()),
		}
		usedNames[attrS.Name] = struct{}{}
	}

	for _, blockS := range schema.Blocks {
		jsonAttr, exists := jsonAttrs[blockS.Type]
		_, used := usedNames[blockS.Type]
		if used || !exists {
			usedNames[blockS.Type] = struct{}{}
			continue
		}
		v := jsonAttr.Value
		diags = append(diags, b.unpackBlock(v, blockS.Type, &jsonAttr.NameRange, blockS.LabelNames, nil, nil, &content.Blocks)...)
	}

	unusedBody := &body{
		obj:         b.obj,
		hiddenAttrs: usedNames,
	}

	return content, unusedBody, diags
}

func (b *body) unpackBlock(v node, typeName string, typeRange *zcl.Range, labelsLeft []string, labelsUsed []string, labelRanges []zcl.Range, blocks *zcl.Blocks) (diags zcl.Diagnostics) {
	if len(labelsLeft) > 0 {
		labelName := labelsLeft[0]
		ov, ok := v.(*objectVal)
		if !ok {
			diags = diags.Append(&zcl.Diagnostic{
				Severity: zcl.DiagError,
				Summary:  "Incorrect JSON value type",
				Detail:   fmt.Sprintf("A JSON object is required, whose keys represent the %q block's %s.", typeName, labelName),
				Subject:  v.StartRange().Ptr(),
			})
			return
		}
		labelsUsed := append(labelsUsed, "")
		labelRanges := append(labelRanges, zcl.Range{})
		for pk, p := range ov.Attrs {
			labelsUsed[len(labelsUsed)-1] = pk
			labelRanges[len(labelRanges)-1] = p.NameRange
			diags = append(diags, b.unpackBlock(p.Value, typeName, typeRange, labelsLeft[1:], labelsUsed, labelRanges, blocks)...)
		}
		return
	}

	// By the time we get here, we've peeled off all the labels and we're ready
	// to deal with the block's actual content.

	// need to copy the label slices because their underlying arrays will
	// continue to be mutated after we return.
	labels := make([]string, len(labelsUsed))
	copy(labels, labelsUsed)
	labelR := make([]zcl.Range, len(labelRanges))
	copy(labelR, labelRanges)

	switch tv := v.(type) {
	case *objectVal:
		// Single instance of the block
		*blocks = append(*blocks, &zcl.Block{
			Type:   typeName,
			Labels: labels,
			Body: &body{
				obj: tv,
			},

			DefRange:    tv.OpenRange,
			TypeRange:   *typeRange,
			LabelRanges: labelR,
		})
	case *arrayVal:
		// Multiple instances of the block
		for _, av := range tv.Values {
			ov, ok := av.(*objectVal)
			if !ok {
				diags = diags.Append(&zcl.Diagnostic{
					Severity: zcl.DiagError,
					Summary:  "Incorrect JSON value type",
					Detail:   fmt.Sprintf("A JSON object is required, representing the contents of a %q block.", typeName),
					Subject:  v.StartRange().Ptr(),
				})
				continue
			}

			*blocks = append(*blocks, &zcl.Block{
				Type:   typeName,
				Labels: labels,
				Body: &body{
					obj: ov,
				},

				DefRange:    tv.OpenRange,
				TypeRange:   *typeRange,
				LabelRanges: labelR,
			})
		}
	default:
		diags = diags.Append(&zcl.Diagnostic{
			Severity: zcl.DiagError,
			Summary:  "Incorrect JSON value type",
			Detail:   fmt.Sprintf("Either a JSON object or a JSON array is required, representing the contents of one or more %q blocks.", typeName),
			Subject:  v.StartRange().Ptr(),
		})
	}
	return
}

func (e *expression) LiteralValue() (cty.Value, zcl.Diagnostics) {
	// TODO: Implement
	return cty.NilVal, nil
}
