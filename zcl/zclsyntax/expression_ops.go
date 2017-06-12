package zclsyntax

import (
	"fmt"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
	"github.com/zclconf/go-zcl/zcl"
)

type Operation rune

const (
	OpNil Operation = 0 // Zero value of Operation. Not a valid Operation.

	OpLogicalOr          Operation = '∨'
	OpLogicalAnd         Operation = '∧'
	OpLogicalNot         Operation = '!'
	OpEqual              Operation = '='
	OpNotEqual           Operation = '≠'
	OpGreaterThan        Operation = '>'
	OpGreaterThanOrEqual Operation = '≥'
	OpLessThan           Operation = '<'
	OpLessThanOrEqual    Operation = '≤'
	OpAdd                Operation = '+'
	OpSubtract           Operation = '-'
	OpMultiply           Operation = '*'
	OpDivide             Operation = '/'
	OpModulo             Operation = '%'
	OpNegate             Operation = '∓'
)

var binaryOps []map[TokenType]Operation

func init() {
	// This operation table maps from the operator's token type
	// to the AST operation type. All expressions produced from
	// binary operators are BinaryOp nodes.
	//
	// Binary operator groups are listed in order of precedence, with
	// the *lowest* precedence first. Operators within the same group
	// have left-to-right associativity.
	binaryOps = []map[TokenType]Operation{
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

var operationImpls = map[Operation]function.Function{
	OpLogicalAnd: stdlib.AndFunc,
	OpLogicalOr:  stdlib.OrFunc,
	OpLogicalNot: stdlib.NotFunc,

	OpEqual:    stdlib.EqualFunc,
	OpNotEqual: stdlib.NotEqualFunc,

	OpGreaterThan:        stdlib.GreaterThanFunc,
	OpGreaterThanOrEqual: stdlib.GreaterThanOrEqualToFunc,
	OpLessThan:           stdlib.LessThanFunc,
	OpLessThanOrEqual:    stdlib.LessThanOrEqualToFunc,

	OpAdd:      stdlib.AddFunc,
	OpSubtract: stdlib.SubtractFunc,
	OpMultiply: stdlib.MultiplyFunc,
	OpDivide:   stdlib.DivideFunc,
	OpModulo:   stdlib.ModuloFunc,
	OpNegate:   stdlib.NegateFunc,
}

type BinaryOpExpr struct {
	LHS Expression
	Op  Operation
	RHS Expression

	SrcRange zcl.Range
}

func (e *BinaryOpExpr) walkChildNodes(w internalWalkFunc) {
	e.LHS = w(e.LHS).(Expression)
	e.RHS = w(e.LHS).(Expression)
}

func (e *BinaryOpExpr) Value(ctx *zcl.EvalContext) (cty.Value, zcl.Diagnostics) {
	impl := operationImpls[e.Op] // assumed to be a function taking exactly two arguments
	params := impl.Params()
	lhsParam := params[0]
	rhsParam := params[1]

	var diags zcl.Diagnostics

	givenLHSVal, lhsDiags := e.LHS.Value(ctx)
	givenRHSVal, rhsDiags := e.RHS.Value(ctx)
	diags = append(diags, lhsDiags...)
	diags = append(diags, rhsDiags...)

	lhsVal, err := convert.Convert(givenLHSVal, lhsParam.Type)
	if err != nil {
		diags = append(diags, &zcl.Diagnostic{
			Severity: zcl.DiagError,
			Summary:  "Invalid operand",
			Detail:   fmt.Sprintf("Unsuitable value for left operand: %s.", err),
			Subject:  e.LHS.Range().Ptr(),
			Context:  &e.SrcRange,
		})
	}
	rhsVal, err := convert.Convert(givenRHSVal, rhsParam.Type)
	if err != nil {
		diags = append(diags, &zcl.Diagnostic{
			Severity: zcl.DiagError,
			Summary:  "Invalid operand",
			Detail:   fmt.Sprintf("Unsuitable value for right operand: %s.", err),
			Subject:  e.RHS.Range().Ptr(),
			Context:  &e.SrcRange,
		})
	}

	if diags.HasErrors() {
		// Don't actually try the call if we have errors already, since the
		// this will probably just produce a confusing duplicative diagnostic.
		// Instead, we'll use the function's type check to figure out what
		// type would be returned, if possible.
		args := []cty.Value{givenLHSVal, givenRHSVal}
		retType, err := impl.ReturnTypeForValues(args)
		if err != nil {
			// can't even get a return type, so we'll bail here.
			return cty.DynamicVal, diags
		}

		return cty.UnknownVal(retType), diags
	}

	args := []cty.Value{lhsVal, rhsVal}
	result, err := impl.Call(args)
	if err != nil {
		diags = append(diags, &zcl.Diagnostic{
			// FIXME: This diagnostic is useless.
			Severity: zcl.DiagError,
			Summary:  "Operation failed",
			Detail:   fmt.Sprintf("Error during operation: %s.", err),
			Subject:  e.RHS.Range().Ptr(),
			Context:  &e.SrcRange,
		})
		retType, err := impl.ReturnTypeForValues(args)
		if err != nil {
			return cty.DynamicVal, diags
		}
		return cty.UnknownVal(retType), diags
	}

	return result, diags
}

func (e *BinaryOpExpr) Range() zcl.Range {
	return e.SrcRange
}

func (e *BinaryOpExpr) StartRange() zcl.Range {
	return e.LHS.StartRange()
}

type UnaryOpExpr struct {
	Op  Operation
	Val Expression

	SrcRange    zcl.Range
	SymbolRange zcl.Range
}

func (e *UnaryOpExpr) walkChildNodes(w internalWalkFunc) {
	e.Val = w(e.Val).(Expression)
}

func (e *UnaryOpExpr) Value(ctx *zcl.EvalContext) (cty.Value, zcl.Diagnostics) {
	panic("UnaryOpExpr.Value not yet implemented")
}

func (e *UnaryOpExpr) Range() zcl.Range {
	return e.SrcRange
}

func (e *UnaryOpExpr) StartRange() zcl.Range {
	return e.SymbolRange
}
