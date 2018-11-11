package hclpack

import (
	"fmt"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// Expression is an implementation of hcl.Expression in terms of some raw
// expression source code. The methods of this type will first parse the
// source code and then pass the call through to the real expression that
// is produced.
type Expression struct {
	// Source is the raw source code of the expression, which should be parsed
	// as the syntax specified by SourceType.
	Source     []byte
	SourceType ExprSourceType

	// Range_ and StartRange_ describe the physical extents of the expression
	// in the original source code. SourceRange_ is its entire range while
	// StartRange is just the tokens that introduce the expression type. For
	// simple expression types, SourceRange and StartRange are identical.
	Range_, StartRange_ hcl.Range
}

var _ hcl.Expression = (*Expression)(nil)

// Value implements the Value method of hcl.Expression but with the additional
// step of first parsing the expression source code. This implementation is
// unusual in that it can potentially return syntax errors, whereas other
// Value implementations usually work with already-parsed expressions.
func (e *Expression) Value(ctx *hcl.EvalContext) (cty.Value, hcl.Diagnostics) {
	expr, diags := e.Parse()
	if diags.HasErrors() {
		return cty.DynamicVal, diags
	}

	val, moreDiags := expr.Value(ctx)
	diags = append(diags, moreDiags...)
	return val, diags
}

// Variables implements the Variables method of hcl.Expression but with the
// additional step of first parsing the expression source code.
//
// Since this method cannot return errors, it will return a nil slice if
// parsing fails, indicating that no variables are present. This is okay in
// practice because a subsequent call to Value would fail with syntax errors
// regardless of what variables are in the context.
func (e *Expression) Variables() []hcl.Traversal {
	expr, diags := e.Parse()
	if diags.HasErrors() {
		return nil
	}
	return expr.Variables()
}

func (e *Expression) Range() hcl.Range {
	return e.Range_
}

func (e *Expression) StartRange() hcl.Range {
	return e.StartRange_
}

// Parse attempts to parse the source code of the receiving expression using
// its indicated source type, returning the expression if possible and any
// diagnostics produced during parsing.
func (e *Expression) Parse() (hcl.Expression, hcl.Diagnostics) {
	switch e.SourceType {
	case ExprNative:
		return hclsyntax.ParseExpression(e.Source, e.Range_.Filename, e.Range_.Start)
	case ExprTemplate:
		return hclsyntax.ParseTemplate(e.Source, e.Range_.Filename, e.Range_.Start)
	default:
		// This should never happen for a valid Expression.
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Invalid expression source type",
				Detail:   fmt.Sprintf("Packed version of this expression has an invalid source type %s. This is always a bug.", e.SourceType),
				Subject:  &e.Range_,
			},
		}
	}
}

func (e *Expression) addRanges(rngs map[hcl.Range]struct{}) {
	rngs[e.Range_] = struct{}{}
	rngs[e.StartRange_] = struct{}{}
}

// ExprSourceType defines the syntax type used for an expression's source code,
// which is then used to select a suitable parser for it when evaluating.
type ExprSourceType rune

//go:generate stringer -type ExprSourceType

const (
	// ExprNative indicates that an expression must be parsed as native
	// expression syntax, with hclsyntax.ParseExpression.
	ExprNative ExprSourceType = 'N'

	// ExprTemplate indicates that an expression must be parsed as nave
	// template syntax, with hclsyntax.ParseTemplate.
	ExprTemplate ExprSourceType = 'T'
)
