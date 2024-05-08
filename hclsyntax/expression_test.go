// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hclsyntax

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

func TestExpressionParseAndValue(t *testing.T) {
	// This is a combo test that exercises both the parser and the Value
	// method, with the focus on the latter but indirectly testing the former.
	tests := []struct {
		input     string
		ctx       *hcl.EvalContext
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
			`(2+3)`,
			nil,
			cty.NumberIntVal(5),
			0,
		},
		{
			`2*5+1`,
			nil,
			cty.NumberIntVal(11),
			0,
		},
		{
			`9%8`,
			nil,
			cty.NumberIntVal(1),
			0,
		},
		{
			`(2+unk)`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unk": cty.UnknownVal(cty.Number),
				},
			},
			cty.UnknownVal(cty.Number).RefineNotNull(),
			0,
		},
		{
			`(2+unk)`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unk": cty.DynamicVal,
				},
			},
			cty.UnknownVal(cty.Number).RefineNotNull(),
			0,
		},
		{
			`(unk+unk)`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unk": cty.DynamicVal,
				},
			},
			cty.UnknownVal(cty.Number).RefineNotNull(),
			0,
		},
		{
			`(2+true)`,
			nil,
			cty.UnknownVal(cty.Number),
			1, // unsuitable type for right operand
		},
		{
			`(false+true)`,
			nil,
			cty.UnknownVal(cty.Number),
			2, // unsuitable type for each operand
		},
		{
			`(5 == 5)`,
			nil,
			cty.True,
			0,
		},
		{
			`(5 == 4)`,
			nil,
			cty.False,
			0,
		},
		{
			`(1 == true)`,
			nil,
			cty.False,
			0,
		},
		{
			`("true" == true)`,
			nil,
			cty.False,
			0,
		},
		{
			`(true == "true")`,
			nil,
			cty.False,
			0,
		},
		{
			`(true != "true")`,
			nil,
			cty.True,
			0,
		},
		{
			`(- 2)`,
			nil,
			cty.NumberIntVal(-2),
			0,
		},
		{
			`(! true)`,
			nil,
			cty.False,
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
			"\"hello `backtick` world\"",
			nil,
			cty.StringVal("hello `backtick` world"),
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
			`"$"`,
			nil,
			cty.StringVal("$"),
			0,
		},
		{
			`"%"`,
			nil,
			cty.StringVal("%"),
			0,
		},
		{
			`upper("foo")`,
			&hcl.EvalContext{
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
			&hcl.EvalContext{
				Functions: map[string]function.Function{
					"upper": stdlib.UpperFunc,
				},
			},
			cty.StringVal("FOO"),
			0,
		},
		{
			`upper(["foo"]...)`,
			&hcl.EvalContext{
				Functions: map[string]function.Function{
					"upper": stdlib.UpperFunc,
				},
			},
			cty.StringVal("FOO"),
			0,
		},
		{
			`upper("foo", []...)`,
			&hcl.EvalContext{
				Functions: map[string]function.Function{
					"upper": stdlib.UpperFunc,
				},
			},
			cty.StringVal("FOO"),
			0,
		},
		{
			`upper("foo", "bar")`,
			&hcl.EvalContext{
				Functions: map[string]function.Function{
					"upper": stdlib.UpperFunc,
				},
			},
			cty.DynamicVal,
			1, // too many function arguments
		},
		{
			`upper(["foo", "bar"]...)`,
			&hcl.EvalContext{
				Functions: map[string]function.Function{
					"upper": stdlib.UpperFunc,
				},
			},
			cty.DynamicVal,
			1, // too many function arguments
		},
		{
			`concat([1, null]...)`,
			&hcl.EvalContext{
				Functions: map[string]function.Function{
					"concat": stdlib.ConcatFunc,
				},
			},
			cty.DynamicVal,
			1, // argument cannot be null
		},
		{
			`concat(var.unknownlist...)`,
			&hcl.EvalContext{
				Functions: map[string]function.Function{
					"concat": stdlib.ConcatFunc,
				},
				Variables: map[string]cty.Value{
					"var": cty.ObjectVal(map[string]cty.Value{
						"unknownlist": cty.UnknownVal(cty.DynamicPseudoType),
					}),
				},
			},
			cty.DynamicVal,
			0,
		},
		{
			`foo::upper("foo")`,
			&hcl.EvalContext{
				Functions: map[string]function.Function{
					"foo::upper": stdlib.UpperFunc,
				},
			},
			cty.StringVal("FOO"),
			0,
		},
		{
			`foo :: upper("foo")`, // spaces are non-idomatic, but valid
			&hcl.EvalContext{
				Functions: map[string]function.Function{
					"foo::upper": stdlib.UpperFunc,
				},
			},
			cty.StringVal("FOO"),
			0,
		},
		{
			`::upper("foo")`, // :: is still not a valid identifier
			&hcl.EvalContext{
				Functions: map[string]function.Function{
					"::upper": stdlib.UpperFunc,
				},
			},
			cty.DynamicVal,
			1,
		},
		{
			`double::::upper("foo")`, // missing name after ::
			&hcl.EvalContext{
				Functions: map[string]function.Function{
					"double::::upper": stdlib.UpperFunc,
				},
			},
			cty.DynamicVal,
			2,
		},
		{
			`missing::("foo")`, // missing name after ::
			&hcl.EvalContext{
				Functions: map[string]function.Function{
					"missing::": stdlib.UpperFunc,
				},
			},
			cty.DynamicVal,
			2,
		},
		{
			`misbehave()`,
			&hcl.EvalContext{
				Functions: map[string]function.Function{
					"misbehave": function.New(&function.Spec{
						Type: func(args []cty.Value) (cty.Type, error) {
							// This function misbehaves by indicating an error
							// on an argument index that is out of range for
							// its declared parameters. That would always be
							// a bug in the function, but we want to avoid
							// panicking in this case and just behave like it
							// was a normal (non-arg) error.
							return cty.NilType, function.NewArgErrorf(1, "out of range")
						},
					}),
				},
			},
			cty.DynamicVal,
			1, // Call to function "misbehave" failed: out of range
		},
		{
			`misbehave() /* variadic */`,
			&hcl.EvalContext{
				Functions: map[string]function.Function{
					"misbehave": function.New(&function.Spec{
						VarParam: &function.Parameter{
							Name: "foo",
							Type: cty.String,
						},
						Type: func(args []cty.Value) (cty.Type, error) {
							// This function misbehaves by indicating an error
							// on an argument index that is out of range for
							// the given arguments. That would always be a
							// bug in the function, but to avoid panicking we
							// just treat it like a problem related to the
							// declared variadic argument.
							return cty.NilType, function.NewArgErrorf(1, "out of range")
						},
					}),
				},
			},
			cty.DynamicVal,
			1, // Invalid value for "foo" parameter: out of range
		},
		{
			`misbehave([]...)`,
			&hcl.EvalContext{
				Functions: map[string]function.Function{
					"misbehave": function.New(&function.Spec{
						VarParam: &function.Parameter{
							Name: "foo",
							Type: cty.String,
						},
						Type: func(args []cty.Value) (cty.Type, error) {
							// This function misbehaves by indicating an error
							// on an argument index that is out of range for
							// the given arguments. That would always be a
							// bug in the function, but to avoid panicking we
							// just treat it like a problem related to the
							// declared variadic argument.
							return cty.NilType, function.NewArgErrorf(1, "out of range")
						},
					}),
				},
			},
			cty.DynamicVal,
			1, // Invalid value for "foo" parameter: out of range
		},
		{
			`argerrorexpand(["a", "b"]...)`,
			&hcl.EvalContext{
				Functions: map[string]function.Function{
					"argerrorexpand": function.New(&function.Spec{
						VarParam: &function.Parameter{
							Name: "foo",
							Type: cty.String,
						},
						Type: func(args []cty.Value) (cty.Type, error) {
							// We should be able to indicate an error in
							// argument 1 because the indices are into the
							// arguments _after_ "..." expansion. An earlier
							// HCL version had a bug where it used the
							// pre-expansion arguments and would thus panic
							// in this case.
							return cty.NilType, function.NewArgErrorf(1, "blah blah")
						},
					}),
				},
			},
			cty.DynamicVal,
			1, // Invalid value for "foo" parameter: blah blah
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
			`{true: "yes"}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"true": cty.StringVal("yes"),
			}),
			0,
		},
		{
			`{false: "yes"}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"false": cty.StringVal("yes"),
			}),
			0,
		},
		{
			`{null: "yes"}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"null": cty.StringVal("yes"),
			}),
			0,
		},
		{
			`{15: "yes"}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"15": cty.StringVal("yes"),
			}),
			0,
		},
		{
			`{[]: "yes"}`,
			nil,
			cty.DynamicVal,
			1, // Incorrect key type; Can't use this value as a key: string required
		},
		{
			`{"centos_7.2_ap-south-1" = "ami-abc123"}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"centos_7.2_ap-south-1": cty.StringVal("ami-abc123"),
			}),
			0,
		},
		{
			// This is syntactically valid (it's similar to foo["bar"])
			// but is rejected during evaluation to force the user to be explicit
			// about which of the following interpretations they mean:
			// -{(foo.bar) = "baz"}
			// -{"foo.bar" = "baz"}
			// naked traversals as keys are allowed when analyzing an expression
			// statically so an application can define object-syntax-based
			// language constructs with looser requirements, but we reject
			// this during normal expression evaluation.
			`{foo.bar = "ami-abc123"}`,
			nil,
			cty.DynamicVal,
			1, // Ambiguous attribute key; If this expression is intended to be a reference, wrap it in parentheses. If it's instead intended as a literal name containing periods, wrap it in quotes to create a string literal.
		},
		{
			// This is a weird variant of the above where a period is followed
			// by a digit, causing the parser to interpret it as an index
			// operator using the legacy HIL/Terraform index syntax.
			// This one _does_ fail parsing, causing it to be subject to
			// parser recovery behavior.
			`{centos_7.2_ap-south-1 = "ami-abc123"}`,
			nil,
			cty.EmptyObjectVal, // (due to parser recovery behavior)
			1,                  // Missing key/value separator; Expected an equals sign ("=") to mark the beginning of the attribute value. If you intended to given an attribute name containing periods or spaces, write the name in quotes to create a string literal.
		},
		{
			`{var.greeting = "world"}`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"var": cty.ObjectVal(map[string]cty.Value{
						"greeting": cty.StringVal("hello"),
					}),
				},
			},
			cty.DynamicVal,
			1, // Ambiguous attribute key
		},
		{
			`{(var.greeting) = "world"}`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"var": cty.ObjectVal(map[string]cty.Value{
						"greeting": cty.StringVal("hello"),
					}),
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"hello": cty.StringVal("world"),
			}),
			0,
		},
		{
			// Marked values as object keys
			`{(var.greeting) = "world", "goodbye" = "earth"}`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"var": cty.ObjectVal(map[string]cty.Value{
						"greeting": cty.StringVal("hello").Mark("marked"),
					}),
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"hello":   cty.StringVal("world"),
				"goodbye": cty.StringVal("earth"),
			}).Mark("marked"),
			0,
		},
		{
			`{"${var.greeting}" = "world"}`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"var": cty.ObjectVal(map[string]cty.Value{
						"greeting": cty.StringVal("hello"),
					}),
				},
			},
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
			"{\n  for k, v in {hello: \"world\"}:\nk => v\n}",
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"hello": cty.StringVal("world"),
			}),
			0,
		},
		{
			// This one is different than the previous because the extra level of
			// object constructor causes the inner for expression to begin parsing
			// in newline-sensitive mode, which it must then properly disable in
			// order to peek the "for" keyword.
			"{\n  a = {\n  for k, v in {hello: \"world\"}:\nk => v\n  }\n}",
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"a": cty.ObjectVal(map[string]cty.Value{
					"hello": cty.StringVal("world"),
				}),
			}),
			0,
		},
		{
			`{for k, v in {hello: "world"}: k => v if k == "hello"}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"hello": cty.StringVal("world"),
			}),
			0,
		},
		{
			`{for k, v in {hello: "world"}: upper(k) => upper(v) if k == "hello"}`,
			&hcl.EvalContext{
				Functions: map[string]function.Function{
					"upper": stdlib.UpperFunc,
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"HELLO": cty.StringVal("WORLD"),
			}),
			0,
		},
		{
			`{for k, v in ["world"]: k => v if k == 0}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"0": cty.StringVal("world"),
			}),
			0,
		},
		{
			`{for v in ["world"]: v => v}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"world": cty.StringVal("world"),
			}),
			0,
		},
		{
			`{for k, v in {hello: "world"}: k => v if k == "foo"}`,
			nil,
			cty.EmptyObjectVal,
			0,
		},
		{
			`{for k, v in {hello: "world"}: 5 => v}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"5": cty.StringVal("world"),
			}),
			0,
		},
		{
			`{for k, v in {hello: "world"}: [] => v}`,
			nil,
			cty.DynamicVal,
			1, // key expression has the wrong type
		},
		{
			`{for k, v in {hello: "world"}: k => k if k == "hello"}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"hello": cty.StringVal("hello"),
			}),
			0,
		},
		{
			`{for k, v in {hello: "world"}: k => foo}`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"foo": cty.StringVal("foo"),
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"hello": cty.StringVal("foo"),
			}),
			0,
		},
		{
			`[for k, v in {hello: "world"}: "${k}=${v}"]`,
			nil,
			cty.TupleVal([]cty.Value{
				cty.StringVal("hello=world"),
			}),
			0,
		},
		{
			`[for k, v in {hello: "world"}: k => v]`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"hello": cty.StringVal("world"),
			}),
			1, // can't have a key expr when producing a tuple
		},
		{
			`{for v in {hello: "world"}: v}`,
			nil,
			cty.TupleVal([]cty.Value{
				cty.StringVal("world"),
			}),
			1, // must have a key expr when producing a map
		},
		{
			`{for i, v in ["a", "b", "c", "b", "d"]: v => i...}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"a": cty.TupleVal([]cty.Value{
					cty.NumberIntVal(0),
				}),
				"b": cty.TupleVal([]cty.Value{
					cty.NumberIntVal(1),
					cty.NumberIntVal(3),
				}),
				"c": cty.TupleVal([]cty.Value{
					cty.NumberIntVal(2),
				}),
				"d": cty.TupleVal([]cty.Value{
					cty.NumberIntVal(4),
				}),
			}),
			0,
		},
		{
			`{for i, v in ["a", "b", "c", "b", "d"]: v => i... if i <= 2}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"a": cty.TupleVal([]cty.Value{
					cty.NumberIntVal(0),
				}),
				"b": cty.TupleVal([]cty.Value{
					cty.NumberIntVal(1),
				}),
				"c": cty.TupleVal([]cty.Value{
					cty.NumberIntVal(2),
				}),
			}),
			0,
		},
		{
			`{for i, v in ["a", "b", "c", "b", "d"]: v => i}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"a": cty.NumberIntVal(0),
				"b": cty.NumberIntVal(1),
				"c": cty.NumberIntVal(2),
				"d": cty.NumberIntVal(4),
			}),
			1, // duplicate key "b"
		},
		{
			`[for v in {hello: "world"}: v...]`,
			nil,
			cty.TupleVal([]cty.Value{
				cty.StringVal("world"),
			}),
			1, // can't use grouping when producing a tuple
		},
		{
			`[for v in "hello": v]`,
			nil,
			cty.DynamicVal,
			1, // can't iterate over a string
		},
		{
			`[for v in null: v]`,
			nil,
			cty.DynamicVal,
			1, // can't iterate over a null value
		},
		{
			`[for v in unk: v]`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unk": cty.UnknownVal(cty.List(cty.String)),
				},
			},
			cty.DynamicVal,
			0,
		},
		{
			`[for v in unk: v]`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unk": cty.DynamicVal,
				},
			},
			cty.DynamicVal,
			0,
		},
		{
			`[for v in unk: v]`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unk": cty.UnknownVal(cty.String),
				},
			},
			cty.DynamicVal,
			1, // can't iterate over a string (even if it's unknown)
		},
		{
			`[for v in ["a", "b"]: v if unkbool]`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unkbool": cty.UnknownVal(cty.Bool),
				},
			},
			cty.DynamicVal,
			0,
		},
		{
			`[for v in ["a", "b"]: v if nullbool]`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"nullbool": cty.NullVal(cty.Bool),
				},
			},
			cty.DynamicVal,
			1, // value of if clause must not be null
		},
		{
			`[for v in ["a", "b"]: v if dyn]`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"dyn": cty.DynamicVal,
				},
			},
			cty.DynamicVal,
			0,
		},
		{
			`[for v in ["a", "b"]: v if unknum]`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unknum": cty.UnknownVal(cty.List(cty.Number)),
				},
			},
			cty.DynamicVal,
			1, // if expression must be bool
		},
		{
			`[for i, v in ["a", "b"]: v if i + i]`,
			nil,
			cty.DynamicVal,
			1, // if expression must be bool
		},
		{
			`[for v in ["a", "b"]: unkstr]`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unkstr": cty.UnknownVal(cty.String),
				},
			},
			cty.TupleVal([]cty.Value{
				cty.UnknownVal(cty.String),
				cty.UnknownVal(cty.String),
			}),
			0,
		},
		{ // Marked sequence results in a marked tuple
			`[for x in things: x if x != ""]`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"things": cty.ListVal([]cty.Value{
						cty.StringVal("a"),
						cty.StringVal("b"),
						cty.StringVal(""),
						cty.StringVal("c"),
					}).Mark("sensitive"),
				},
			},
			cty.TupleVal([]cty.Value{
				cty.StringVal("a"),
				cty.StringVal("b"),
				cty.StringVal("c"),
			}).Mark("sensitive"),
			0,
		},
		{ // Marked map results in a marked object
			`{for k, v in things: k => !v}`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"things": cty.MapVal(map[string]cty.Value{
						"a": cty.True,
						"b": cty.False,
					}).Mark("sensitive"),
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"a": cty.False,
				"b": cty.True,
			}).Mark("sensitive"),
			0,
		},
		{ // Marked map member carries marks through
			`{for k, v in things: k => !v}`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"things": cty.MapVal(map[string]cty.Value{
						"a": cty.True.Mark("sensitive"),
						"b": cty.False,
					}),
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"a": cty.False.Mark("sensitive"),
				"b": cty.True,
			}),
			0,
		},
		{
			// Mark object if keys include marked values, members retain
			// their original marks in their values
			`{for v in things: v => "${v}-friend"}`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"things": cty.MapVal(map[string]cty.Value{
						"a": cty.StringVal("rosie").Mark("marked"),
						"b": cty.StringVal("robin"),
						// Check for double-marking when a key val has a duplicate mark
						"c": cty.StringVal("rowan").Mark("marked"),
						"d": cty.StringVal("ruben").Mark("also-marked"),
					}),
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"rosie": cty.StringVal("rosie-friend").Mark("marked"),
				"robin": cty.StringVal("robin-friend"),
				"rowan": cty.StringVal("rowan-friend").Mark("marked"),
				"ruben": cty.StringVal("ruben-friend").Mark("also-marked"),
			}).WithMarks(cty.NewValueMarks("marked", "also-marked")),
			0,
		},
		{ // object itself is marked, contains marked value
			`{for v in things: v => "${v}-friend"}`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"things": cty.MapVal(map[string]cty.Value{
						"a": cty.StringVal("rosie").Mark("marked"),
						"b": cty.StringVal("robin"),
					}).Mark("marks"),
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"rosie": cty.StringVal("rosie-friend").Mark("marked"),
				"robin": cty.StringVal("robin-friend"),
			}).WithMarks(cty.NewValueMarks("marked", "marks")),
			0,
		},
		{ // Sequence for loop with marked conditional expression
			`[for x in things: x if x != secret]`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"things": cty.ListVal([]cty.Value{
						cty.StringVal("a"),
						cty.StringVal("b"),
						cty.StringVal("c"),
					}),
					"secret": cty.StringVal("b").Mark("sensitive"),
				},
			},
			cty.TupleVal([]cty.Value{
				cty.StringVal("a"),
				cty.StringVal("c"),
			}).Mark("sensitive"),
			0,
		},
		{ // Map for loop with marked conditional expression
			`{ for k, v in things: k => v if k != secret }`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"things": cty.MapVal(map[string]cty.Value{
						"a": cty.True,
						"b": cty.False,
						"c": cty.False,
					}),
					"secret": cty.StringVal("b").Mark("sensitive"),
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"a": cty.True,
				"c": cty.False,
			}).Mark("sensitive"),
			0,
		},
		{
			`[{name: "Steve"}, {name: "Ermintrude"}].*.name`,
			nil,
			cty.TupleVal([]cty.Value{
				cty.StringVal("Steve"),
				cty.StringVal("Ermintrude"),
			}),
			0,
		},
		{
			`{name: "Steve"}.*.name`,
			nil,
			cty.TupleVal([]cty.Value{
				cty.StringVal("Steve"),
			}),
			0,
		},
		{
			`null[*]`,
			nil,
			cty.EmptyTupleVal,
			0,
		},
		{
			`{name: "Steve"}[*].name`,
			nil,
			cty.TupleVal([]cty.Value{
				cty.StringVal("Steve"),
			}),
			0,
		},
		{
			`set.*.name`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"set": cty.SetVal([]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"name": cty.StringVal("Steve"),
						}),
					}),
				},
			},
			cty.ListVal([]cty.Value{
				cty.StringVal("Steve"),
			}),
			0,
		},
		{
			`unkstr[*]`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unkstr": cty.UnknownVal(cty.String),
				},
			},
			cty.DynamicVal,
			0,
		},
		{
			`unkstr[*]`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unkstr": cty.UnknownVal(cty.String).RefineNotNull(),
				},
			},
			// If the unknown string is definitely not null then we already
			// know that the result will be a single-element tuple.
			cty.TupleVal([]cty.Value{
				cty.UnknownVal(cty.String).RefineNotNull(),
			}),
			0,
		},
		{
			`unkstr.*.name`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unkstr": cty.UnknownVal(cty.String),
				},
			},
			cty.DynamicVal,
			1, // a string has no attribute "name"
		},
		{
			`dyn.*.name`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"dyn": cty.DynamicVal,
				},
			},
			cty.DynamicVal,
			0,
		},
		{
			`unkobj.*.name`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unkobj": cty.UnknownVal(cty.Object(map[string]cty.Type{
						"name": cty.String,
					})),
				},
			},
			cty.DynamicVal,
			0,
		},
		{
			`unkobj.*.name`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unkobj": cty.UnknownVal(cty.Object(map[string]cty.Type{
						"name": cty.String,
					})).RefineNotNull(),
				},
			},
			cty.TupleVal([]cty.Value{
				cty.UnknownVal(cty.String),
			}),
			0,
		},
		{
			`unkobj.*.names`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unkobj": cty.UnknownVal(cty.Object(map[string]cty.Type{
						"names": cty.List(cty.String),
					})),
				},
			},
			cty.DynamicVal,
			0,
		},
		{
			`unklistobj.*.name`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unklistobj": cty.UnknownVal(cty.List(cty.Object(map[string]cty.Type{
						"name": cty.String,
					}))),
				},
			},
			cty.UnknownVal(cty.List(cty.String)).RefineNotNull(),
			0,
		},
		{
			`unklistobj.*.name`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unklistobj": cty.UnknownVal(cty.List(cty.Object(map[string]cty.Type{
						"name": cty.String,
					}))).Refine().
						CollectionLengthUpperBound(5).
						NewValue(),
				},
			},
			cty.UnknownVal(cty.List(cty.String)).Refine().
				NotNull().
				CollectionLengthUpperBound(5).
				NewValue(),
			0,
		},
		{
			`unktupleobj.*.name`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unktupleobj": cty.UnknownVal(
						cty.Tuple([]cty.Type{
							cty.Object(map[string]cty.Type{
								"name": cty.String,
							}),
							cty.Object(map[string]cty.Type{
								"name": cty.Bool,
							}),
						}),
					),
				},
			},
			cty.UnknownVal(cty.Tuple([]cty.Type{cty.String, cty.Bool})).RefineNotNull(),
			0,
		},
		{
			`nullobj.*.name`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"nullobj": cty.NullVal(cty.Object(map[string]cty.Type{
						"name": cty.String,
					})),
				},
			},
			cty.TupleVal([]cty.Value{}),
			0,
		},
		{
			`nulllist.*.name`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"nulllist": cty.NullVal(cty.List(cty.Object(map[string]cty.Type{
						"name": cty.String,
					}))),
				},
			},
			cty.DynamicVal,
			1, // splat cannot be applied to null sequence
		},
		{
			`["hello", "goodbye"].*`,
			nil,
			cty.TupleVal([]cty.Value{
				cty.StringVal("hello"),
				cty.StringVal("goodbye"),
			}),
			0,
		},
		{
			`"hello".*`,
			nil,
			cty.TupleVal([]cty.Value{
				cty.StringVal("hello"),
			}),
			0,
		},
		{
			`[["hello"], ["world", "unused"]].*.0`,
			nil,
			cty.TupleVal([]cty.Value{
				cty.StringVal("hello"),
				cty.StringVal("world"),
			}),
			0,
		},
		{
			`[[{name:"foo"}], [{name:"bar"}, {name:"baz"}]].*.0.name`,
			nil,
			cty.TupleVal([]cty.Value{
				cty.StringVal("foo"),
				cty.StringVal("bar"),
			}),
			0,
		},
		{
			`[[[{name:"foo"}]], [[{name:"bar"}], [{name:"baz"}]]].*.0.0.name`,
			nil,
			cty.TupleVal([]cty.Value{
				cty.DynamicVal,
				cty.DynamicVal,
			}),
			1, // can't chain legacy index syntax together, like .0.0 (because 0.0 parses as a single number)
		},
		{
			// For an "attribute-only" splat, an index operator applies to
			// the splat result as a whole, rather than being incorporated
			// into the splat traversal itself.
			`[{name: "Steve"}, {name: "Ermintrude"}].*.name[0]`,
			nil,
			cty.StringVal("Steve"),
			0,
		},
		{
			// For a "full" splat, an index operator is consumed as part
			// of the splat's traversal.
			`[{names: ["Steve"]}, {names: ["Ermintrude"]}][*].names[0]`,
			nil,
			cty.TupleVal([]cty.Value{cty.StringVal("Steve"), cty.StringVal("Ermintrude")}),
			0,
		},
		{
			// Another "full" splat, this time with the index first.
			`[[{name: "Steve"}], [{name: "Ermintrude"}]][*][0].name`,
			nil,
			cty.TupleVal([]cty.Value{cty.StringVal("Steve"), cty.StringVal("Ermintrude")}),
			0,
		},
		{
			// Full splats can nest, which produces nested tuples.
			`[[{name: "Steve"}], [{name: "Ermintrude"}]][*][*].name`,
			nil,
			cty.TupleVal([]cty.Value{
				cty.TupleVal([]cty.Value{cty.StringVal("Steve")}),
				cty.TupleVal([]cty.Value{cty.StringVal("Ermintrude")}),
			}),
			0,
		},
		{
			`[["hello"], ["goodbye"]].*.*`,
			nil,
			cty.TupleVal([]cty.Value{
				cty.TupleVal([]cty.Value{cty.StringVal("hello")}),
				cty.TupleVal([]cty.Value{cty.StringVal("goodbye")}),
			}),
			1,
		},
		{ // splat with sensitive collection
			`maps.*.enabled`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"maps": cty.ListVal([]cty.Value{
						cty.MapVal(map[string]cty.Value{"enabled": cty.True}),
						cty.MapVal(map[string]cty.Value{"enabled": cty.False}),
					}).Mark("sensitive"),
				},
			},
			cty.ListVal([]cty.Value{
				cty.True,
				cty.False,
			}).Mark("sensitive"),
			0,
		},
		{ // splat with sensitive collection that's unknown
			`maps.*.enabled`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"maps": cty.UnknownVal(cty.List(cty.Map(cty.Bool))).Mark("sensitive"),
				},
			},
			cty.UnknownVal(cty.List(cty.Bool)).RefineNotNull().Mark("sensitive"),
			0,
		},
		{ // splat with sensitive non-collection that's unknown
			`not_a_list.*`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"not_a_list": cty.UnknownVal(cty.EmptyObject).RefineNotNull().Mark("sensitive"),
				},
			},
			cty.TupleVal([]cty.Value{cty.UnknownVal(cty.EmptyObject).RefineNotNull().Mark("sensitive")}).Mark("sensitive"),
			0,
		},
		{ // splat with sensitive collection that's unknown and not null
			`maps.*.enabled`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"maps": cty.UnknownVal(cty.List(cty.Map(cty.Bool))).RefineNotNull().Mark("sensitive"),
				},
			},
			cty.UnknownVal(cty.List(cty.Bool)).RefineNotNull().Mark("sensitive"),
			0,
		},
		{ // splat with collection with sensitive elements
			`maps.*.x`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"maps": cty.ListVal([]cty.Value{
						cty.MapVal(map[string]cty.Value{
							"x": cty.StringVal("foo").Mark("sensitive"),
						}),
						cty.MapVal(map[string]cty.Value{
							"x": cty.StringVal("bar"),
						}),
					}),
				},
			},
			cty.ListVal([]cty.Value{
				cty.StringVal("foo").Mark("sensitive"),
				cty.StringVal("bar"),
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
			`["hello"].0`,
			nil,
			cty.StringVal("hello"),
			0,
		},
		{
			`[["hello"]].0.0`,
			nil,
			cty.DynamicVal,
			1, // can't chain legacy index syntax together (because 0.0 parses as 0)
		},
		{
			`[{greeting = "hello"}].0.greeting`,
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
			&hcl.EvalContext{
				Functions: map[string]function.Function{
					"negate": stdlib.NegateFunc,
				},
			},
			cty.StringVal("hello"),
			0,
		},
		{
			`[][negate(0)]`,
			&hcl.EvalContext{
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
			`["boop"].foo[index]`, // index is a variable to force IndexExpr instead of traversal
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"index": cty.NumberIntVal(0),
				},
			},
			cty.DynamicVal,
			1, // expression ["boop"] does not have attributes
		},

		{
			`foo`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"foo": cty.StringVal("hello"),
				},
			},
			cty.StringVal("hello"),
			0,
		},
		{
			`bar`,
			&hcl.EvalContext{},
			cty.DynamicVal,
			1, // variables not allowed here
		},
		{
			`foo.bar`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"foo": cty.StringVal("hello"),
				},
			},
			cty.DynamicVal,
			1, // foo does not have attributes
		},
		{
			`foo.baz`,
			&hcl.EvalContext{
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
			&hcl.EvalContext{
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
			&hcl.EvalContext{
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
			&hcl.EvalContext{
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
			`
<<EOT
Foo
Bar
Baz
EOT
`,
			nil,
			cty.StringVal("Foo\nBar\nBaz\n"),
			0,
		},
		{
			`
<<EOT
Foo
${bar}
Baz
EOT
`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"bar": cty.StringVal("Bar"),
				},
			},
			cty.StringVal("Foo\nBar\nBaz\n"),
			0,
		},
		{
			`
<<EOT
Foo
%{for x in bars}${x}%{endfor}
Baz
EOT
`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"bars": cty.ListVal([]cty.Value{
						cty.StringVal("Bar"),
						cty.StringVal("Bar"),
						cty.StringVal("Bar"),
					}),
				},
			},
			cty.StringVal("Foo\nBarBarBar\nBaz\n"),
			0,
		},
		{
			`[
  <<EOT
  Foo
  Bar
  Baz
  EOT
]
`,
			nil,
			cty.TupleVal([]cty.Value{cty.StringVal("  Foo\n  Bar\n  Baz\n")}),
			0,
		},
		{
			`[
  <<-EOT
  Foo
  Bar
  Baz
  EOT
]
`,
			nil,
			cty.TupleVal([]cty.Value{cty.StringVal("Foo\nBar\nBaz\n")}),
			0,
		},
		{
			`[
  <<-EOT
  Foo
    Bar
    Baz
  EOT
]
`,
			nil,
			cty.TupleVal([]cty.Value{cty.StringVal("Foo\n  Bar\n  Baz\n")}),
			0,
		},
		{
			`[
  <<-EOT
    Foo
  Bar
    Baz
  EOT
]
`,
			nil,
			cty.TupleVal([]cty.Value{cty.StringVal("  Foo\nBar\n  Baz\n")}),
			0,
		},
		{
			`[
  <<-EOT
    Foo
  ${bar}
    Baz
    EOT
]
`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"bar": cty.StringVal("  Bar"), // Spaces in the interpolation result don't affect the outcome
				},
			},
			cty.TupleVal([]cty.Value{cty.StringVal("  Foo\n  Bar\n  Baz\n")}),
			0,
		},
		{
			`[
  <<EOT
  Foo

  Bar

  Baz
  EOT
]
`,
			nil,
			cty.TupleVal([]cty.Value{cty.StringVal("  Foo\n\n  Bar\n\n  Baz\n")}),
			0,
		},
		{
			`[
  <<-EOT
  Foo

  Bar

  Baz
  EOT
]
`,
			nil,
			cty.TupleVal([]cty.Value{cty.StringVal("Foo\n\nBar\n\nBaz\n")}),
			0,
		},

		{
			`unk["baz"]`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unk": cty.UnknownVal(cty.String),
				},
			},
			cty.DynamicVal,
			1, // value does not have indices (because we know it's a string)
		},
		{
			`unk["boop"]`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unk": cty.UnknownVal(cty.Map(cty.String)),
				},
			},
			cty.UnknownVal(cty.String), // we know it's a map of string
			0,
		},
		{
			`dyn["boop"]`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"dyn": cty.DynamicVal,
				},
			},
			cty.DynamicVal, // don't know what it is yet
			0,
		},
		{
			`nullstr == "foo"`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"nullstr": cty.NullVal(cty.String),
				},
			},
			cty.False,
			0,
		},
		{
			`nullstr == nullstr`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"nullstr": cty.NullVal(cty.String),
				},
			},
			cty.True,
			0,
		},
		{
			`nullstr == null`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"nullstr": cty.NullVal(cty.String),
				},
			},
			cty.True,
			0,
		},
		{
			`nullstr == nullnum`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"nullstr": cty.NullVal(cty.String),
					"nullnum": cty.NullVal(cty.Number),
				},
			},
			cty.True,
			0,
		},
		{
			`"" == nulldyn`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"nulldyn": cty.NullVal(cty.DynamicPseudoType),
				},
			},
			cty.False,
			0,
		},
		{
			`true ? var : null`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"var": cty.ObjectVal(map[string]cty.Value{"a": cty.StringVal("A")}),
				},
			},
			cty.ObjectVal(map[string]cty.Value{"a": cty.StringVal("A")}),
			0,
		},
		{
			`true ? var : null`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"var": cty.UnknownVal(cty.DynamicPseudoType),
				},
			},
			cty.UnknownVal(cty.DynamicPseudoType),
			0,
		},
		{
			`true ? ["a", "b"] : null`,
			nil,
			cty.TupleVal([]cty.Value{cty.StringVal("a"), cty.StringVal("b")}),
			0,
		},
		{
			`true ? null: ["a", "b"]`,
			nil,
			cty.NullVal(cty.Tuple([]cty.Type{cty.String, cty.String})),
			0,
		},
		{
			`false ? ["a", "b"] : null`,
			nil,
			cty.NullVal(cty.Tuple([]cty.Type{cty.String, cty.String})),
			0,
		},
		{
			`false ? null: ["a", "b"]`,
			nil,
			cty.TupleVal([]cty.Value{cty.StringVal("a"), cty.StringVal("b")}),
			0,
		},
		{
			`false ? null: null`,
			nil,
			cty.NullVal(cty.DynamicPseudoType),
			0,
		},
		{
			`false ? var: {a = "b"}`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"var": cty.DynamicVal,
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"a": cty.StringVal("b"),
			}),
			0,
		},
		{
			`true ? ["a", "b"]: var`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"var": cty.UnknownVal(cty.DynamicPseudoType),
				},
			},
			cty.TupleVal([]cty.Value{
				cty.StringVal("a"),
				cty.StringVal("b"),
			}),
			0,
		},
		{
			`false ? ["a", "b"]: var`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"var": cty.DynamicVal,
				},
			},
			cty.DynamicVal,
			0,
		},
		{
			`false ? ["a", "b"]: var`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"var": cty.UnknownVal(cty.DynamicPseudoType),
				},
			},
			cty.DynamicVal,
			0,
		},
		{
			`unknown ? 1 : 0`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unknown": cty.UnknownVal(cty.Bool),
				},
			},
			cty.UnknownVal(cty.Number).Refine().
				NotNull().
				NumberRangeLowerBound(cty.Zero, true).
				NumberRangeUpperBound(cty.NumberIntVal(1), true).
				NewValue(),
			0,
		},
		{
			`unknown ? 0 : 1`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unknown": cty.UnknownVal(cty.Bool),
				},
			},
			cty.UnknownVal(cty.Number).Refine().
				NotNull().
				NumberRangeLowerBound(cty.Zero, true).
				NumberRangeUpperBound(cty.NumberIntVal(1), true).
				NewValue(),
			0,
		},
		{
			`unknown ? i : j`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unknown": cty.UnknownVal(cty.Bool),
					"i":       cty.NullVal(cty.Number),
					"j":       cty.NullVal(cty.Number),
				},
			},
			cty.NullVal(cty.Number),
			0,
		},
		{
			`unknown ? im : jm`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unknown": cty.UnknownVal(cty.Bool),
					"im":      cty.NullVal(cty.Number).Mark("a"),
					"jm":      cty.NullVal(cty.Number).Mark("b"),
				},
			},
			cty.NullVal(cty.Number).Mark("a").Mark("b"),
			0,
		},
		{
			`unknown ? im : jm`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unknown": cty.UnknownVal(cty.Bool).Mark("a"),
					"im":      cty.UnknownVal(cty.Number),
					"jm":      cty.UnknownVal(cty.Number).Mark("b"),
				},
			},
			// the empty refinement may eventually be removed, but does nothing here
			cty.UnknownVal(cty.Number).Refine().NewValue().Mark("a").Mark("b"),
			0,
		},
		{
			`unknown ? ix : jx`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unknown": cty.UnknownVal(cty.Bool),
					"ix":      cty.UnknownVal(cty.Number),
					"jx":      cty.UnknownVal(cty.Number),
				},
			},
			// the empty refinement may eventually be removed, but does nothing here
			cty.UnknownVal(cty.Number).Refine().NewValue(),
			0,
		},
		{
			`unknown ? ir : jr`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unknown": cty.UnknownVal(cty.Bool),
					"ir": cty.UnknownVal(cty.Number).Refine().
						NumberRangeLowerBound(cty.NumberIntVal(1), false).
						NumberRangeUpperBound(cty.NumberIntVal(3), false).NewValue(),
					"jr": cty.UnknownVal(cty.Number).Refine().
						NumberRangeLowerBound(cty.NumberIntVal(2), true).
						NumberRangeUpperBound(cty.NumberIntVal(4), true).NewValue(),
				},
			},
			cty.UnknownVal(cty.Number).Refine().
				NumberRangeLowerBound(cty.NumberIntVal(1), false).
				NumberRangeUpperBound(cty.NumberIntVal(4), true).NewValue(),
			0,
		},
		{
			`unknown ? a : b`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unknown": cty.UnknownVal(cty.Bool),
					"a":       cty.UnknownVal(cty.Bool).RefineNotNull(),
					"b":       cty.UnknownVal(cty.Bool).RefineNotNull(),
				},
			},
			cty.UnknownVal(cty.Bool).RefineNotNull(),
			0,
		},
		{
			`unknown ? al : bl`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unknown": cty.UnknownVal(cty.Bool),
					"al":      cty.ListValEmpty(cty.String),
					"bl":      cty.ListValEmpty(cty.String),
				},
			},
			cty.ListValEmpty(cty.String), // deduced through refinements
			0,
		},
		{
			`unknown ? am : bm`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unknown": cty.UnknownVal(cty.Bool),
					"am":      cty.MapValEmpty(cty.String),
					"bm":      cty.MapValEmpty(cty.String).Mark("test"),
				},
			},
			cty.MapValEmpty(cty.String).Mark("test"), // deduced through refinements
			0,
		},
		{
			`unknown ? ar : br`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unknown": cty.UnknownVal(cty.Bool),
					"ar": cty.UnknownVal(cty.Set(cty.String)).Refine().
						CollectionLengthLowerBound(1).CollectionLengthUpperBound(3).NewValue(),
					"br": cty.UnknownVal(cty.Set(cty.String)).Refine().
						CollectionLengthLowerBound(2).CollectionLengthUpperBound(4).NewValue(),
				},
			},
			cty.UnknownVal(cty.Set(cty.String)).Refine().CollectionLengthLowerBound(1).CollectionLengthUpperBound(4).NewValue(), // deduced through refinements
			0,
		},
		{
			`unknown ? arn : brn`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unknown": cty.UnknownVal(cty.Bool),
					"arn": cty.UnknownVal(cty.Set(cty.String)).Refine().NotNull().
						CollectionLengthLowerBound(1).CollectionLengthUpperBound(2).NewValue(),
					"brn": cty.UnknownVal(cty.Set(cty.String)).Refine().NotNull().
						CollectionLengthLowerBound(3).CollectionLengthUpperBound(4).NewValue(),
				},
			},
			cty.UnknownVal(cty.Set(cty.String)).Refine().NotNull().CollectionLengthLowerBound(1).CollectionLengthUpperBound(4).NewValue(), // deduced through refinements
			0,
		},
		{
			`unknown ? amr : bmr`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unknown": cty.UnknownVal(cty.Bool),
					"amr": cty.UnknownVal(cty.Set(cty.String)).Mark("test").Refine().
						CollectionLengthLowerBound(1).CollectionLengthUpperBound(2).NewValue(),
					"bmr": cty.UnknownVal(cty.Set(cty.String)).Mark("test").Refine().
						CollectionLengthLowerBound(3).CollectionLengthUpperBound(4).NewValue(),
				},
			},
			cty.UnknownVal(cty.Set(cty.String)).Refine().CollectionLengthLowerBound(1).CollectionLengthUpperBound(4).NewValue().Mark("test"), // deduced through refinements
			0,
		},
		{
			`unknown ? a : b`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unknown": cty.UnknownVal(cty.Bool),
					"a":       cty.ListValEmpty(cty.String),
					"b":       cty.ListVal([]cty.Value{cty.UnknownVal(cty.String)}),
				},
			},
			cty.UnknownVal(cty.List(cty.String)).Refine().
				NotNull().
				CollectionLengthUpperBound(1).
				NewValue(),
			0,
		},
		{
			`unknown ? a : b`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unknown": cty.UnknownVal(cty.Bool),
					"a":       cty.ListVal([]cty.Value{cty.StringVal("hello")}),
					"b":       cty.ListVal([]cty.Value{cty.UnknownVal(cty.String)}),
				},
			},
			cty.ListVal([]cty.Value{cty.UnknownVal(cty.String)}), // deduced through refinements
			0,
		},
		{ // marked conditional
			`var.foo ? 1 : 0`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"var": cty.ObjectVal(map[string]cty.Value{
						"foo": cty.BoolVal(true),
					}).Mark("sensitive"),
				},
			},
			cty.NumberIntVal(1),
			0,
		},
		{ // auto-converts collection types
			`true ? listOf1Tuple : listOf0Tuple`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"listOf1Tuple": cty.ListVal([]cty.Value{cty.TupleVal([]cty.Value{cty.True})}),
					"listOf0Tuple": cty.ListVal([]cty.Value{cty.EmptyTupleVal}),
				},
			},
			cty.ListVal([]cty.Value{cty.ListVal([]cty.Value{cty.True})}),
			0,
		},
		{
			`true ? setOf1Tuple : setOf0Tuple`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"setOf1Tuple": cty.SetVal([]cty.Value{cty.TupleVal([]cty.Value{cty.True})}),
					"setOf0Tuple": cty.SetVal([]cty.Value{cty.EmptyTupleVal}),
				},
			},
			cty.SetVal([]cty.Value{cty.ListVal([]cty.Value{cty.True})}),
			0,
		},
		{ // marked argument expansion
			`min(xs...)`,
			&hcl.EvalContext{
				Functions: map[string]function.Function{
					"min": stdlib.MinFunc,
				},
				Variables: map[string]cty.Value{
					"xs": cty.ListVal([]cty.Value{
						cty.NumberIntVal(3),
						cty.NumberIntVal(1),
						cty.NumberIntVal(4),
					}).Mark("sensitive"),
				},
			},
			cty.NumberIntVal(1).Mark("sensitive"),
			0,
		},
		{
			`test ? sensitiveString : ""`,
			&hcl.EvalContext{
				Functions: map[string]function.Function{},
				Variables: map[string]cty.Value{
					"test":            cty.UnknownVal(cty.Bool),
					"sensitiveString": cty.StringVal("test").Mark("sensitive"),
				},
			},
			cty.UnknownVal(cty.String).RefineNotNull().Mark("sensitive"),
			0,
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			expr, parseDiags := ParseExpression([]byte(test.input), "", hcl.Pos{Line: 1, Column: 1, Byte: 0})
			var got cty.Value
			var valDiags hcl.Diagnostics

			if expr != nil {
				got, valDiags = expr.Value(test.ctx)
			}

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

func TestExpressionErrorMessages(t *testing.T) {
	tests := []struct {
		input       string
		ctx         *hcl.EvalContext
		wantSummary string
		wantDetail  string
	}{
		// Error messages describing inconsistent result types for conditional expressions.
		{
			"true ? 1 : true",
			nil,
			"Inconsistent conditional result types",
			"The true and false result expressions must have consistent types. The 'true' value is number, but the 'false' value is bool.",
		},
		{
			"true ? [1] : [true]",
			nil,
			"Inconsistent conditional result types",
			"The true and false result expressions must have consistent types. Type mismatch for tuple element 0: The 'true' value is number, but the 'false' value is bool.",
		},
		{
			"true ? [1] : [1, true]",
			nil,
			"Inconsistent conditional result types",
			"The true and false result expressions must have consistent types. The 'true' tuple has length 1, but the 'false' tuple has length 2.",
		},
		{
			"true ? { a = 1 } : { a = true }",
			nil,
			"Inconsistent conditional result types",
			"The true and false result expressions must have consistent types. Type mismatch for object attribute \"a\": The 'true' value is number, but the 'false' value is bool.",
		},
		{
			"true ? { a = true, b = 1 } : { a = true }",
			nil,
			"Inconsistent conditional result types",
			"The true and false result expressions must have consistent types. The 'true' value includes object attribute \"b\", which is absent in the 'false' value.",
		},
		{
			"true ? { a = true } : { a = true, b = 1 }",
			nil,
			"Inconsistent conditional result types",
			"The true and false result expressions must have consistent types. The 'false' value includes object attribute \"b\", which is absent in the 'true' value.",
		},
		{
			// Failing cases for automatic collection conversions. HCL and cty
			// will attempt to unify tuples into lists. We have to make sure
			// the tuple inner types have no common base type, so we mix and
			// match booleans and numbers and validate the error messages.
			"true ? listOf2Tuple : listOf1Tuple",
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"listOf2Tuple": cty.ListVal([]cty.Value{cty.TupleVal([]cty.Value{cty.True, cty.Zero})}),
					"listOf1Tuple": cty.ListVal([]cty.Value{cty.TupleVal([]cty.Value{cty.True})}),
				},
			},
			"Inconsistent conditional result types",
			"The true and false result expressions must have consistent types. Mismatched list element types: The 'true' tuple has length 2, but the 'false' tuple has length 1.",
		},
		{
			"true ? setOf2Tuple : setOf1Tuple",
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"setOf2Tuple": cty.SetVal([]cty.Value{cty.TupleVal([]cty.Value{cty.True, cty.Zero})}),
					"setOf1Tuple": cty.SetVal([]cty.Value{cty.TupleVal([]cty.Value{cty.True})}),
				},
			},
			"Inconsistent conditional result types",
			"The true and false result expressions must have consistent types. Mismatched set element types: The 'true' tuple has length 2, but the 'false' tuple has length 1.",
		},
		{
			"true ? mapOf1Tuple : mapOf2Tuple",
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"mapOf1Tuple": cty.MapVal(map[string]cty.Value{"a": cty.TupleVal([]cty.Value{cty.True})}),
					"mapOf2Tuple": cty.MapVal(map[string]cty.Value{"a": cty.TupleVal([]cty.Value{cty.True, cty.Zero})}),
				},
			},
			"Inconsistent conditional result types",
			"The true and false result expressions must have consistent types. Mismatched map element types: The 'true' tuple has length 1, but the 'false' tuple has length 2.",
		},
		{
			"true ? listOfListOf2Tuple : listOfListOf1Tuple",
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"listOfListOf2Tuple": cty.ListVal([]cty.Value{cty.ListVal([]cty.Value{cty.TupleVal([]cty.Value{cty.True, cty.Zero})})}),
					"listOfListOf1Tuple": cty.ListVal([]cty.Value{cty.ListVal([]cty.Value{cty.TupleVal([]cty.Value{cty.True})})}),
				},
			},
			"Inconsistent conditional result types",
			// This is our totally non-specific last-resort of an error message,
			// for situations that are too complex for any of our rules to
			// describe coherently.
			"The true and false result expressions must have consistent types. At least one deeply-nested attribute or element is not compatible across both the 'true' and the 'false' value.",
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			var diags hcl.Diagnostics
			expr, parseDiags := ParseExpression([]byte(test.input), "", hcl.Pos{Line: 1, Column: 1, Byte: 0})
			diags = append(diags, parseDiags...)
			_, valDiags := expr.Value(test.ctx)
			diags = append(diags, valDiags...)

			if !diags.HasErrors() {
				t.Fatalf("unexpected success\nwant error:\n%s; %s", test.wantSummary, test.wantDetail)
			}

			for _, diag := range diags {
				if diag.Severity != hcl.DiagError {
					continue
				}
				if diag.Summary == test.wantSummary && diag.Detail == test.wantDetail {
					// Success! We'll return early to conclude this test case.
					return
				}
			}
			// If we fall out here then we didn't find the diagnostic
			// we were looking for.
			t.Fatalf("missing expected error\ngot:\n%s\n\nwant error:\n%s; %s", diags.Error(), test.wantSummary, test.wantDetail)
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
		ctx       *hcl.EvalContext
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
			&hcl.EvalContext{
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
			&hcl.EvalContext{
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
			&hcl.EvalContext{
				Functions: funcs,
			},
			cty.UnknownVal(cty.Number).Refine().NotNull().NumberRangeLowerBound(cty.NumberIntVal(0), true).NewValue(),
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
			&hcl.EvalContext{
				Functions: funcs,
			},
			cty.UnknownVal(cty.Number).Refine().NotNull().NumberRangeLowerBound(cty.NumberIntVal(0), true).NewValue(),
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
			&hcl.EvalContext{
				Functions: funcs,
			},
			cty.UnknownVal(cty.Number).Refine().NotNull().NumberRangeLowerBound(cty.NumberIntVal(0), true).NewValue(),
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
			&hcl.EvalContext{
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
			&hcl.EvalContext{
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
			&hcl.EvalContext{
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
			&hcl.EvalContext{
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
			&hcl.EvalContext{
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

func TestExpressionAsTraversal(t *testing.T) {
	expr, _ := ParseExpression([]byte("a.b[0][\"c\"]"), "", hcl.Pos{})
	traversal, diags := hcl.AbsTraversalForExpr(expr)
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics:\n%s", diags.Error())
	}
	if len(traversal) != 4 {
		t.Fatalf("wrong traversal %#v; want length 3", traversal)
	}
	if traversal.RootName() != "a" {
		t.Errorf("wrong root name %q; want %q", traversal.RootName(), "a")
	}
	if step, ok := traversal[1].(hcl.TraverseAttr); ok {
		if got, want := step.Name, "b"; got != want {
			t.Errorf("wrong name %q for step 1; want %q", got, want)
		}
	} else {
		t.Errorf("wrong type %T for step 1; want %T", traversal[1], step)
	}
	if step, ok := traversal[2].(hcl.TraverseIndex); ok {
		if got, want := step.Key, cty.Zero; !want.RawEquals(got) {
			t.Errorf("wrong name %#v for step 2; want %#v", got, want)
		}
	} else {
		t.Errorf("wrong type %T for step 2; want %T", traversal[2], step)
	}
	if step, ok := traversal[3].(hcl.TraverseIndex); ok {
		if got, want := step.Key, cty.StringVal("c"); !want.RawEquals(got) {
			t.Errorf("wrong name %#v for step 3; want %#v", got, want)
		}
	} else {
		t.Errorf("wrong type %T for step 3; want %T", traversal[3], step)
	}
}

func TestStaticExpressionList(t *testing.T) {
	expr, _ := ParseExpression([]byte("[0, a, true]"), "", hcl.Pos{})
	exprs, diags := hcl.ExprList(expr)
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics:\n%s", diags.Error())
	}
	if len(exprs) != 3 {
		t.Fatalf("wrong result %#v; want length 3", exprs)
	}
	first, ok := exprs[0].(*LiteralValueExpr)
	if !ok {
		t.Fatalf("first expr has wrong type %T; want *hclsyntax.LiteralValueExpr", exprs[0])
	}
	if !first.Val.RawEquals(cty.Zero) {
		t.Fatalf("wrong first value %#v; want cty.Zero", first.Val)
	}
}

// Check that function call w/ incomplete argument still reports correct range
func TestParseExpression_incompleteFunctionCall(t *testing.T) {
	tests := []struct {
		cfg           string
		expectedRange hcl.Range
	}{
		{
			`object({ foo = })`,
			hcl.Range{
				Filename: "test.hcl",
				Start:    hcl.InitialPos,
				End:      hcl.Pos{Line: 1, Column: 18, Byte: 17},
			},
		},
		{
			`object({
  foo =
})`,
			hcl.Range{
				Filename: "test.hcl",
				Start:    hcl.InitialPos,
				End:      hcl.Pos{Line: 3, Column: 3, Byte: 19},
			},
		},
		{
			`object({ foo = }`,
			hcl.Range{
				Filename: "test.hcl",
				Start:    hcl.InitialPos,
				End:      hcl.Pos{Line: 0, Column: 0, Byte: 0},
			},
		},
		{
			`object({
  foo =
}`,
			hcl.Range{
				Filename: "test.hcl",
				Start:    hcl.InitialPos,
				End:      hcl.Pos{Line: 0, Column: 0, Byte: 0},
			},
		},
		{
			`object({
  foo =
`,
			hcl.Range{
				Filename: "test.hcl",
				Start:    hcl.InitialPos,
				End:      hcl.Pos{Line: 0, Column: 0, Byte: 0},
			},
		},
		{
			`object({
  foo =
)`,
			hcl.Range{
				Filename: "test.hcl",
				Start:    hcl.InitialPos,
				End:      hcl.Pos{Line: 0, Column: 0, Byte: 0},
			},
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			expr, _ := ParseExpression([]byte(tc.cfg), "test.hcl", hcl.InitialPos)
			if diff := cmp.Diff(tc.expectedRange, expr.Range()); diff != "" {
				t.Fatalf("range mismatch: %s", diff)
			}
		})
	}
}
