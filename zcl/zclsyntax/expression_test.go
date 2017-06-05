package zclsyntax

import (
	"testing"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
	"github.com/zclconf/go-zcl/zcl"
)

func TestExpressionParseAndValue(t *testing.T) {
	// This is a combo test that exercises both the parser and the Value
	// method, with the focus on the latter but indirectly testing the former.
	tests := []struct {
		input     string
		ctx       *zcl.EvalContext
		want      cty.Value
		diagCount int
	}{
		{
			`1`,
			nil,
			cty.NumberIntVal(1),
			0,
		},
		{
			`(1)`,
			nil,
			cty.NumberIntVal(1),
			0,
		},
		{
			`(
    1
)`,
			nil,
			cty.NumberIntVal(1),
			0,
		},
		{
			`(1`,
			nil,
			cty.NumberIntVal(1),
			1, // Unbalanced parentheses
		},
		{
			`true`,
			nil,
			cty.True,
			0,
		},
		{
			`false`,
			nil,
			cty.False,
			0,
		},
		{
			`null`,
			nil,
			cty.NullVal(cty.DynamicPseudoType),
			0,
		},
		{
			`true true`,
			nil,
			cty.True,
			1, // extra characters after expression
		},
		{
			`"hello"`,
			nil,
			cty.StringVal("hello"),
			0,
		},
		{
			`"hello\nworld"`,
			nil,
			cty.StringVal("hello\nworld"),
			0,
		},
		{
			`"unclosed`,
			nil,
			cty.StringVal("unclosed"),
			1, // Unterminated template string
		},
		{
			`"hello ${"world"}"`,
			nil,
			cty.StringVal("hello world"),
			0,
		},
		{
			`"hello ${12.5}"`,
			nil,
			cty.StringVal("hello 12.5"),
			0,
		},
		{
			`"silly ${"${"nesting"}"}"`,
			nil,
			cty.StringVal("silly nesting"),
			0,
		},
		{
			`"silly ${"${true}"}"`,
			nil,
			cty.StringVal("silly true"),
			0,
		},
		{
			`"hello $${escaped}"`,
			nil,
			cty.StringVal("hello ${escaped}"),
			0,
		},
		{
			`"hello $$nonescape"`,
			nil,
			cty.StringVal("hello $$nonescape"),
			0,
		},
		{
			`upper("foo")`,
			&zcl.EvalContext{
				Functions: map[string]function.Function{
					"upper": stdlib.UpperFunc,
				},
			},
			cty.StringVal("FOO"),
			0,
		},
		{
			`
upper(
    "foo"
)
`,
			&zcl.EvalContext{
				Functions: map[string]function.Function{
					"upper": stdlib.UpperFunc,
				},
			},
			cty.StringVal("FOO"),
			0,
		},
		{
			`[]`,
			nil,
			cty.EmptyTupleVal,
			0,
		},
		{
			`[1]`,
			nil,
			cty.TupleVal([]cty.Value{cty.NumberIntVal(1)}),
			0,
		},
		{
			`[1,]`,
			nil,
			cty.TupleVal([]cty.Value{cty.NumberIntVal(1)}),
			0,
		},
		{
			`[1,true]`,
			nil,
			cty.TupleVal([]cty.Value{cty.NumberIntVal(1), cty.True}),
			0,
		},
		{
			`[
  1,
  true
]`,
			nil,
			cty.TupleVal([]cty.Value{cty.NumberIntVal(1), cty.True}),
			0,
		},
		{
			`{}`,
			nil,
			cty.EmptyObjectVal,
			0,
		},
		{
			`{"hello": "world"}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"hello": cty.StringVal("world"),
			}),
			0,
		},
		{
			`{"hello" = "world"}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"hello": cty.StringVal("world"),
			}),
			0,
		},
		{
			`{hello = "world"}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"hello": cty.StringVal("world"),
			}),
			0,
		},
		{
			`{hello: "world"}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"hello": cty.StringVal("world"),
			}),
			0,
		},
		{
			`{"hello" = "world", "goodbye" = "cruel world"}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"hello":   cty.StringVal("world"),
				"goodbye": cty.StringVal("cruel world"),
			}),
			0,
		},
		{
			`{
  "hello" = "world"
}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"hello": cty.StringVal("world"),
			}),
			0,
		},
		{
			`{
  "hello" = "world"
  "goodbye" = "cruel world"
}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"hello":   cty.StringVal("world"),
				"goodbye": cty.StringVal("cruel world"),
			}),
			0,
		},
		{
			`{
  "hello" = "world",
  "goodbye" = "cruel world"
}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"hello":   cty.StringVal("world"),
				"goodbye": cty.StringVal("cruel world"),
			}),
			0,
		},
		{
			`{
  "hello" = "world",
  "goodbye" = "cruel world",
}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"hello":   cty.StringVal("world"),
				"goodbye": cty.StringVal("cruel world"),
			}),
			0,
		},

		{
			`["hello"][0]`,
			nil,
			cty.StringVal("hello"),
			0,
		},
		{
			`[][0]`,
			nil,
			cty.DynamicVal,
			1, // invalid index
		},
		{
			`["hello"][negate(0)]`,
			&zcl.EvalContext{
				Functions: map[string]function.Function{
					"negate": stdlib.NegateFunc,
				},
			},
			cty.StringVal("hello"),
			0,
		},
		{
			`[][negate(0)]`,
			&zcl.EvalContext{
				Functions: map[string]function.Function{
					"negate": stdlib.NegateFunc,
				},
			},
			cty.DynamicVal,
			1, // invalid index
		},
		{
			`["hello"]["0"]`, // key gets converted to number
			nil,
			cty.StringVal("hello"),
			0,
		},

		{
			`foo`,
			&zcl.EvalContext{
				Variables: map[string]cty.Value{
					"foo": cty.StringVal("hello"),
				},
			},
			cty.StringVal("hello"),
			0,
		},
		{
			`bar`,
			&zcl.EvalContext{},
			cty.DynamicVal,
			1, // variables not allowed here
		},
		{
			`foo.bar`,
			&zcl.EvalContext{
				Variables: map[string]cty.Value{
					"foo": cty.StringVal("hello"),
				},
			},
			cty.DynamicVal,
			1, // foo does not have attributes
		},
		{
			`foo.baz`,
			&zcl.EvalContext{
				Variables: map[string]cty.Value{
					"foo": cty.ObjectVal(map[string]cty.Value{
						"baz": cty.StringVal("hello"),
					}),
				},
			},
			cty.StringVal("hello"),
			0,
		},
		{
			`foo["baz"]`,
			&zcl.EvalContext{
				Variables: map[string]cty.Value{
					"foo": cty.ObjectVal(map[string]cty.Value{
						"baz": cty.StringVal("hello"),
					}),
				},
			},
			cty.StringVal("hello"),
			0,
		},
		{
			`foo[true]`, // key is converted to string
			&zcl.EvalContext{
				Variables: map[string]cty.Value{
					"foo": cty.ObjectVal(map[string]cty.Value{
						"true": cty.StringVal("hello"),
					}),
				},
			},
			cty.StringVal("hello"),
			0,
		},
		{
			`foo[0].baz`,
			&zcl.EvalContext{
				Variables: map[string]cty.Value{
					"foo": cty.ListVal([]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"baz": cty.StringVal("hello"),
						}),
					}),
				},
			},
			cty.StringVal("hello"),
			0,
		},
		{
			`unk["baz"]`,
			&zcl.EvalContext{
				Variables: map[string]cty.Value{
					"unk": cty.UnknownVal(cty.String),
				},
			},
			cty.DynamicVal,
			1, // value does not have indices (because we know it's a string)
		},
		{
			`unk["boop"]`,
			&zcl.EvalContext{
				Variables: map[string]cty.Value{
					"unk": cty.UnknownVal(cty.Map(cty.String)),
				},
			},
			cty.UnknownVal(cty.String), // we know it's a map of string
			0,
		},
		{
			`dyn["boop"]`,
			&zcl.EvalContext{
				Variables: map[string]cty.Value{
					"dyn": cty.DynamicVal,
				},
			},
			cty.DynamicVal, // don't know what it is yet
			0,
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			expr, parseDiags := ParseExpression([]byte(test.input), "", zcl.Pos{Line: 1, Column: 1, Byte: 0})

			got, valDiags := expr.Value(test.ctx)

			diagCount := len(parseDiags) + len(valDiags)

			if diagCount != test.diagCount {
				t.Errorf("wrong number of diagnostics %d; want %d", diagCount, test.diagCount)
				for _, diag := range parseDiags {
					t.Logf(" - %s", diag.Error())
				}
				for _, diag := range valDiags {
					t.Logf(" - %s", diag.Error())
				}
			}

			if !got.RawEquals(test.want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.want)
			}
		})
	}

}

func TestFunctionCallExprValue(t *testing.T) {
	funcs := map[string]function.Function{
		"length":     stdlib.StrlenFunc,
		"jsondecode": stdlib.JSONDecodeFunc,
	}

	tests := map[string]struct {
		expr      *FunctionCallExpr
		ctx       *zcl.EvalContext
		want      cty.Value
		diagCount int
	}{
		"valid call with no conversions": {
			&FunctionCallExpr{
				Name: "length",
				Args: []Expression{
					&LiteralValueExpr{
						Val: cty.StringVal("hello"),
					},
				},
			},
			&zcl.EvalContext{
				Functions: funcs,
			},
			cty.NumberIntVal(5),
			0,
		},
		"valid call with arg conversion": {
			&FunctionCallExpr{
				Name: "length",
				Args: []Expression{
					&LiteralValueExpr{
						Val: cty.BoolVal(true),
					},
				},
			},
			&zcl.EvalContext{
				Functions: funcs,
			},
			cty.NumberIntVal(4), // length of string "true"
			0,
		},
		"valid call with unknown arg": {
			&FunctionCallExpr{
				Name: "length",
				Args: []Expression{
					&LiteralValueExpr{
						Val: cty.UnknownVal(cty.String),
					},
				},
			},
			&zcl.EvalContext{
				Functions: funcs,
			},
			cty.UnknownVal(cty.Number),
			0,
		},
		"valid call with unknown arg needing conversion": {
			&FunctionCallExpr{
				Name: "length",
				Args: []Expression{
					&LiteralValueExpr{
						Val: cty.UnknownVal(cty.Bool),
					},
				},
			},
			&zcl.EvalContext{
				Functions: funcs,
			},
			cty.UnknownVal(cty.Number),
			0,
		},
		"valid call with dynamic arg": {
			&FunctionCallExpr{
				Name: "length",
				Args: []Expression{
					&LiteralValueExpr{
						Val: cty.DynamicVal,
					},
				},
			},
			&zcl.EvalContext{
				Functions: funcs,
			},
			cty.UnknownVal(cty.Number),
			0,
		},
		"invalid arg type": {
			&FunctionCallExpr{
				Name: "length",
				Args: []Expression{
					&LiteralValueExpr{
						Val: cty.ListVal([]cty.Value{cty.StringVal("hello")}),
					},
				},
			},
			&zcl.EvalContext{
				Functions: funcs,
			},
			cty.DynamicVal,
			1,
		},
		"function with dynamic return type": {
			&FunctionCallExpr{
				Name: "jsondecode",
				Args: []Expression{
					&LiteralValueExpr{
						Val: cty.StringVal(`"hello"`),
					},
				},
			},
			&zcl.EvalContext{
				Functions: funcs,
			},
			cty.StringVal("hello"),
			0,
		},
		"function with dynamic return type unknown arg": {
			&FunctionCallExpr{
				Name: "jsondecode",
				Args: []Expression{
					&LiteralValueExpr{
						Val: cty.UnknownVal(cty.String),
					},
				},
			},
			&zcl.EvalContext{
				Functions: funcs,
			},
			cty.DynamicVal, // type depends on arg value
			0,
		},
		"error in function": {
			&FunctionCallExpr{
				Name: "jsondecode",
				Args: []Expression{
					&LiteralValueExpr{
						Val: cty.StringVal("invalid-json"),
					},
				},
			},
			&zcl.EvalContext{
				Functions: funcs,
			},
			cty.DynamicVal,
			1, // JSON parse error
		},
		"unknown function": {
			&FunctionCallExpr{
				Name: "lenth",
				Args: []Expression{},
			},
			&zcl.EvalContext{
				Functions: funcs,
			},
			cty.DynamicVal,
			1,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, diags := test.expr.Value(test.ctx)

			if len(diags) != test.diagCount {
				t.Errorf("wrong number of diagnostics %d; want %d", len(diags), test.diagCount)
				for _, diag := range diags {
					t.Logf(" - %s", diag.Error())
				}
			}

			if !got.RawEquals(test.want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.want)
			}
		})
	}
}
