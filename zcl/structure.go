package zcl

import (
	"github.com/apparentlymart/go-cty/cty"
)

// File is the top-level node that results from parsing a ZCL file.
type File struct {
	Body Body
}

// Block represents a nested block within a Body.
type Block struct {
	Type   string
	Labels []string
	Body   Body

	DefRange    Range   // Range that can be considered the "definition" for seeking in an editor
	TypeRange   Range   // Range for the block type declaration specifically.
	LabelRanges []Range // Ranges for the label values specifically.
}

// Blocks is a sequence of Block.
type Blocks []*Block

// Body is a container for attributes and blocks. It serves as the primary
// unit of heirarchical structure within configuration.
//
// The content of a body cannot be meaningfully intepreted without a schema,
// so Body represents the raw body content and has methods that allow the
// content to be extracted in terms of a given schema.
type Body interface {
	// Content verifies that the entire body content conforms to the given
	// schema and then returns it, and/or returns diagnostics. The returned
	// body content is valid if non-nil, regardless of whether Diagnostics
	// are provided, but diagnostics should still be eventually shown to
	// the user.
	Content(schema *BodySchema) (*BodyContent, Diagnostics)

	// PartialContent is like Content except that it permits the configuration
	// to contain additional blocks or attributes not specified in the
	// schema. If any are present, the returned Body is non-nil and contains
	// the remaining items from the body that were not selected by the schema.
	PartialContent(schema *BodySchema) (*BodyContent, Body, Diagnostics)
}

// BodyContent is the result of applying a BodySchema to a Body.
type BodyContent struct {
	Attributes map[string]Attribute
	Blocks     Blocks
}

// Attribute represents an attribute from within a body.
type Attribute struct {
	Name string
	Expr Expression

	Range     Range
	NameRange Range
	ExprRange Range
}

// Expression is a literal value or an expression provided in the
// configuration, which can be evaluated within a scope to produce a value.
type Expression interface {
	LiteralValue() cty.Value
	// TODO: evaluation of non-literal expressions
}

// OfType filters the receiving block sequence by block type name,
// returning a new block sequence including only the blocks of the
// requested type.
func (els Blocks) OfType(typeName string) Blocks {
	ret := make(Blocks, 0)
	for _, el := range els {
		if el.Type == typeName {
			ret = append(ret, el)
		}
	}
	return ret
}

// ByType transforms the receiving block sequence into a map from type
// name to block sequences of only that type.
func (els Blocks) ByType() map[string]Blocks {
	ret := make(map[string]Blocks)
	for _, el := range els {
		ty := el.Type
		if ret[ty] == nil {
			ret[ty] = make(Blocks, 0, 1)
		}
		ret[ty] = append(ret[ty], el)
	}
	return ret
}
