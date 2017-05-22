package hclhil

import (
	"testing"

	"github.com/apparentlymart/go-cty/cty"
	"github.com/apparentlymart/go-cty/cty/function"
	"github.com/apparentlymart/go-cty/cty/function/stdlib"
	"github.com/apparentlymart/go-zcl/zcl"
)

func TestTemplateExpression(t *testing.T) {
	tests := []struct {
		input     string
		ctx       *zcl.EvalContext
		want      cty.Value
		diagCount int
	}{
		{
			``,
			nil,
			cty.StringVal(""),
			0,
		},
		{
			`hello`,
			nil,
			cty.StringVal("hello"),
			0,
		},
		{
			`hello ${"world"}`,
			nil,
			cty.StringVal("hello world"),
			0,
		},
		{
			`${"hello"}`,
			nil,
			cty.StringVal("hello"),
			0,
		},
		{
			`Hello ${planet}!`,
			&zcl.EvalContext{
				Variables: map[string]cty.Value{
					"planet": cty.StringVal("Earth"),
				},
			},
			cty.StringVal("Hello Earth!"),
			0,
		},
		{
			`${names}`,
			&zcl.EvalContext{
				Variables: map[string]cty.Value{
					"names": cty.ListVal([]cty.Value{
						cty.StringVal("Ermintrude"),
						cty.StringVal("Tom"),
					}),
				},
			},
			cty.TupleVal([]cty.Value{
				cty.StringVal("Ermintrude"),
				cty.StringVal("Tom"),
			}),
			0,
		},
		{
			`${doodads}`,
			&zcl.EvalContext{
				Variables: map[string]cty.Value{
					"doodads": cty.MapVal(map[string]cty.Value{
						"Captain":       cty.StringVal("Ermintrude"),
						"First Officer": cty.StringVal("Tom"),
					}),
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"Captain":       cty.StringVal("Ermintrude"),
				"First Officer": cty.StringVal("Tom"),
			}),
			0,
		},
		{
			`${names}`,
			&zcl.EvalContext{
				Variables: map[string]cty.Value{
					"names": cty.TupleVal([]cty.Value{
						cty.StringVal("Ermintrude"),
						cty.NumberIntVal(5),
					}),
				},
			},
			cty.TupleVal([]cty.Value{
				cty.StringVal("Ermintrude"),
				cty.StringVal("5"),
			}),
			0,
		},
		{
			`${messytuple}`,
			&zcl.EvalContext{
				Variables: map[string]cty.Value{
					"messytuple": cty.TupleVal([]cty.Value{
						cty.StringVal("Ermintrude"),
						cty.ListValEmpty(cty.String),
					}),
				},
			},
			cty.TupleVal([]cty.Value{
				cty.StringVal("Ermintrude"),
				cty.ListValEmpty(cty.String), // HIL's sloppy type checker actually lets us get away with this
			}),
			0,
		},
		{
			`number ${num}`,
			&zcl.EvalContext{
				Variables: map[string]cty.Value{
					"num": cty.NumberIntVal(5),
				},
			},
			cty.StringVal("number 5"),
			0,
		},
		{
			`${length("hello")}`,
			&zcl.EvalContext{
				Functions: map[string]function.Function{
					"length": stdlib.StrlenFunc,
				},
			},
			cty.StringVal("5"), // HIL always stringifies numbers on output
			0,
		},
		{
			`${true}`,
			nil,
			cty.StringVal("true"), // HIL always stringifies bools on output
			0,
		},
		{
			`cannot ${names}`,
			&zcl.EvalContext{
				Variables: map[string]cty.Value{
					"names": cty.ListVal([]cty.Value{
						cty.StringVal("Ermintrude"),
						cty.StringVal("Tom"),
					}),
				},
			},
			cty.DynamicVal,
			1, // can't concatenate a list
		},
		{
			`${syntax error`,
			nil,
			cty.NilVal,
			1,
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			expr, diags := ParseTemplate([]byte(test.input), "test.hil")
			if expr != nil {
				val, valDiags := expr.Value(test.ctx)
				diags = append(diags, valDiags...)
				if !val.RawEquals(test.want) {
					t.Errorf("wrong result\ngot:  %#v\nwant: %#v", val, test.want)
				}
			} else {
				if test.want != cty.NilVal {
					t.Errorf("Unexpected diagnostics during parse: %s", diags.Error())
				}
			}

			if len(diags) != test.diagCount {
				t.Errorf("wrong number of diagnostics %d; want %d", len(diags), test.diagCount)
				for _, diag := range diags {
					t.Logf(" - %s", diag.Error())
				}
			}
		})
	}
}
