package gozcl

import (
	"fmt"

	"github.com/apparentlymart/go-cty/cty/convert"
	"github.com/apparentlymart/go-cty/cty/gocty"
	"github.com/apparentlymart/go-zcl/zcl"
)

// DecodeBody extracts the configuration within the given body into the given
// value. This value must be a non-nil pointer to either a struct or
// a map, where in the former case the configuration will be decoded using
// struct tags and in the latter case only attributes are allowed and their
// values are decoded into the map.
//
// The given EvalContext is used to resolve any variables or functions in
// expressions encountered while decoding. This may be nil to require only
// constant values, for simple applications that do not support variables or
// functions.
//
// The returned diagnostics should be inspected with its HasErrors method to
// determine if the populated value is valid and complete. If error diagnostics
// are returned then the given value may have been partially-populated but
// may still be accessed by a careful caller for static analysis and editor
// integration use-cases.
func DecodeBody(body zcl.Body, ctx *zcl.EvalContext, val interface{}) zcl.Diagnostics {
	// TODO: Implement
	return nil
}

// DecodeExpression extracts the value of the given expression into the given
// value. This value must be something that gocty is able to decode into,
// since the final decoding is delegated to that package.
//
// The given EvalContext is used to resolve any variables or functions in
// expressions encountered while decoding. This may be nil to require only
// constant values, for simple applications that do not support variables or
// functions.
//
// The returned diagnostics should be inspected with its HasErrors method to
// determine if the populated value is valid and complete. If error diagnostics
// are returned then the given value may have been partially-populated but
// may still be accessed by a careful caller for static analysis and editor
// integration use-cases.
func DecodeExpression(expr zcl.Expression, ctx *zcl.EvalContext, val interface{}) zcl.Diagnostics {
	srcVal, diags := expr.Value(ctx)

	convTy, err := gocty.ImpliedType(val)
	if err != nil {
		panic(fmt.Sprintf("unsuitable DecodeExpression target: %s", err))
	}

	srcVal, err = convert.Convert(srcVal, convTy)
	if err != nil {
		diags = append(diags, &zcl.Diagnostic{
			Severity: zcl.DiagError,
			Summary:  "Unsuitable value type",
			Detail:   fmt.Sprintf("Incorrect value type: %s", err.Error()),
			Subject:  expr.StartRange().Ptr(),
			Context:  expr.Range().Ptr(),
		})
		return diags
	}

	err = gocty.FromCtyValue(srcVal, val)
	if err != nil {
		diags = append(diags, &zcl.Diagnostic{
			Severity: zcl.DiagError,
			Summary:  "Unsuitable value type",
			Detail:   fmt.Sprintf("Incorrect value type: %s", err.Error()),
			Subject:  expr.StartRange().Ptr(),
			Context:  expr.Range().Ptr(),
		})
	}

	return diags
}
