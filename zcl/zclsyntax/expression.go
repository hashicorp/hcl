package zclsyntax

import (
	"fmt"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-zcl/zcl"
)

// Expression is the abstract type for nodes that behave as zcl expressions.
type Expression interface {
	Node

	// The zcl.Expression methods are duplicated here, rather than simply
	// embedded, because both Node and zcl.Expression have a Range method
	// and so they conflict.

	Value(ctx *zcl.EvalContext) (cty.Value, zcl.Diagnostics)
	Variables() []zcl.Traversal
	StartRange() zcl.Range
}

// Assert that Expression implements zcl.Expression
var assertExprImplExpr zcl.Expression = Expression(nil)

// LiteralValueExpr is an expression that just always returns a given value.
type LiteralValueExpr struct {
	Val      cty.Value
	SrcRange zcl.Range
}

func (e *LiteralValueExpr) walkChildNodes(w internalWalkFunc) {
	// Literal values have no child nodes
}

func (e *LiteralValueExpr) Value(ctx *zcl.EvalContext) (cty.Value, zcl.Diagnostics) {
	return e.Val, nil
}

func (e *LiteralValueExpr) Range() zcl.Range {
	return e.SrcRange
}

func (e *LiteralValueExpr) StartRange() zcl.Range {
	return e.SrcRange
}

// ScopeTraversalExpr is an Expression that retrieves a value from the scope
// using a traversal.
type ScopeTraversalExpr struct {
	Traversal zcl.Traversal
	SrcRange  zcl.Range
}

func (e *ScopeTraversalExpr) walkChildNodes(w internalWalkFunc) {
	// Scope traversals have no child nodes
}

func (e *ScopeTraversalExpr) Value(ctx *zcl.EvalContext) (cty.Value, zcl.Diagnostics) {
	panic("ScopeTraversalExpr.Value not yet implemented")
}

func (e *ScopeTraversalExpr) Range() zcl.Range {
	return e.SrcRange
}

func (e *ScopeTraversalExpr) StartRange() zcl.Range {
	return e.SrcRange
}

// FunctionCallExpr is an Expression that calls a function from the EvalContext
// and returns its result.
type FunctionCallExpr struct {
	Name string
	Args []Expression

	NameRange       zcl.Range
	OpenParenRange  zcl.Range
	CloseParenRange zcl.Range
}

func (e *FunctionCallExpr) walkChildNodes(w internalWalkFunc) {
	for i, arg := range e.Args {
		e.Args[i] = w(arg).(Expression)
	}
}

func (e *FunctionCallExpr) Value(ctx *zcl.EvalContext) (cty.Value, zcl.Diagnostics) {
	var diags zcl.Diagnostics

	f, exists := ctx.Functions[e.Name]
	if !exists {
		avail := make([]string, 0, len(ctx.Functions))
		for name := range ctx.Functions {
			avail = append(avail, name)
		}
		suggestion := nameSuggestion(e.Name, avail)
		if suggestion != "" {
			suggestion = fmt.Sprintf(" Did you mean %q?", suggestion)
		}

		return cty.DynamicVal, zcl.Diagnostics{
			{
				Severity: zcl.DiagError,
				Summary:  "Call to unknown function",
				Detail:   fmt.Sprintf("There is no function named %q.%s", e.Name, suggestion),
				Subject:  &e.NameRange,
				Context:  e.Range().Ptr(),
			},
		}
	}

	params := f.Params()
	varParam := f.VarParam()

	if len(e.Args) < len(params) {
		missing := params[len(e.Args)]
		qual := ""
		if varParam != nil {
			qual = " at least"
		}
		return cty.DynamicVal, zcl.Diagnostics{
			{
				Severity: zcl.DiagError,
				Summary:  "Not enough function arguments",
				Detail: fmt.Sprintf(
					"Function %q expects%s %d argument(s). Missing value for %q.",
					e.Name, qual, len(params), missing.Name,
				),
				Subject: &e.CloseParenRange,
				Context: e.Range().Ptr(),
			},
		}
	}

	if varParam == nil && len(e.Args) > len(params) {
		return cty.DynamicVal, zcl.Diagnostics{
			{
				Severity: zcl.DiagError,
				Summary:  "Too many function arguments",
				Detail: fmt.Sprintf(
					"Function %q expects only %d argument(s).",
					e.Name, len(params),
				),
				Subject: e.Args[len(params)].StartRange().Ptr(),
				Context: e.Range().Ptr(),
			},
		}
	}

	argVals := make([]cty.Value, len(e.Args))

	for i, argExpr := range e.Args {
		var param *function.Parameter
		if i < len(params) {
			param = &params[i]
		} else {
			param = varParam
		}

		val, argDiags := argExpr.Value(ctx)
		if len(argDiags) > 0 {
			diags = append(diags, argDiags...)
		}

		// Try to convert our value to the parameter type
		val, err := convert.Convert(val, param.Type)
		if err != nil {
			diags = append(diags, &zcl.Diagnostic{
				Severity: zcl.DiagError,
				Summary:  "Invalid function argument",
				Detail: fmt.Sprintf(
					"Invalid value for %q parameter: %s.",
					param.Name, err,
				),
				Subject: argExpr.StartRange().Ptr(),
				Context: e.Range().Ptr(),
			})
		}

		argVals[i] = val
	}

	if diags.HasErrors() {
		// Don't try to execute the function if we already have errors with
		// the arguments, because the result will probably be a confusing
		// error message.
		return cty.DynamicVal, diags
	}

	resultVal, err := f.Call(argVals)
	if err != nil {
		switch terr := err.(type) {
		case function.ArgError:
			i := terr.Index
			var param *function.Parameter
			if i < len(params) {
				param = &params[i]
			} else {
				param = varParam
			}
			argExpr := e.Args[i]

			// TODO: we should also unpick a PathError here and show the
			// path to the deep value where the error was detected.
			diags = append(diags, &zcl.Diagnostic{
				Severity: zcl.DiagError,
				Summary:  "Invalid function argument",
				Detail: fmt.Sprintf(
					"Invalid value for %q parameter: %s.",
					param.Name, err,
				),
				Subject: argExpr.StartRange().Ptr(),
				Context: e.Range().Ptr(),
			})

		default:
			diags = append(diags, &zcl.Diagnostic{
				Severity: zcl.DiagError,
				Summary:  "Error in function call",
				Detail: fmt.Sprintf(
					"Call to function %q failed: %s.",
					e.Name, err,
				),
				Subject: e.StartRange().Ptr(),
				Context: e.Range().Ptr(),
			})
		}

		return cty.DynamicVal, diags
	}

	return resultVal, diags
}

func (e *FunctionCallExpr) Range() zcl.Range {
	return zcl.RangeBetween(e.NameRange, e.CloseParenRange)
}

func (e *FunctionCallExpr) StartRange() zcl.Range {
	return zcl.RangeBetween(e.NameRange, e.OpenParenRange)
}
