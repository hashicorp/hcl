package hcl

import (
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
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

// Parent returns the parent of the receiver, or nil if the receiver has
// no parent.
func (ctx *EvalContext) Parent() *EvalContext {
	return ctx.parent
}

// NewChildAllVariablesUnknown is an extension of NewChild which, in addition
// to creating a child context, also pre-populates its Variables table
// with variable definitions masking every variable define in the reciever
// and its ancestors with an unknown value of the same type as the original.
//
// The child does not initially have any of its own functions defined, and so
// it can still inherit any defined functions from the reciever.
//
// Because is function effectively takes a snapshot of the variables as they
// are defined at the time of the call, it is incorrect to subsequently
// modify the variables in any of the ancestor contexts in a way that would
// change which variables are defined or what value types they each have.
//
// This is a specialized helper function intended to support type-checking
// use-cases, where the goal is only to check whether values are being used
// in a way that makes sense for their types while not reacting to their
// actual values.
func (ctx *EvalContext) NewChildAllVariablesUnknown() *EvalContext {
	ret := ctx.NewChild()
	ret.Variables = make(map[string]cty.Value)

	currentAncestor := ctx
	for currentAncestor != nil {
		for name, val := range currentAncestor.Variables {
			if _, ok := ret.Variables[name]; !ok {
				ret.Variables[name] = cty.UnknownVal(val.Type())
			}
		}
		currentAncestor = currentAncestor.parent
	}

	return ret
}
