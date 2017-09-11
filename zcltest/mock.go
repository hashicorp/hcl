package zcltest

import (
	"fmt"

	"github.com/hashicorp/hcl2/zcl"
	"github.com/zclconf/go-cty/cty"
)

// MockBody returns a zcl.Body implementation that works in terms of a
// caller-constructed zcl.BodyContent, thus avoiding the need to parse
// a "real" zcl config file to use as input to a test.
func MockBody(content *zcl.BodyContent) zcl.Body {
	return mockBody{content}
}

type mockBody struct {
	C *zcl.BodyContent
}

func (b mockBody) Content(schema *zcl.BodySchema) (*zcl.BodyContent, zcl.Diagnostics) {
	content, remainI, diags := b.PartialContent(schema)
	remain := remainI.(mockBody)
	for _, attr := range remain.C.Attributes {
		diags = append(diags, &zcl.Diagnostic{
			Severity: zcl.DiagError,
			Summary:  "Extraneous attribute in mock body",
			Detail:   fmt.Sprintf("Mock body has extraneous attribute %q.", attr.Name),
			Subject:  &attr.NameRange,
		})
	}
	for _, block := range remain.C.Blocks {
		diags = append(diags, &zcl.Diagnostic{
			Severity: zcl.DiagError,
			Summary:  "Extraneous block in mock body",
			Detail:   fmt.Sprintf("Mock body has extraneous block of type %q.", block.Type),
			Subject:  &block.DefRange,
		})
	}
	return content, diags
}

func (b mockBody) PartialContent(schema *zcl.BodySchema) (*zcl.BodyContent, zcl.Body, zcl.Diagnostics) {
	ret := &zcl.BodyContent{
		Attributes:       map[string]*zcl.Attribute{},
		Blocks:           []*zcl.Block{},
		MissingItemRange: b.C.MissingItemRange,
	}
	remain := &zcl.BodyContent{
		Attributes:       map[string]*zcl.Attribute{},
		Blocks:           []*zcl.Block{},
		MissingItemRange: b.C.MissingItemRange,
	}
	var diags zcl.Diagnostics

	if len(schema.Attributes) != 0 {
		for _, attrS := range schema.Attributes {
			name := attrS.Name
			attr, ok := b.C.Attributes[name]
			if !ok {
				if attrS.Required {
					diags = append(diags, &zcl.Diagnostic{
						Severity: zcl.DiagError,
						Summary:  "Missing required attribute",
						Detail:   fmt.Sprintf("Mock body doesn't have attribute %q", name),
						Subject:  b.C.MissingItemRange.Ptr(),
					})
				}
				continue
			}
			ret.Attributes[name] = attr
		}
	}

	for attrN, attr := range b.C.Attributes {
		if _, ok := ret.Attributes[attrN]; !ok {
			remain.Attributes[attrN] = attr
		}
	}

	wantedBlocks := map[string]zcl.BlockHeaderSchema{}
	for _, blockS := range schema.Blocks {
		wantedBlocks[blockS.Type] = blockS
	}

	for _, block := range b.C.Blocks {
		if blockS, ok := wantedBlocks[block.Type]; ok {
			if len(block.Labels) != len(blockS.LabelNames) {
				diags = append(diags, &zcl.Diagnostic{
					Severity: zcl.DiagError,
					Summary:  "Wrong number of block labels",
					Detail:   fmt.Sprintf("Block of type %q requires %d labels, but got %d", blockS.Type, len(blockS.LabelNames), len(block.Labels)),
					Subject:  b.C.MissingItemRange.Ptr(),
				})
			}

			ret.Blocks = append(ret.Blocks, block)
		} else {
			remain.Blocks = append(remain.Blocks, block)
		}
	}

	return ret, mockBody{remain}, diags
}

func (b mockBody) JustAttributes() (zcl.Attributes, zcl.Diagnostics) {
	var diags zcl.Diagnostics
	if len(b.C.Blocks) != 0 {
		diags = append(diags, &zcl.Diagnostic{
			Severity: zcl.DiagError,
			Summary:  "Mock body has blocks",
			Detail:   "Can't use JustAttributes on a mock body with blocks.",
			Subject:  b.C.MissingItemRange.Ptr(),
		})
	}

	return b.C.Attributes, diags
}

func (b mockBody) MissingItemRange() zcl.Range {
	return b.C.MissingItemRange
}

// MockExprLiteral returns a zcl.Expression that evaluates to the given literal
// value.
func MockExprLiteral(val cty.Value) zcl.Expression {
	return mockExprLiteral{val}
}

type mockExprLiteral struct {
	V cty.Value
}

func (e mockExprLiteral) Value(ctx *zcl.EvalContext) (cty.Value, zcl.Diagnostics) {
	return e.V, nil
}

func (e mockExprLiteral) Variables() []zcl.Traversal {
	return nil
}

func (e mockExprLiteral) Range() zcl.Range {
	return zcl.Range{
		Filename: "MockExprLiteral",
	}
}

func (e mockExprLiteral) StartRange() zcl.Range {
	return e.Range()
}

// MockExprVariable returns a zcl.Expression that evaluates to the value of
// the variable with the given name.
func MockExprVariable(name string) zcl.Expression {
	return mockExprVariable(name)
}

type mockExprVariable string

func (e mockExprVariable) Value(ctx *zcl.EvalContext) (cty.Value, zcl.Diagnostics) {
	name := string(e)
	for ctx != nil {
		if val, ok := ctx.Variables[name]; ok {
			return val, nil
		}
		ctx = ctx.Parent()
	}

	// If we fall out here then there is no variable with the given name
	return cty.DynamicVal, zcl.Diagnostics{
		{
			Severity: zcl.DiagError,
			Summary:  "Reference to undefined variable",
			Detail:   fmt.Sprintf("Variable %q is not defined.", name),
		},
	}
}

func (e mockExprVariable) Variables() []zcl.Traversal {
	return []zcl.Traversal{
		{
			zcl.TraverseRoot{
				Name:     string(e),
				SrcRange: e.Range(),
			},
		},
	}
}

func (e mockExprVariable) Range() zcl.Range {
	return zcl.Range{
		Filename: "MockExprVariable",
	}
}

func (e mockExprVariable) StartRange() zcl.Range {
	return e.Range()
}

// MockAttrs constructs and returns a zcl.Attributes map with attributes
// derived from the given expression map.
//
// Each entry in the map becomes an attribute whose name is the key and
// whose expression is the value.
func MockAttrs(exprs map[string]zcl.Expression) zcl.Attributes {
	ret := make(zcl.Attributes)
	for name, expr := range exprs {
		ret[name] = &zcl.Attribute{
			Name: name,
			Expr: expr,
			Range: zcl.Range{
				Filename: "MockAttrs",
			},
			NameRange: zcl.Range{
				Filename: "MockAttrs",
			},
		}
	}
	return ret
}
