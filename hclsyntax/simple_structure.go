// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package hclsyntax

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
)

type AttributeKind int

const (
	// AttributeKindOther indicates the expression is neither literal nor traversal (e.g. a function call).
	AttributeKindOther AttributeKind = iota
	// AttributeKindLiteral indicates the expression is a literal value.
	AttributeKindLiteral
	// AttributeKindTraversal indicates the expression is a traversal (e.g. a variable reference).
	AttributeKindTraversal
)

// SimpleAttribute describes the source expression that an attribute value is assigned from,
// along with any traversal information if the expression is a single absolute traversal.
type SimpleAttribute struct {
	hcl.Attribute
	// Traversal is set only when the expression is a single absolute traversal.
	Traversal hcl.Traversal

	// Kind indicates the type of expression.
	Kind AttributeKind
}

func (a *SimpleAttribute) IsLiteral() bool {
	return a.Kind == AttributeKindLiteral
}

func (a *SimpleAttribute) IsTraversal() bool {
	return a.Kind == AttributeKindTraversal
}

// SimpleBody contains expression summaries for all attributes in a body,
// along with recursively-discovered nested blocks.
// TODO: Make SimpleBody implement hcl.Body?.
type SimpleBody struct {
	src        hcl.Body
	Attributes map[string]SimpleAttribute
	Blocks     []SimpleBlock
}

// SimpleBlock contains the recursive expression summary for a single
// block instance.
type SimpleBlock struct {
	hcl.Block
	Body *SimpleBody
}

// ParseSimpleBody recursively analyzes all of the attributes in the given
// body and its nested blocks.
// The returned body contains only attributes that are either
// simple traversals or literal values.
func ParseSimpleBody(body hcl.Body) (*SimpleBody, hcl.Diagnostics) {
	syntaxBody, ok := body.(*Body)
	if !ok {
		return nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unsupported body type",
				Detail:   fmt.Sprintf("ParseSimpleBody requires a body produced by the native HCL syntax parser, not %T.", body),
				Subject:  body.MissingItemRange().Ptr(),
			},
		}
	}

	return parseSimpleBody(syntaxBody)
}

func parseSimpleBody(body *Body) (*SimpleBody, hcl.Diagnostics) {
	ret := &SimpleBody{src: body}

	if len(body.Attributes) != 0 {
		ret.Attributes = make(map[string]SimpleAttribute, len(body.Attributes))
		for name, attr := range body.Attributes {
			attrExpr, diags := parseSimpleAttribute(attr)
			if diags.HasErrors() {
				return ret, diags
			}
			ret.Attributes[name] = attrExpr
		}
	}

	if len(body.Blocks) != 0 {
		ret.Blocks = make([]SimpleBlock, 0, len(body.Blocks))
		for _, block := range body.Blocks {
			bodyExpr, diags := parseSimpleBody(block.Body)
			if diags.HasErrors() {
				return ret, diags
			}
			ret.Blocks = append(ret.Blocks, SimpleBlock{
				Block: *block.AsHCLBlock(),
				Body:  bodyExpr,
			})
		}
	}

	return ret, nil
}

func parseSimpleAttribute(attr *Attribute) (SimpleAttribute, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	rawExpr := hcl.UnwrapExpression(attr.Expr)
	ret := SimpleAttribute{
		Attribute: *attr.AsHCLAttribute(),
	}

	syntaxExpr, ok := rawExpr.(Expression)
	if ok {
		if isLiteralExpression(syntaxExpr) {
			ret.Kind = AttributeKindLiteral
			return ret, nil
		}
	}

	ret.Traversal, diags = hcl.AbsTraversalForExpr(rawExpr)
	// ignore errors, assume the expression is not a simple traversal
	if diags.HasErrors() {
		return ret, nil
	}
	ret.Kind = AttributeKindTraversal
	return ret, nil
}

func isLiteralExpression(expr Expression) bool {
	switch expr := expr.(type) {
	case *ParenthesesExpr:
		return isLiteralExpression(expr.Expression)

	case *LiteralValueExpr:
		return true

	case *TemplateExpr:
		return expr.IsStringLiteral()

	case *TupleConsExpr:
		for _, itemExpr := range expr.Exprs {
			if !isLiteralExpression(itemExpr) {
				return false
			}
		}
		return true

	case *ObjectConsExpr:
		for _, item := range expr.Items {
			keyExpr, ok := item.KeyExpr.(*ObjectConsKeyExpr)
			if !ok {
				// This should never happen, as the key expression is always a *ObjectConsKeyExpr
				// in native HCL syntax.
				panic("unexpected key expression type")
			}

			if _, diags := keyExpr.Value(nil); diags.HasErrors() {
				return false
			}
			if !isLiteralExpression(item.ValueExpr) {
				return false
			}
		}
		return true

	default:
		return false
	}
}
