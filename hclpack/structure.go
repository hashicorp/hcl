package hclpack

import (
	"fmt"

	"github.com/hashicorp/hcl2/hcl"
)

// Body is an implementation of hcl.Body.
type Body struct {
	Attributes  map[string]Attribute
	ChildBlocks []Block

	MissingItemRange_ hcl.Range
}

var _ hcl.Body = (*Body)(nil)

// Content is an implementation of the method of the same name on hcl.Body.
//
// When Content is called directly on a hclpack.Body, all child block bodies
// are guaranteed to be of type *hclpack.Body, so callers can type-assert
// to obtain a child Body in order to serialize it separately if needed.
func (b *Body) Content(schema *hcl.BodySchema) (*hcl.BodyContent, hcl.Diagnostics) {
	return b.content(schema, nil)
}

// PartialContent is an implementation of the method of the same name on hcl.Body.
//
// The returned "remain" body may share some backing objects with the receiver,
// so neither the receiver nor the returned remain body, or any descendent
// objects within them, may be mutated after this method is used.
//
// When Content is called directly on a hclpack.Body, all child block bodies
// and the returned "remain" body are guaranteed to be of type *hclpack.Body,
// so callers can type-assert to obtain a child Body in order to serialize it
// separately if needed.
func (b *Body) PartialContent(schema *hcl.BodySchema) (*hcl.BodyContent, hcl.Body, hcl.Diagnostics) {
	remain := &Body{
		MissingItemRange_: b.MissingItemRange_,
	}
	content, diags := b.content(schema, remain)
	return content, remain, diags
}

func (b *Body) content(schema *hcl.BodySchema, remain *Body) (*hcl.BodyContent, hcl.Diagnostics) {
	if b == nil {
		b = &Body{} // We'll treat a nil body like an empty one, for convenience
	}
	var diags hcl.Diagnostics

	var attrs map[string]*hcl.Attribute
	var attrUsed map[string]struct{}
	if len(b.Attributes) > 0 {
		attrs = make(map[string]*hcl.Attribute, len(b.Attributes))
		attrUsed = make(map[string]struct{}, len(b.Attributes))
	}
	for _, attrS := range schema.Attributes {
		name := attrS.Name
		attr, exists := b.Attributes[name]
		if !exists {
			if attrS.Required {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Missing required argument",
					Detail:   fmt.Sprintf("The argument %q is required, but no definition was found.", attrS.Name),
					Subject:  &b.MissingItemRange_,
				})
			}
			continue
		}

		attrs[name] = attr.asHCLAttribute(name)
		attrUsed[name] = struct{}{}
	}

	for name, attr := range b.Attributes {
		if _, used := attrUsed[name]; used {
			continue
		}
		if remain != nil {
			remain.setAttribute(name, attr)
			continue
		}
		var suggestions []string
		for _, attrS := range schema.Attributes {
			if _, defined := attrs[name]; defined {
				continue
			}
			suggestions = append(suggestions, attrS.Name)
		}
		suggestion := nameSuggestion(name, suggestions)
		if suggestion != "" {
			suggestion = fmt.Sprintf(" Did you mean %q?", suggestion)
		} else {
			// Is there a block of the same name?
			for _, blockS := range schema.Blocks {
				if blockS.Type == name {
					suggestion = fmt.Sprintf(" Did you mean to define a block of type %q?", name)
					break
				}
			}
		}

		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Unsupported argument",
			Detail:   fmt.Sprintf("An argument named %q is not expected here.%s", name, suggestion),
			Subject:  attr.NameRange.Ptr(),
		})
	}

	blocksWanted := make(map[string]hcl.BlockHeaderSchema)
	for _, blockS := range schema.Blocks {
		blocksWanted[blockS.Type] = blockS
	}

	var blocks []*hcl.Block
	for _, block := range b.ChildBlocks {
		// Redeclare block on stack so the pointer to the body is set on the
		// correct block. https://github.com/hashicorp/hcl2/issues/72
		block := block

		blockTy := block.Type
		blockS, wanted := blocksWanted[blockTy]
		if !wanted {
			if remain != nil {
				remain.appendBlock(block)
				continue
			}
			var suggestions []string
			for _, blockS := range schema.Blocks {
				suggestions = append(suggestions, blockS.Type)
			}
			suggestion := nameSuggestion(blockTy, suggestions)
			if suggestion != "" {
				suggestion = fmt.Sprintf(" Did you mean %q?", suggestion)
			} else {
				// Is there an attribute of the same name?
				for _, attrS := range schema.Attributes {
					if attrS.Name == blockTy {
						suggestion = fmt.Sprintf(" Did you mean to define argument %q?", blockTy)
						break
					}
				}
			}

			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unsupported block type",
				Detail:   fmt.Sprintf("Blocks of type %q are not expected here.%s", blockTy, suggestion),
				Subject:  &block.TypeRange,
			})
			continue
		}

		if len(block.Labels) != len(blockS.LabelNames) {
			if len(blockS.LabelNames) == 0 {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("Extraneous label for %s", blockTy),
					Detail: fmt.Sprintf(
						"No labels are expected for %s blocks.", blockTy,
					),
					Subject: &block.DefRange,
					Context: &block.DefRange,
				})
			} else {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("Wrong label count for %s", blockTy),
					Detail: fmt.Sprintf(
						"%s blocks expect %d label(s), but got %d.",
						blockTy, len(blockS.LabelNames), len(block.Labels),
					),
					Subject: &block.DefRange,
					Context: &block.DefRange,
				})
			}
			continue
		}

		blocks = append(blocks, block.asHCLBlock())
	}

	return &hcl.BodyContent{
		Attributes:       attrs,
		Blocks:           blocks,
		MissingItemRange: b.MissingItemRange_,
	}, diags
}

