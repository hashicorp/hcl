package zclsyntax

import (
	"github.com/zclconf/go-cty/cty"
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
			TokenEqual:    OpEqual,
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
	Op  Operation
	RHS Expression

	SrcRange zcl.Range
}

func (e *BinaryOpExpr) walkChildNodes(w internalWalkFunc) {
	e.LHS = w(e.LHS).(Expression)
	e.RHS = w(e.LHS).(Expression)
}

func (e *BinaryOpExpr) Value(ctx *zcl.EvalContext) (cty.Value, zcl.Diagnostics) {
	panic("BinaryOpExpr.Value not yet implemented")
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
