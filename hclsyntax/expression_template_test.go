package hclsyntax

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

func TestTemplateExprParseAndValue(t *testing.T) {
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
			`hello %${"world"}`,
			nil,
			cty.StringVal("hello %world"),
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

		{
			`%{ if true ~} hello %{~ endif }`,
			nil,
			cty.StringVal("hello"),
			0,
		},
		{
			`%{ if false ~} hello %{~ endif}`,
			nil,
			cty.StringVal(""),
			0,
		},
		{
			`%{ if true ~} hello %{~ else ~} goodbye %{~ endif }`,
			nil,
			cty.StringVal("hello"),
			0,
		},
		{
			`%{ if false ~} hello %{~ else ~} goodbye %{~ endif }`,
			nil,
			cty.StringVal("goodbye"),
			0,
		},
		{
			`%{ if true ~} %{~ if false ~} hello %{~ else ~} goodbye %{~ endif ~} %{~ endif }`,
			nil,
			cty.StringVal("goodbye"),
			0,
		},
		{
			`%{ if false ~} %{~ if false ~} hello %{~ else ~} goodbye %{~ endif ~} %{~ endif }`,
			nil,
			cty.StringVal(""),
			0,
		},
		{
			`%{ of true ~} hello %{~ endif}`,
			nil,
			cty.UnknownVal(cty.String),
			2, // "of" is not a valid control keyword, and "endif" is therefore also unexpected
		},
		{
			`%{ for v in ["a", "b", "c"] }${v}%{ endfor }`,
			nil,
			cty.StringVal("abc"),
			0,
		},
		{
			`%{ for v in ["a", "b", "c"] } ${v} %{ endfor }`,
			nil,
			cty.StringVal(" a  b  c "),
			0,
		},
		{
			`%{ for v in ["a", "b", "c"] ~} ${v} %{~ endfor }`,
			nil,
			cty.StringVal("abc"),
			0,
		},
		{
			`%{ for v in [] }${v}%{ endfor }`,
			nil,
			cty.StringVal(""),
			0,
		},
		{
			`%{ for i, v in ["a", "b", "c"] }${i}${v}%{ endfor }`,
			nil,
			cty.StringVal("0a1b2c"),
			0,
		},
		{
			`%{ for k, v in {"A" = "a", "B" = "b", "C" = "c"} }${k}${v}%{ endfor }`,
			nil,
			cty.StringVal("AaBbCc"),
			0,
		},
		{
			`%{ for v in ["a", "b", "c"] }${v}${nl}%{ endfor }`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"nl": cty.StringVal("\n"),
				},
			},
			cty.StringVal("a\nb\nc\n"),
			0,
		},
		{
			`\n`, // backslash escapes are not interpreted in template literals
			nil,
			cty.StringVal("\\n"),
			0,
		},
		{
			`\uu1234`, // backslash escapes are not interpreted in template literals
			nil,       // (this is intentionally an invalid one to ensure we don't produce an error)
			cty.StringVal("\\uu1234"),
			0,
		},
		{
			`$`,
			nil,
			cty.StringVal("$"),
			0,
		},
		{
			`$$`,
			nil,
			cty.StringVal("$$"),
			0,
		},
		{
			`%`,
			nil,
			cty.StringVal("%"),
			0,
		},
		{
			`%%`,
			nil,
			cty.StringVal("%%"),
			0,
		},
		{
			`hello %%{ if true }world%%{ endif }`,
			nil,
			cty.StringVal(`hello %{ if true }world%{ endif }`),
			0,
		},
		{
			`hello $%{ if true }world%{ endif }`,
			nil,
			cty.StringVal("hello $world"),
			0,
		},
		{
			`%{ endif }`,
			nil,
			cty.UnknownVal(cty.String),
			1, // Unexpected endif directive
		},
		{
			`%{ endfor }`,
			nil,
			cty.UnknownVal(cty.String),
			1, // Unexpected endfor directive
		},
		{ // marks from uninterpolated values are ignored
			`hello%{ if false } ${target}%{ endif }`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"target": cty.StringVal("world").Mark("sensitive"),
				},
			},
			cty.StringVal("hello"),
			0,
		},
		{ // marks from interpolated values are passed through
			`${greeting} ${target}`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"greeting": cty.StringVal("hello").Mark("english"),
					"target":   cty.StringVal("world").Mark("sensitive"),
				},
			},
			cty.StringVal("hello world").WithMarks(cty.NewValueMarks("english", "sensitive")),
			0,
		},
		{ // can use marks by traversing complex values
			`Authenticate with "${secrets.passphrase}"`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"secrets": cty.MapVal(map[string]cty.Value{
						"passphrase": cty.StringVal("my voice is my passport").Mark("sensitive"),
					}).Mark("sensitive"),
				},
			},
			cty.StringVal(`Authenticate with "my voice is my passport"`).WithMarks(cty.NewValueMarks("sensitive")),
			0,
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			expr, parseDiags := ParseTemplate([]byte(test.input), "", hcl.Pos{Line: 1, Column: 1, Byte: 0})

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
