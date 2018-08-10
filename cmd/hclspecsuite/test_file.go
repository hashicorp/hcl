package main

import (
	"fmt"

	"github.com/hashicorp/hcl2/ext/typeexpr"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

type TestFile struct {
	Result     cty.Value
	ResultType cty.Type

	Traversals []hcl.Traversal
}

func (r *Runner) LoadTestFile(filename string) (*TestFile, hcl.Diagnostics) {
	f, diags := r.parser.ParseHCLFile(filename)
	if diags.HasErrors() {
		return nil, diags
	}

	content, moreDiags := f.Body.Content(testFileSchema)
	diags = append(diags, moreDiags...)
	if moreDiags.HasErrors() {
		return nil, diags
	}

	ret := &TestFile{
		ResultType: cty.DynamicPseudoType,
	}

	if typeAttr, exists := content.Attributes["result_type"]; exists {
		ty, moreDiags := typeexpr.TypeConstraint(typeAttr.Expr)
		diags = append(diags, moreDiags...)
		if !moreDiags.HasErrors() {
			ret.ResultType = ty
		}
	}

	if resultAttr, exists := content.Attributes["result"]; exists {
		resultVal, moreDiags := resultAttr.Expr.Value(nil)
		diags = append(diags, moreDiags...)
		if !moreDiags.HasErrors() {
			resultVal, err := convert.Convert(resultVal, ret.ResultType)
			if err != nil {
				diags = diags.Append(&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid result value",
					Detail:   fmt.Sprintf("The result value does not conform to the given result type: %s.", err),
					Subject:  resultAttr.Expr.Range().Ptr(),
				})
			} else {
				ret.Result = resultVal
			}
		}
	}

	return ret, diags
}

var testFileSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: "result",
		},
		{
			Name: "result_type",
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: "traversals",
		},
	},
}