// JustAttributes is an implementation of the method of the same name on hcl.Body.
func (b *Body) JustAttributes() (hcl.Attributes, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	if len(b.ChildBlocks) > 0 {
		for _, block := range b.ChildBlocks {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("Unexpected %s block", block.Type),
				Detail:   "Blocks are not allowed here.",
				Context:  &block.TypeRange,
			})
		}
		// We'll continue processing anyway, and return any attributes we find
		// so that the caller can do careful partial analysis.
	}

	if len(b.Attributes) == 0 {
		return nil, diags
	}

	ret := make(hcl.Attributes, len(b.Attributes))
	for n, a := range b.Attributes {
		ret[n] = a.asHCLAttribute(n)
	}
	return ret, diags
}

// MissingItemRange is an implementation of the method of the same name on hcl.Body.
func (b *Body) MissingItemRange() hcl.Range {
	return b.MissingItemRange_
}

func (b *Body) setAttribute(name string, attr Attribute) {
	if b.Attributes == nil {
		b.Attributes = make(map[string]Attribute)
	}
	b.Attributes[name] = attr
}

func (b *Body) appendBlock(block Block) {
	b.ChildBlocks = append(b.ChildBlocks, block)
}

func (b *Body) addRanges(rngs map[hcl.Range]struct{}) {
	rngs[b.MissingItemRange_] = struct{}{}
	for _, attr := range b.Attributes {
		attr.addRanges(rngs)
	}
	for _, block := range b.ChildBlocks {
		block.addRanges(rngs)
	}
}

// Block represents a nested block within a body.
type Block struct {
	Type   string
	Labels []string
	Body   Body

	DefRange, TypeRange hcl.Range
	LabelRanges         []hcl.Range
}

func (b *Block) asHCLBlock() *hcl.Block {
	return &hcl.Block{
		Type:   b.Type,
		Labels: b.Labels,
		Body:   &b.Body,

		TypeRange:   b.TypeRange,
		DefRange:    b.DefRange,
		LabelRanges: b.LabelRanges,
	}
}

func (b *Block) addRanges(rngs map[hcl.Range]struct{}) {
	rngs[b.DefRange] = struct{}{}
	rngs[b.TypeRange] = struct{}{}
	for _, rng := range b.LabelRanges {
		rngs[rng] = struct{}{}
	}
	b.Body.addRanges(rngs)
}

// Attribute represents an attribute definition within a body.
type Attribute struct {
	Expr Expression

	Range, NameRange hcl.Range
}

func (a *Attribute) asHCLAttribute(name string) *hcl.Attribute {
	return &hcl.Attribute{
		Name:      name,
		Expr:      &a.Expr,
		Range:     a.Range,
		NameRange: a.NameRange,
	}
}

func (a *Attribute) addRanges(rngs map[hcl.Range]struct{}) {
	rngs[a.Range] = struct{}{}
	rngs[a.NameRange] = struct{}{}
	a.Expr.addRanges(rngs)
}
