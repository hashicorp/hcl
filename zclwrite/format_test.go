package zclwrite

import (
	"fmt"
	"testing"

	"reflect"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/hcl2/zcl/zclsyntax"
)

func TestFormat(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			``,
			``,
		},
		{
			`a=1`,
			`a = 1`,
		},
		{
			`a=b.c`,
			`a = b.c`,
		},
		{
			`( a+2 )`,
			`(a + 2)`,
		},
		{
			`( a*2 )`,
			`(a * 2)`,
		},
		{
			`( a+-2 )`,
			`(a + -2)`,
		},
		{
			`( a*-2 )`,
			`(a * -2)`,
		},
		{
			`(-2+1)`,
			`(-2 + 1)`,
		},
		{
			`foo(1, -2,a*b, b,c)`,
			`foo(1, -2, a * b, b, c)`,
		},
		{
			`a="hello ${ name }"`,
			`a = "hello ${name}"`,
		},
		{
			`a="hello ${~ name ~}"`,
			`a = "hello ${~name~}"`,
		},
		{
			`b{}`,
			`b {}`,
		},
		{
			`
"${
hello
}"
`,
			`
"${
  hello
}"
`,
		},
		{
			`
foo(
1,
- 2,
a*b,
b,
c,
)
`,
			`
foo(
  1,
  -2,
  a * b,
  b,
  c,
)
`,
		},
		{
			`[ [ ] ]`,
			`[[]]`,
		},
		{
			`
[
[
a
]
]
`,
			`
[
  [
    a
  ]
]
`,
		},
		{
			`
[[
a
]]
`,
			`
[[
  a
]]
`,
		},
		{
			`
[[
[
a
]
]]
`,
			`
[[
  [
    a
  ]
]]
`,
		},
		{
			// degenerate case with asymmetrical brackets
			`
[[
[
a
]]
]
`,
			`
[[
  [
    a
  ]]
]
`,
		},
		{
			`
b {
a = 1
}
`,
			`
b {
  a = 1
}
`,
		},
		{
			`
a = 1
bungle = 2
`,
			`
a      = 1
bungle = 2
`,
		},
		{
			`
a = 1

bungle = 2
`,
			`
a = 1

bungle = 2
`,
		},
		{
			`
a = 1 # foo
bungle = 2
`,
			`
a      = 1 # foo
bungle = 2
`,
		},
		{
			`
a = 1 # foo
bungle = "bonce" # baz
`,
			`
a      = 1       # foo
bungle = "bonce" # baz
`,
		},
		{
			`
# here we go
a = 1 # foo
bungle = "bonce" # baz
`,
			`
# here we go
a      = 1       # foo
bungle = "bonce" # baz
`,
		},
		{
			`
foo {} # here we go
a = 1 # foo
bungle = "bonce" # baz
`,
			`
foo {}           # here we go
a      = 1       # foo
bungle = "bonce" # baz
`,
		},
		{
			`
a = 1 # foo
bungle = "bonce" # baz
zebra = "striped" # baz
`,
			`
a      = 1         # foo
bungle = "bonce"   # baz
zebra  = "striped" # baz
`,
		},
		{
			`
a = 1 # foo
bungle = (
    "bonce"
) # baz
zebra = "striped" # baz
`,
			`
a = 1 # foo
bungle = (
  "bonce"
)                 # baz
zebra = "striped" # baz
`,
		},
		{
			`
a="apple"# foo
bungle=(# woo parens
"bonce"
)# baz
zebra="striped"# baz
`,
			`
a = "apple" # foo
bungle = (  # woo parens
  "bonce"
)                 # baz
zebra = "striped" # baz
`,
		},
		{
			`
𝒜 = 1 # foo
bungle = "🇬🇧" # baz
zebra = "striped" # baz
`,
			`
𝒜      = 1         # foo
bungle = "🇬🇧"       # baz
zebra  = "striped" # baz
`,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			tokens := lexConfig([]byte(test.input))
			format(tokens)
			t.Logf("tokens %s\n", spew.Sdump(tokens))
			got := string(tokens.Bytes())

			if got != test.want {
				t.Errorf("wrong result\ninput:\n%s\ngot:\n%s\nwant:\n%s", test.input, got, test.want)
			}
		})
	}

}

