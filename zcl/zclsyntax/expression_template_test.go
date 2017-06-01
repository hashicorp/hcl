package zclsyntax

import (
	"testing"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-zcl/zcl"
)

func TestTemplateExprParseAndValue(t *testing.T) {
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
			cty.StringVal("1"),
			0,
		},
		{
			`(1)`,
			nil,
			cty.StringVal("(1)"),
			0,
		},
		{
			`true`,
			nil,
			cty.StringVal("true"),
			0,
		},
		{
			`
hello world
`,
			nil,
			cty.StringVal("\nhello world\n"),
			0,
		},
		{
			`hello ${"world"}`,
			nil,
			cty.StringVal("hello world"),
			0,
		},
		{
			`hello\nworld`, // backslash escapes not supported in bare templates
			nil,
			cty.StringVal("hello\\nworld"),
			0,
		},
		{
			`hello ${12.5}`,
			nil,
			cty.StringVal("hello 12.5"),
			0,
		},
		{
			`silly ${"${"nesting"}"}`,
			nil,
			cty.StringVal("silly nesting"),
			0,
		},
		{
			`silly ${"${true}"}`,
			nil,
			cty.StringVal("silly true"),
			0,
		},
		{
			`hello $${escaped}`,
			nil,
			cty.StringVal("hello ${escaped}"),
			0,
		},
		{
			`hello $$nonescape`,
			nil,
			cty.StringVal("hello $$nonescape"),
			0,
		},
		{
			`${true}`,
			nil,
			cty.True, // any single expression is unwrapped without stringification
			0,
		},
		{
			`trim ${~ "trim"}`,
			nil,
			cty.StringVal("trimtrim"),
			0,
		},
		{
			`${"trim" ~} trim`,
			nil,
			cty.StringVal("trimtrim"),
			0,
		},
		{
			`trim
${~"trim"~}
trim`,
			nil,
			cty.StringVal("trimtrimtrim"),
			0,
		},
		{
			` ${~ true ~} `,
			nil,
			cty.StringVal("true"), // can't trim space to reduce to a single expression
			0,
		},
		{
			`${"hello "}${~"trim"~}${" hello"}`,
			nil,
			cty.StringVal("hello trim hello"), // trimming can't reach into a neighboring interpolation
			0,
		},
		{
			`${true}${~"trim"~}${true}`,
			nil,
			cty.StringVal("truetrimtrue"), // trimming is no-op of neighbors aren't literal strings
			0,
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			expr, parseDiags := ParseTemplate([]byte(test.input), "", zcl.Pos{Line: 1, Column: 1, Byte: 0})

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
