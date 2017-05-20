package zcl

import (
	"github.com/apparentlymart/go-cty/cty"
	"github.com/apparentlymart/go-cty/cty/function"
)

// An EvalContext provides the variables and functions that should be used
// to evaluate an expression.
type EvalContext struct {
	Variables map[string]cty.Value
	Functions map[string]function.Function
	parent    *EvalContext
}

// NewChild returns a new EvalContext that is a child of the receiver.
func (ctx *EvalContext) NewChild() *EvalContext {
	return &EvalContext{parent: ctx}
}
