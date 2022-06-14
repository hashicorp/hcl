package hclsyntax

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

type Operation struct {
	Impl function.Function
	Type cty.Type

	// ShortCircuit is an optional callback for binary operations which, if
	// set, will be called with the result of evaluating the LHS expression.
	//
	// ShortCircuit may return cty.NilVal to allow evaluation to proceed
	// as normal, or it may return a non-nil value to force the operation
	// to return that value and perform only type checking on the RHS
	// expression, as opposed to full evaluation.
	ShortCircuit func(lhs cty.Value) cty.Value
}

var (
	OpLogicalOr = &Operation{
		Impl: stdlib.OrFunc,
		Type: cty.Bool,

		ShortCircuit: func(lhs cty.Value) cty.Value {
			if lhs.RawEquals(cty.True) {
				return cty.True
			}
			return cty.NilVal
		},
	}
	OpLogicalAnd = &Operation{
		Impl: stdlib.AndFunc,
		Type: cty.Bool,

		ShortCircuit: func(lhs cty.Value) cty.Value {
			if lhs.RawEquals(cty.False) {
				return cty.False
			}
			return cty.NilVal
		},
	}
	OpLogicalNot = &Operation{
		Impl: stdlib.NotFunc,
		Type: cty.Bool,
	}

	OpEqual = &Operation{
		Impl: stdlib.EqualFunc,
		Type: cty.Bool,
	}
	OpNotEqual = &Operation{
		Impl: stdlib.NotEqualFunc,
		Type: cty.Bool,
	}

	OpGreaterThan = &Operation{
		Impl: stdlib.GreaterThanFunc,
		Type: cty.Bool,
	}
	OpGreaterThanOrEqual = &Operation{
		Impl: stdlib.GreaterThanOrEqualToFunc,
		Type: cty.Bool,
	}
	OpLessThan = &Operation{
		Impl: stdlib.LessThanFunc,
		Type: cty.Bool,
	}
	OpLessThanOrEqual = &Operation{
		Impl: stdlib.LessThanOrEqualToFunc,
		Type: cty.Bool,
	}

	OpAdd = &Operation{
		Impl: stdlib.AddFunc,
		Type: cty.Number,
	}
	OpSubtract = &Operation{
		Impl: stdlib.SubtractFunc,
		Type: cty.Number,
	}
	OpMultiply = &Operation{
		Impl: stdlib.MultiplyFunc,
		Type: cty.Number,
	}
	OpDivide = &Operation{
		Impl: stdlib.DivideFunc,
		Type: cty.Number,
	}
	OpModulo = &Operation{
		Impl: stdlib.ModuloFunc,
		Type: cty.Number,
	}
	OpNegate = &Operation{
		Impl: stdlib.NegateFunc,
		Type: cty.Number,
	}
)

var binaryOps []map[TokenType]*Operation
var rightAssociativeBinaryOps = map[TokenType]struct{}{
	TokenOr:  {},
	TokenAnd: {},
}

func init() {
	// This operation table maps from the operator's token type
	// to the AST operation type. All expressions produced from
	// binary operators are BinaryOp nodes.
	//
	// Binary operator groups are listed in order of precedence, with
	// the *lowest* precedence first. Operators within the same group
	// have left-to-right associativity.
	binaryOps = []map[TokenType]*Operation{
		{
			TokenOr: OpLogicalOr,
		},
		{
			TokenAnd: OpLogicalAnd,
		},
		{
			TokenEqualOp:  OpEqual,
			TokenNotEqual: OpNotEqual,
		},
		{
			TokenGreaterThan:   OpGreaterThan,
			TokenGreaterThanEq: OpGreaterThanOrEqual,
			TokenLessThan:      OpLessThan,
			TokenLessThanEq:    OpLessThanOrEqual,
		},
		{
			TokenPlus:  OpAdd,
			TokenMinus: OpSubtract,
		},
		{
			TokenStar:    OpMultiply,
			TokenSlash:   OpDivide,
			TokenPercent: OpModulo,
		},
	}
}

type BinaryOpExpr struct {
	LHS Expression
	Op  *Operation
	RHS Expression

	SrcRange hcl.Range
}

func (e *BinaryOpExpr) walkChildNodes(w internalWalkFunc) {
	w(e.LHS)
	w(e.RHS)
}

