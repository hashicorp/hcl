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

// RelativeTraversalExpr is an Expression that retrieves a value from another
// value using a _relative_ traversal.
type RelativeTraversalExpr struct {
	Source    Expression
	Traversal zcl.Traversal
	SrcRange  zcl.Range
}

func (e *RelativeTraversalExpr) walkChildNodes(w internalWalkFunc) {
	// Scope traversals have no child nodes
}

func (e *RelativeTraversalExpr) Value(ctx *zcl.EvalContext) (cty.Value, zcl.Diagnostics) {
	src, diags := e.Source.Value(ctx)
	ret, travDiags := e.Traversal.TraverseRel(src)
	diags = append(diags, travDiags...)
	return ret, diags
}

func (e *RelativeTraversalExpr) Range() zcl.Range {
	return e.SrcRange
}

func (e *RelativeTraversalExpr) StartRange() zcl.Range {
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

	if ctx == nil || ctx.Functions == nil {
		return cty.DynamicVal, zcl.Diagnostics{
			{
				Severity: zcl.DiagError,
				Summary:  "Function calls not allowed",
				Detail:   "Functions may not be called here.",
				Subject:  &e.NameRange,
				Context:  e.Range().Ptr(),
			},
		}
	}

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

type ConditionalExpr struct {
	Condition   Expression
	TrueResult  Expression
	FalseResult Expression

	SrcRange zcl.Range
}

func (e *ConditionalExpr) walkChildNodes(w internalWalkFunc) {
	e.Condition = w(e.Condition).(Expression)
	e.TrueResult = w(e.TrueResult).(Expression)
	e.FalseResult = w(e.FalseResult).(Expression)
}

func (e *ConditionalExpr) Value(ctx *zcl.EvalContext) (cty.Value, zcl.Diagnostics) {
	trueResult, trueDiags := e.TrueResult.Value(ctx)
	falseResult, falseDiags := e.FalseResult.Value(ctx)
	var diags zcl.Diagnostics

	// Try to find a type that both results can be converted to.
	resultType, convs := convert.UnifyUnsafe([]cty.Type{trueResult.Type(), falseResult.Type()})
	if resultType == cty.NilType {
		return cty.DynamicVal, zcl.Diagnostics{
			{
				Severity: zcl.DiagError,
				Summary:  "Inconsistent conditional result types",
				Detail: fmt.Sprintf(
					// FIXME: Need a helper function for showing natural-language type diffs,
					// since this will generate some useless messages in some cases, like
					// "These expressions are object and object respectively" if the
					// object types don't exactly match.
					"The true and false result expressions must have consistent types. The given expressions are %s and %s, respectively.",
					trueResult.Type(), falseResult.Type(),
				),
				Subject: zcl.RangeBetween(e.TrueResult.Range(), e.FalseResult.Range()).Ptr(),
				Context: &e.SrcRange,
			},
		}
	}

	condResult, condDiags := e.Condition.Value(ctx)
	diags = append(diags, condDiags...)
	if condResult.IsNull() {
		diags = append(diags, &zcl.Diagnostic{
			Severity: zcl.DiagError,
			Summary:  "Null condition",
			Detail:   "The condition value is null. Conditions must either be true or false.",
			Subject:  e.Condition.Range().Ptr(),
			Context:  &e.SrcRange,
		})
		return cty.UnknownVal(resultType), diags
	}
	if !condResult.IsKnown() {
		return cty.UnknownVal(resultType), diags
	}
	condResult, err := convert.Convert(condResult, cty.Bool)
	if err != nil {
		diags = append(diags, &zcl.Diagnostic{
			Severity: zcl.DiagError,
			Summary:  "Incorrect condition type",
			Detail:   fmt.Sprintf("The condition expression must be of type bool."),
			Subject:  e.Condition.Range().Ptr(),
			Context:  &e.SrcRange,
		})
		return cty.UnknownVal(resultType), diags
	}

	if condResult.True() {
		diags = append(diags, trueDiags...)
		if convs[0] != nil {
			var err error
			trueResult, err = convs[0](trueResult)
			if err != nil {
				// Unsafe conversion failed with the concrete result value
				diags = append(diags, &zcl.Diagnostic{
					Severity: zcl.DiagError,
					Summary:  "Inconsistent conditional result types",
					Detail: fmt.Sprintf(
						"The true result value has the wrong type: %s.",
						err.Error(),
					),
					Subject: e.TrueResult.Range().Ptr(),
					Context: &e.SrcRange,
				})
				trueResult = cty.UnknownVal(resultType)
			}
		}
		return trueResult, diags
	} else {
		diags = append(diags, falseDiags...)
		if convs[1] != nil {
			var err error
			falseResult, err = convs[1](falseResult)
			if err != nil {
				// Unsafe conversion failed with the concrete result value
				diags = append(diags, &zcl.Diagnostic{
					Severity: zcl.DiagError,
					Summary:  "Inconsistent conditional result types",
					Detail: fmt.Sprintf(
						"The false result value has the wrong type: %s.",
						err.Error(),
					),
					Subject: e.TrueResult.Range().Ptr(),
					Context: &e.SrcRange,
				})
				falseResult = cty.UnknownVal(resultType)
			}
		}
		return falseResult, diags
	}
}

func (e *ConditionalExpr) Range() zcl.Range {
	return e.SrcRange
}

func (e *ConditionalExpr) StartRange() zcl.Range {
	return e.Condition.StartRange()
}

type IndexExpr struct {
	Collection Expression
	Key        Expression

	SrcRange  zcl.Range
	OpenRange zcl.Range
}

func (e *IndexExpr) walkChildNodes(w internalWalkFunc) {
	e.Collection = w(e.Collection).(Expression)
	e.Key = w(e.Key).(Expression)
}

func (e *IndexExpr) Value(ctx *zcl.EvalContext) (cty.Value, zcl.Diagnostics) {
	var diags zcl.Diagnostics
	coll, collDiags := e.Collection.Value(ctx)
	key, keyDiags := e.Key.Value(ctx)
	diags = append(diags, collDiags...)
	diags = append(diags, keyDiags...)

	return zcl.Index(coll, key, &e.SrcRange)
}

func (e *IndexExpr) Range() zcl.Range {
	return e.SrcRange
}

func (e *IndexExpr) StartRange() zcl.Range {
	return e.OpenRange
}

type TupleConsExpr struct {
	Exprs []Expression

	SrcRange  zcl.Range
	OpenRange zcl.Range
}

func (e *TupleConsExpr) walkChildNodes(w internalWalkFunc) {
	for i, expr := range e.Exprs {
		e.Exprs[i] = w(expr).(Expression)
	}
}

func (e *TupleConsExpr) Value(ctx *zcl.EvalContext) (cty.Value, zcl.Diagnostics) {
	var vals []cty.Value
	var diags zcl.Diagnostics

	vals = make([]cty.Value, len(e.Exprs))
	for i, expr := range e.Exprs {
		val, valDiags := expr.Value(ctx)
		vals[i] = val
		diags = append(diags, valDiags...)
	}

	return cty.TupleVal(vals), diags
}

func (e *TupleConsExpr) Range() zcl.Range {
	return e.SrcRange
}

func (e *TupleConsExpr) StartRange() zcl.Range {
	return e.OpenRange
}

type ObjectConsExpr struct {
	Items []ObjectConsItem

	SrcRange  zcl.Range
	OpenRange zcl.Range
}

type ObjectConsItem struct {
	KeyExpr   Expression
	ValueExpr Expression
}

func (e *ObjectConsExpr) walkChildNodes(w internalWalkFunc) {
	for i, item := range e.Items {
		e.Items[i].KeyExpr = w(item.KeyExpr).(Expression)
		e.Items[i].ValueExpr = w(item.ValueExpr).(Expression)
	}
}

func (e *ObjectConsExpr) Value(ctx *zcl.EvalContext) (cty.Value, zcl.Diagnostics) {
	var vals map[string]cty.Value
	var diags zcl.Diagnostics

	// This will get set to true if we fail to produce any of our keys,
	// either because they are actually unknown or if the evaluation produces
	// errors. In all of these case we must return DynamicPseudoType because
	// we're unable to know the full set of keys our object has, and thus
	// we can't produce a complete value of the intended type.
	//
	// We still evaluate all of the item keys and values to make sure that we
	// get as complete as possible a set of diagnostics.
	known := true

	vals = make(map[string]cty.Value, len(e.Items))
	for _, item := range e.Items {
		key, keyDiags := item.KeyExpr.Value(ctx)
		diags = append(diags, keyDiags...)

		val, valDiags := item.ValueExpr.Value(ctx)
		diags = append(diags, valDiags...)

		if keyDiags.HasErrors() {
			known = false
			continue
		}

		if key.IsNull() {
			diags = append(diags, &zcl.Diagnostic{
				Severity: zcl.DiagError,
				Summary:  "Null value as key",
				Detail:   "Can't use a null value as a key.",
				Subject:  item.ValueExpr.Range().Ptr(),
			})
			known = false
			continue
		}

		var err error
		key, err = convert.Convert(key, cty.String)
		if err != nil {
			diags = append(diags, &zcl.Diagnostic{
				Severity: zcl.DiagError,
				Summary:  "Incorrect key type",
				Detail:   fmt.Sprintf("Can't use this value as a key: %s.", err.Error()),
				Subject:  item.ValueExpr.Range().Ptr(),
			})
			known = false
			continue
		}

		if !key.IsKnown() {
			known = false
			continue
		}

		keyStr := key.AsString()

		vals[keyStr] = val
	}

	if !known {
		return cty.DynamicVal, diags
	}

	return cty.ObjectVal(vals), diags
}

func (e *ObjectConsExpr) Range() zcl.Range {
	return e.SrcRange
}

func (e *ObjectConsExpr) StartRange() zcl.Range {
	return e.OpenRange
}
