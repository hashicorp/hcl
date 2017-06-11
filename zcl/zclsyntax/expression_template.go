package zclsyntax

import (
	"bytes"
	"fmt"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-zcl/zcl"
)

type TemplateExpr struct {
	Parts  []Expression
	Unwrap bool

	SrcRange zcl.Range
}

func (e *TemplateExpr) walkChildNodes(w internalWalkFunc) {
	for i, part := range e.Parts {
		e.Parts[i] = w(part).(Expression)
	}
}

func (e *TemplateExpr) Value(ctx *zcl.EvalContext) (cty.Value, zcl.Diagnostics) {
	if e.Unwrap {
		if len(e.Parts) != 1 {
			// should never happen - parser bug, if so
			panic("Unwrap set with len(e.Parts) != 1")
		}
		return e.Parts[0].Value(ctx)
	}

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