func (e *BinaryOpExpr) Value(ctx *hcl.EvalContext) (cty.Value, hcl.Diagnostics) {
	impl := e.Op.Impl // assumed to be a function taking exactly two arguments
	params := impl.Params()
	lhsParam := params[0]
	rhsParam := params[1]

	var diags hcl.Diagnostics

	givenLHSVal, lhsDiags := e.LHS.Value(ctx)
	diags = append(diags, lhsDiags...)
	lhsVal, err := convert.Convert(givenLHSVal, lhsParam.Type)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity:    hcl.DiagError,
			Summary:     "Invalid operand",
			Detail:      fmt.Sprintf("Unsuitable value for left operand: %s.", err),
			Subject:     e.LHS.Range().Ptr(),
			Context:     &e.SrcRange,
			Expression:  e.LHS,
			EvalContext: ctx,
		})
	}

	// If this is a short-circuiting operator and the LHS produces a
	// short-circuiting result then we'll evaluate the RHS only for type
	// checking purposes, ignoring any specific values, as a compromise
	// between the convenience of a total short-circuit behavior and the
	// benefit of not masking type errors on the RHS that we could still
	// give earlier feedback about.
	var forceResult cty.Value
	rhsCtx := ctx
	if e.Op.ShortCircuit != nil {
		if !givenLHSVal.IsKnown() {
			// If this is a short-circuit operator and our LHS value is
			// unknown then we can't predict whether we would short-circuit
			// yet, and so we must proceed under the assumption that we _will_
			// short-circuit to avoid raising any errors on the RHS that would
			// eventually be hidden by the short-circuit behavior once LHS
			// becomes known.
			forceResult = cty.UnknownVal(e.Op.Type)
			rhsCtx = ctx.NewChildAllVariablesUnknown()
		} else if forceResult = e.Op.ShortCircuit(givenLHSVal); forceResult != cty.NilVal {
			// This ensures that we'll only be type-checking against any
			// variables used on the RHS, while not raising any errors about
			// their values.
			rhsCtx = ctx.NewChildAllVariablesUnknown()
		}
	}

	givenRHSVal, rhsDiags := e.RHS.Value(rhsCtx)
	diags = append(diags, rhsDiags...)
	rhsVal, err := convert.Convert(givenRHSVal, rhsParam.Type)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity:    hcl.DiagError,
			Summary:     "Invalid operand",
			Detail:      fmt.Sprintf("Unsuitable value for right operand: %s.", err),
			Subject:     e.RHS.Range().Ptr(),
			Context:     &e.SrcRange,
			Expression:  e.RHS,
			EvalContext: ctx,
		})
	}

	if diags.HasErrors() {
		// Don't actually try the call if we have errors already, since the
		// this will probably just produce a confusing duplicative diagnostic.
		return cty.UnknownVal(e.Op.Type), diags
	}

	// If we short-circuited above and still passed the type-check of RHS then
	// we'll halt here and return the short-circuit result rather than actually
	// executing the opertion.
	if forceResult != cty.NilVal {
		return forceResult, diags
	}

	args := []cty.Value{lhsVal, rhsVal}
	result, err := impl.Call(args)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			// FIXME: This diagnostic is useless.
			Severity:    hcl.DiagError,
			Summary:     "Operation failed",
			Detail:      fmt.Sprintf("Error during operation: %s.", err),
			Subject:     &e.SrcRange,
			Expression:  e,
			EvalContext: ctx,
		})
		return cty.UnknownVal(e.Op.Type), diags
	}

	return result, diags
}

func (e *BinaryOpExpr) Range() hcl.Range {
	return e.SrcRange
}

func (e *BinaryOpExpr) StartRange() hcl.Range {
	return e.LHS.StartRange()
}

type UnaryOpExpr struct {
	Op  *Operation
	Val Expression

	SrcRange    hcl.Range
	SymbolRange hcl.Range
}

func (e *UnaryOpExpr) walkChildNodes(w internalWalkFunc) {
	w(e.Val)
}

func (e *UnaryOpExpr) Value(ctx *hcl.EvalContext) (cty.Value, hcl.Diagnostics) {
	impl := e.Op.Impl // assumed to be a function taking exactly one argument
	params := impl.Params()
	param := params[0]

	givenVal, diags := e.Val.Value(ctx)

	val, err := convert.Convert(givenVal, param.Type)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity:    hcl.DiagError,
			Summary:     "Invalid operand",
			Detail:      fmt.Sprintf("Unsuitable value for unary operand: %s.", err),
			Subject:     e.Val.Range().Ptr(),
			Context:     &e.SrcRange,
			Expression:  e.Val,
			EvalContext: ctx,
		})
	}

	if diags.HasErrors() {
		// Don't actually try the call if we have errors already, since the
		// this will probably just produce a confusing duplicative diagnostic.
		return cty.UnknownVal(e.Op.Type), diags
	}

	args := []cty.Value{val}
	result, err := impl.Call(args)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			// FIXME: This diagnostic is useless.
			Severity:    hcl.DiagError,
			Summary:     "Operation failed",
			Detail:      fmt.Sprintf("Error during operation: %s.", err),
			Subject:     &e.SrcRange,
			Expression:  e,
			EvalContext: ctx,
		})
		return cty.UnknownVal(e.Op.Type), diags
	}

	return result, diags
}

func (e *UnaryOpExpr) Range() hcl.Range {
	return e.SrcRange
}

func (e *UnaryOpExpr) StartRange() hcl.Range {
	return e.SymbolRange
}
