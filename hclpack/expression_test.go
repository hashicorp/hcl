package hclpack

import (
	"testing"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/zclconf/go-cty/cty"
)

func TestExpressionValue(t *testing.T) {
	tests := map[string]struct {
		Expr *Expression
		Ctx  *hcl.EvalContext
		Want cty.Value
	}{
		"simple literal expr": {
			&Expression{
				Source:     []byte(`"hello"`),
				SourceType: ExprNative,
			},
			nil,
			cty.StringVal("hello"),
		},
		"simple literal template": {
			&Expression{
				Source:     []byte(`hello ${5}`),
				SourceType: ExprTemplate,
			},
			nil,
			cty.StringVal("hello 5"),
		},
		"expr with variable": {
			&Expression{
				Source:     []byte(`foo`),
				SourceType: ExprNative,
			},
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"foo": cty.StringVal("bar"),
				},
			},
			cty.StringVal("bar"),
		},
		"template with variable": {
			&Expression{
				Source:     []byte(`foo ${foo}`),
				SourceType: ExprTemplate,
			},
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"foo": cty.StringVal("bar"),
				},
			},
			cty.StringVal("foo bar"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, diags := test.Expr.Value(test.Ctx)
			for _, diag := range diags {
				t.Errorf("unexpected diagnostic: %s", diag.Error())
			}

			if !test.Want.RawEquals(got) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}