func TestLinesForFormat(t *testing.T) {
	tests := []struct {
		tokens Tokens
		want   []formatLine
	}{
		{
			Tokens{
				&Token{Type: zclsyntax.TokenEOF},
			},
			[]formatLine{
				{
					lead: Tokens{},
				},
			},
		},
		{
			Tokens{
				&Token{Type: zclsyntax.TokenIdent},
				&Token{Type: zclsyntax.TokenEOF},
			},
			[]formatLine{
				{
					lead: Tokens{
						&Token{Type: zclsyntax.TokenIdent},
					},
				},
			},
		},
		{
			Tokens{
				&Token{Type: zclsyntax.TokenIdent},
				&Token{Type: zclsyntax.TokenNewline},
				&Token{Type: zclsyntax.TokenNumberLit},
				&Token{Type: zclsyntax.TokenEOF},
			},
			[]formatLine{
				{
					lead: Tokens{
						&Token{Type: zclsyntax.TokenIdent},
						&Token{Type: zclsyntax.TokenNewline},
					},
				},
				{
					lead: Tokens{
						&Token{Type: zclsyntax.TokenNumberLit},
					},
				},
			},
		},
		{
			Tokens{
				&Token{Type: zclsyntax.TokenIdent},
				&Token{Type: zclsyntax.TokenComment, Bytes: []byte("#foo\n")},
				&Token{Type: zclsyntax.TokenNumberLit},
				&Token{Type: zclsyntax.TokenEOF},
			},
			[]formatLine{
				{
					lead: Tokens{
						&Token{Type: zclsyntax.TokenIdent},
					},
					comment: Tokens{
						&Token{Type: zclsyntax.TokenComment, Bytes: []byte("#foo\n")},
					},
				},
				{
					lead: Tokens{
						&Token{Type: zclsyntax.TokenNumberLit},
					},
				},
			},
		},
		{
			Tokens{
				&Token{Type: zclsyntax.TokenIdent},
				&Token{Type: zclsyntax.TokenEqual},
				&Token{Type: zclsyntax.TokenNumberLit},
				&Token{Type: zclsyntax.TokenEOF},
			},
			[]formatLine{
				{
					lead: Tokens{
						&Token{Type: zclsyntax.TokenIdent},
					},
					assign: Tokens{
						&Token{Type: zclsyntax.TokenEqual},
						&Token{Type: zclsyntax.TokenNumberLit},
					},
				},
			},
		},
		{
			Tokens{
				&Token{Type: zclsyntax.TokenIdent},
				&Token{Type: zclsyntax.TokenEqual},
				&Token{Type: zclsyntax.TokenNumberLit},
				&Token{Type: zclsyntax.TokenComment, Bytes: []byte("#foo\n")},
				&Token{Type: zclsyntax.TokenEOF},
			},
			[]formatLine{
				{
					lead: Tokens{
						&Token{Type: zclsyntax.TokenIdent},
					},
					assign: Tokens{
						&Token{Type: zclsyntax.TokenEqual},
						&Token{Type: zclsyntax.TokenNumberLit},
					},
					comment: Tokens{
						&Token{Type: zclsyntax.TokenComment, Bytes: []byte("#foo\n")},
					},
				},
				{
					lead: Tokens{},
				},
			},
		},
		{
			Tokens{
				// A comment goes into a comment cell only if it is after
				// some non-comment tokens, since whole-line comments must
				// stay flush with the indent level.
				&Token{Type: zclsyntax.TokenComment, Bytes: []byte("#foo\n")},
				&Token{Type: zclsyntax.TokenEOF},
			},
			[]formatLine{
				{
					lead: Tokens{
						&Token{Type: zclsyntax.TokenComment, Bytes: []byte("#foo\n")},
					},
				},
				{
					lead: Tokens{},
				},
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			got := linesForFormat(test.tokens)

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.want)
			}
		})
	}
}
