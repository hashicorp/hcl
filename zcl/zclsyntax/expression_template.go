package zclsyntax

import (
	"bytes"
	"fmt"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-zcl/zcl"
)

type TemplateExpr struct {
	Parts []Expression

	SrcRange zcl.Range
}

func (e *TemplateExpr) walkChildNodes(w internalWalkFunc) {
	for i, part := range e.Parts {
		e.Parts[i] = w(part).(Expression)
	}
}

func (e *TemplateExpr) Value(ctx *zcl.EvalContext) (cty.Value, zcl.Diagnostics) {
	buf := &bytes.Buffer{}
	var diags zcl.Diagnostics
	isKnown := true

	for _, part := range e.Parts {
		partVal, partDiags := part.Value(ctx)
		diags = append(diags, partDiags...)

		if partVal.IsNull() {
			diags = append(diags, &zcl.Diagnostic{
				Severity: zcl.DiagError,
				Summary:  "Invalid template interpolation value",
				Detail: fmt.Sprintf(
					"The expression result is null. Cannot include a null value in a string template.",
				),
				Subject: part.Range().Ptr(),
				Context: &e.SrcRange,
			})
			continue
		}

		if !partVal.IsKnown() {
			// If any part is unknown then the result as a whole must be
			// unknown too. We'll keep on processing the rest of the parts
			// anyway, because we want to still emit any diagnostics resulting
			// from evaluating those.
			isKnown = false
			continue
		}

		strVal, err := convert.Convert(partVal, cty.String)
		if err != nil {
			diags = append(diags, &zcl.Diagnostic{
				Severity: zcl.DiagError,
				Summary:  "Invalid template interpolation value",
				Detail: fmt.Sprintf(
					"Cannot include the given value in a string template: %s.",
					err.Error(),
				),
				Subject: part.Range().Ptr(),
				Context: &e.SrcRange,
			})
			continue
		}

		buf.WriteString(strVal.AsString())
	}

	if !isKnown {
		return cty.UnknownVal(cty.String), diags
	}

	return cty.StringVal(buf.String()), diags
}

func (e *TemplateExpr) Range() zcl.Range {
	return e.SrcRange
}

func (e *TemplateExpr) StartRange() zcl.Range {
	return e.Parts[0].StartRange()
}

// TemplateWrapExpr is used instead of a TemplateExpr when a template
// consists _only_ of a single interpolation sequence. In that case, the
// template's result is the single interpolation's result, verbatim with
// no type conversions.
type TemplateWrapExpr struct {
	Wrapped Expression

	SrcRange zcl.Range
}

func (e *TemplateWrapExpr) walkChildNodes(w internalWalkFunc) {
	e.Wrapped = w(e.Wrapped).(Expression)
}

func (e *TemplateWrapExpr) Value(ctx *zcl.EvalContext) (cty.Value, zcl.Diagnostics) {
	return e.Wrapped.Value(ctx)
}

func (e *TemplateWrapExpr) Range() zcl.Range {
	return e.SrcRange
}

func (e *TemplateWrapExpr) StartRange() zcl.Range {
	return e.SrcRange
}
