package zcl

import (
	"github.com/apparentlymart/go-cty/cty"
)

// File is the top-level node that results from parsing a ZCL file.
type File struct {
	Body Body
}

// Element represents a nested block within a Body.
type Element struct {
	Type   string
	Labels []string
	Body   Body

	DefRange    Range   // Range that can be considered the "definition" for seeking in an editor
	TypeRange   Range   // Range for the element type declaration specifically.
	LabelRanges []Range // Ranges for the label values specifically.
}

// Elements is a sequence of Element.
type Elements []*Element

// Body is a container for attributes and elements. It serves as the primary
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
	// to contain additional elements or attributes not specified in the
	// schema. If any are present, the returned Body is non-nil and contains
	// the remaining items from the body that were not selected by the schema.
	PartialContent(schema *BodySchema) (*BodyContent, Body, Diagnostics)
}

// BodyContent is the result of applying a BodySchema to a Body.
type BodyContent struct {
	Attributes map[string]Attribute
	Elements   Elements
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

// OfType filters the receiving element sequence by element type name,
// returning a new element sequence including only the elements of the
// requested type.
func (els Elements) OfType(typeName string) Elements {
	ret := make(Elements, 0)
	for _, el := range els {
		if el.Type == typeName {
			ret = append(ret, el)
		}
	}
	return ret
}

// ByType transforms the receiving elements sequence into a map from type
// name to element sequences of only that type.
func (els Elements) ByType() map[string]Elements {
	ret := make(map[string]Elements)
	for _, el := range els {
		ty := el.Type
		if ret[ty] == nil {
			ret[ty] = make(Elements, 0, 1)
		}
		ret[ty] = append(ret[ty], el)
	}
	return ret
}
