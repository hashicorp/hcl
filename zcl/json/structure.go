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
	content, newBody, diags := b.PartialContent(schema)

	hiddenAttrs := newBody.(*body).hiddenAttrs

	var nameSuggestions []string
	for _, attrS := range schema.Attributes {
		if _, ok := hiddenAttrs[attrS.Name]; !ok {
			// Only suggest an attribute name if we didn't use it already.
			nameSuggestions = append(nameSuggestions, attrS.Name)
		}
	}
	for _, blockS := range schema.Blocks {
		// Blocks can appear multiple times, so we'll suggest their type
		// names regardless of whether they've already been used.
		nameSuggestions = append(nameSuggestions, blockS.Type)
	}

	for k, attr := range b.obj.Attrs {
		if _, ok := hiddenAttrs[k]; !ok {
			var fixItHint string
			suggestion := nameSuggestion(k, nameSuggestions)
			if suggestion != "" {
				fixItHint = fmt.Sprintf(" Did you mean %q?", suggestion)
			}

			diags = append(diags, &zcl.Diagnostic{
				Severity: zcl.DiagError,
				Summary:  "Extraneous JSON object property",
				Detail:   fmt.Sprintf("No attribute or block type is named %q.%s", k, fixItHint),
				Subject:  &attr.NameRange,
				Context:  attr.Range().Ptr(),
			})
		}
	}

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
			continue
		}
		content.Attributes[attrS.Name] = &zcl.Attribute{
			Name:      attrS.Name,
			Expr:      &expression{src: jsonAttr.Value},
			Range:     zcl.RangeBetween(jsonAttr.NameRange, jsonAttr.Value.Range()),
			NameRange: jsonAttr.NameRange,
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
		usedNames[blockS.Type] = struct{}{}
	}

	unusedBody := &body{
		obj:         b.obj,
		hiddenAttrs: usedNames,
	}

	return content, unusedBody, diags
}

// JustAttributes for JSON bodies interprets all properties of the wrapped
// JSON object as attributes and returns them.
func (b *body) JustAttributes() (map[string]*zcl.Attribute, zcl.Diagnostics) {
	attrs := make(map[string]*zcl.Attribute)
	for name, jsonAttr := range b.obj.Attrs {
		if _, hidden := b.hiddenAttrs[name]; hidden {
			continue
		}
		attrs[name] = &zcl.Attribute{
			Name:      name,
			Expr:      &expression{src: jsonAttr.Value},
			Range:     zcl.RangeBetween(jsonAttr.NameRange, jsonAttr.Value.Range()),
			NameRange: jsonAttr.NameRange,
		}
	}

	// No diagnostics possible here, since the parser already took care of
	// finding duplicates and every JSON value can be a valid attribute value.
	return attrs, nil
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

func (e *expression) Value(ctx *zcl.EvalContext) (cty.Value, zcl.Diagnostics) {
	// TEMP: Since we've not yet implemented the zcl native template language
	// parser, for the moment we'll support only literal values here.
	// FIXME: Once the template language parser is implemented, parse string
	// values as templates and evaluate them.

	switch v := e.src.(type) {
	case *stringVal:
		return cty.StringVal(v.Value), nil
	case *numberVal:
		return cty.NumberVal(v.Value), nil
	case *booleanVal:
		return cty.BoolVal(v.Value), nil
	case *arrayVal:
		vals := []cty.Value{}
		for _, jsonVal := range v.Values {
			val, _ := (&expression{src: jsonVal}).Value(ctx)
			vals = append(vals, val)
		}
		return cty.TupleVal(vals), nil
	case *objectVal:
		attrs := map[string]cty.Value{}
		for name, jsonAttr := range v.Attrs {
			val, _ := (&expression{src: jsonAttr.Value}).Value(ctx)
			attrs[name] = val
		}
		return cty.ObjectVal(attrs), nil
	default:
		// Default to DynamicVal so that ASTs containing invalid nodes can
		// still be partially-evaluated.
		return cty.DynamicVal, nil
	}
}

func (e *expression) Range() zcl.Range {
	return e.src.Range()
}

func (e *expression) StartRange() zcl.Range {
	return e.src.StartRange()
}
