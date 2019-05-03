package hclsyntax

import (
	"reflect"
	"testing"

	"github.com/go-test/deep"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/kylelemons/godebug/pretty"
)

func TestScanTokens_normal(t *testing.T) {
	tests := []struct {
		input string
		want  []Token
	}{
		// Empty input
		{
			``,
			[]Token{
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 0, Line: 1, Column: 1},
					},
				},
			},
		},
		{
			` `,
			[]Token{
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
			},
		},
		{
			"\n\n",
			[]Token{
				{
					Type:  TokenNewline,
					Bytes: []byte("\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 2, Column: 1},
					},
				},
				{
					Type:  TokenNewline,
					Bytes: []byte("\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 2, Column: 1},
						End:   hcl.Pos{Byte: 2, Line: 3, Column: 1},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 2, Line: 3, Column: 1},
						End:   hcl.Pos{Byte: 2, Line: 3, Column: 1},
					},
				},
			},
		},

		// Byte-order mark
		{
			"\xef\xbb\xbf", // Leading UTF-8 byte-order mark is ignored...
			[]Token{
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{ // ...but its bytes still count when producing ranges
						Start: hcl.Pos{Byte: 3, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 3, Line: 1, Column: 1},
					},
				},
			},
		},
		{
			" \xef\xbb\xbf", // Non-leading BOM is invalid
			[]Token{
				{
					Type:  TokenInvalid,
					Bytes: utf8BOM,
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 4, Line: 1, Column: 3},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 4, Line: 1, Column: 3},
						End:   hcl.Pos{Byte: 4, Line: 1, Column: 3},
					},
				},
			},
		},
		{
			"\xfe\xff", // UTF-16 BOM is invalid
			[]Token{
				{
					Type:  TokenBadUTF8,
					Bytes: []byte{0xfe},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenBadUTF8,
					Bytes: []byte{0xff},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 2, Line: 1, Column: 3},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 2, Line: 1, Column: 3},
						End:   hcl.Pos{Byte: 2, Line: 1, Column: 3},
					},
				},
			},
		},

		// TokenNumberLit
		{
			`1`,
			[]Token{
				{
					Type:  TokenNumberLit,
					Bytes: []byte(`1`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
			},
		},
		{
			`12`,
			[]Token{
				{
					Type:  TokenNumberLit,
					Bytes: []byte(`12`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 2, Line: 1, Column: 3},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 2, Line: 1, Column: 3},
						End:   hcl.Pos{Byte: 2, Line: 1, Column: 3},
					},
				},
			},
		},
		{
			`12.3`,
			[]Token{
				{
					Type:  TokenNumberLit,
					Bytes: []byte(`12.3`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 4, Line: 1, Column: 5},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 4, Line: 1, Column: 5},
						End:   hcl.Pos{Byte: 4, Line: 1, Column: 5},
					},
				},
			},
		},
		{
			`1e2`,
			[]Token{
				{
					Type:  TokenNumberLit,
					Bytes: []byte(`1e2`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 3, Line: 1, Column: 4},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 3, Line: 1, Column: 4},
						End:   hcl.Pos{Byte: 3, Line: 1, Column: 4},
					},
				},
			},
		},
		{
			`1e+2`,
			[]Token{
				{
					Type:  TokenNumberLit,
					Bytes: []byte(`1e+2`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 4, Line: 1, Column: 5},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 4, Line: 1, Column: 5},
						End:   hcl.Pos{Byte: 4, Line: 1, Column: 5},
					},
				},
			},
		},

		// TokenIdent
		{
			`hello`,
			[]Token{
				{
					Type:  TokenIdent,
					Bytes: []byte(`hello`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 5, Line: 1, Column: 6},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 5, Line: 1, Column: 6},
						End:   hcl.Pos{Byte: 5, Line: 1, Column: 6},
					},
				},
			},
		},
		{
			`_ello`,
			[]Token{
				{
					Type:  TokenIdent,
					Bytes: []byte(`_ello`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 5, Line: 1, Column: 6},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 5, Line: 1, Column: 6},
						End:   hcl.Pos{Byte: 5, Line: 1, Column: 6},
					},
				},
			},
		},
		{
			`hel_o`,
			[]Token{
				{
					Type:  TokenIdent,
					Bytes: []byte(`hel_o`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 5, Line: 1, Column: 6},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 5, Line: 1, Column: 6},
						End:   hcl.Pos{Byte: 5, Line: 1, Column: 6},
					},
				},
			},
		},
		{
			`hel-o`,
			[]Token{
				{
					Type:  TokenIdent,
					Bytes: []byte(`hel-o`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 5, Line: 1, Column: 6},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 5, Line: 1, Column: 6},
						End:   hcl.Pos{Byte: 5, Line: 1, Column: 6},
					},
				},
			},
		},
		{
			`h3ll0`,
			[]Token{
				{
					Type:  TokenIdent,
					Bytes: []byte(`h3ll0`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 5, Line: 1, Column: 6},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 5, Line: 1, Column: 6},
						End:   hcl.Pos{Byte: 5, Line: 1, Column: 6},
					},
				},
			},
		},
		{
			`héllo`, // combining acute accent
			[]Token{
				{
					Type:  TokenIdent,
					Bytes: []byte(`héllo`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 7, Line: 1, Column: 6},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 7, Line: 1, Column: 6},
						End:   hcl.Pos{Byte: 7, Line: 1, Column: 6},
					},
				},
			},
		},

		// Literal-only Templates (string literals, effectively)
		{
			`""`,
			[]Token{
				{
					Type:  TokenOQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenCQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 2, Line: 1, Column: 3},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 2, Line: 1, Column: 3},
						End:   hcl.Pos{Byte: 2, Line: 1, Column: 3},
					},
				},
			},
		},
		{
			`"hello"`,
			[]Token{
				{
					Type:  TokenOQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenQuotedLit,
					Bytes: []byte(`hello`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 6, Line: 1, Column: 7},
					},
				},
				{
					Type:  TokenCQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 6, Line: 1, Column: 7},
						End:   hcl.Pos{Byte: 7, Line: 1, Column: 8},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 7, Line: 1, Column: 8},
						End:   hcl.Pos{Byte: 7, Line: 1, Column: 8},
					},
				},
			},
		},
		{
			`"hello, \"world\"!"`,
			[]Token{
				{
					Type:  TokenOQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenQuotedLit,
					Bytes: []byte(`hello, \"world\"!`), // The escapes are handled by the parser, not the scanner
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 18, Line: 1, Column: 19},
					},
				},
				{
					Type:  TokenCQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 18, Line: 1, Column: 19},
						End:   hcl.Pos{Byte: 19, Line: 1, Column: 20},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 19, Line: 1, Column: 20},
						End:   hcl.Pos{Byte: 19, Line: 1, Column: 20},
					},
				},
			},
		},
		{
			`"hello $$"`,
			[]Token{
				{
					Type:  TokenOQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenQuotedLit,
					Bytes: []byte(`hello `),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 7, Line: 1, Column: 8},
					},
				},
				// This one scans a little oddly because of how the scanner
				// handles the escaping of the dollar sign, but it's still
				// good enough for the parser since it'll just concatenate
				// these two string literals together anyway.
				{
					Type:  TokenQuotedLit,
					Bytes: []byte(`$`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 7, Line: 1, Column: 8},
						End:   hcl.Pos{Byte: 8, Line: 1, Column: 9},
					},
				},
				{
					Type:  TokenQuotedLit,
					Bytes: []byte(`$`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 8, Line: 1, Column: 9},
						End:   hcl.Pos{Byte: 9, Line: 1, Column: 10},
					},
				},
				{
					Type:  TokenCQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 9, Line: 1, Column: 10},
						End:   hcl.Pos{Byte: 10, Line: 1, Column: 11},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 10, Line: 1, Column: 11},
						End:   hcl.Pos{Byte: 10, Line: 1, Column: 11},
					},
				},
			},
		},
		{
			`"hello %%"`,
			[]Token{
				{
					Type:  TokenOQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenQuotedLit,
					Bytes: []byte(`hello `),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 7, Line: 1, Column: 8},
					},
				},
				// This one scans a little oddly because of how the scanner
				// handles the escaping of the percent sign, but it's still
				// good enough for the parser since it'll just concatenate
				// these two string literals together anyway.
				{
					Type:  TokenQuotedLit,
					Bytes: []byte(`%`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 7, Line: 1, Column: 8},
						End:   hcl.Pos{Byte: 8, Line: 1, Column: 9},
					},
				},
				{
					Type:  TokenQuotedLit,
					Bytes: []byte(`%`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 8, Line: 1, Column: 9},
						End:   hcl.Pos{Byte: 9, Line: 1, Column: 10},
					},
				},
				{
					Type:  TokenCQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 9, Line: 1, Column: 10},
						End:   hcl.Pos{Byte: 10, Line: 1, Column: 11},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 10, Line: 1, Column: 11},
						End:   hcl.Pos{Byte: 10, Line: 1, Column: 11},
					},
				},
			},
		},
		{
			`"hello $"`,
			[]Token{
				{
					Type:  TokenOQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenQuotedLit,
					Bytes: []byte(`hello `),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 7, Line: 1, Column: 8},
					},
				},
				{
					Type:  TokenQuotedLit,
					Bytes: []byte(`$`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 7, Line: 1, Column: 8},
						End:   hcl.Pos{Byte: 8, Line: 1, Column: 9},
					},
				},
				{
					Type:  TokenCQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 8, Line: 1, Column: 9},
						End:   hcl.Pos{Byte: 9, Line: 1, Column: 10},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 9, Line: 1, Column: 10},
						End:   hcl.Pos{Byte: 9, Line: 1, Column: 10},
					},
				},
			},
		},
		{
			`"hello %"`,
			[]Token{
				{
					Type:  TokenOQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenQuotedLit,
					Bytes: []byte(`hello `),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 7, Line: 1, Column: 8},
					},
				},
				{
					Type:  TokenQuotedLit,
					Bytes: []byte(`%`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 7, Line: 1, Column: 8},
						End:   hcl.Pos{Byte: 8, Line: 1, Column: 9},
					},
				},
				{
					Type:  TokenCQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 8, Line: 1, Column: 9},
						End:   hcl.Pos{Byte: 9, Line: 1, Column: 10},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 9, Line: 1, Column: 10},
						End:   hcl.Pos{Byte: 9, Line: 1, Column: 10},
					},
				},
			},
		},
		{
			`"hello $${world}"`,
			[]Token{
				{
					Type:  TokenOQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenQuotedLit,
					Bytes: []byte(`hello `),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 7, Line: 1, Column: 8},
					},
				},
				{
					Type:  TokenQuotedLit,
					Bytes: []byte(`$${`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 7, Line: 1, Column: 8},
						End:   hcl.Pos{Byte: 10, Line: 1, Column: 11},
					},
				},
				{
					Type:  TokenQuotedLit,
					Bytes: []byte(`world}`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 10, Line: 1, Column: 11},
						End:   hcl.Pos{Byte: 16, Line: 1, Column: 17},
					},
				},
				{
					Type:  TokenCQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 16, Line: 1, Column: 17},
						End:   hcl.Pos{Byte: 17, Line: 1, Column: 18},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 17, Line: 1, Column: 18},
						End:   hcl.Pos{Byte: 17, Line: 1, Column: 18},
					},
				},
			},
		},
		{
			`"hello %%{world}"`,
			[]Token{
				{
					Type:  TokenOQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenQuotedLit,
					Bytes: []byte(`hello `),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 7, Line: 1, Column: 8},
					},
				},
				{
					Type:  TokenQuotedLit,
					Bytes: []byte(`%%{`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 7, Line: 1, Column: 8},
						End:   hcl.Pos{Byte: 10, Line: 1, Column: 11},
					},
				},
				{
					Type:  TokenQuotedLit,
					Bytes: []byte(`world}`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 10, Line: 1, Column: 11},
						End:   hcl.Pos{Byte: 16, Line: 1, Column: 17},
					},
				},
				{
					Type:  TokenCQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 16, Line: 1, Column: 17},
						End:   hcl.Pos{Byte: 17, Line: 1, Column: 18},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 17, Line: 1, Column: 18},
						End:   hcl.Pos{Byte: 17, Line: 1, Column: 18},
					},
				},
			},
		},
		{
			`"hello %${world}"`,
			[]Token{
				{
					Type:  TokenOQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenQuotedLit,
					Bytes: []byte(`hello `),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 7, Line: 1, Column: 8},
					},
				},
				{
					Type:  TokenQuotedLit,
					Bytes: []byte(`%`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 7, Line: 1, Column: 8},
						End:   hcl.Pos{Byte: 8, Line: 1, Column: 9},
					},
				},
				{
					Type:  TokenTemplateInterp,
					Bytes: []byte(`${`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 8, Line: 1, Column: 9},
						End:   hcl.Pos{Byte: 10, Line: 1, Column: 11},
					},
				},
				{
					Type:  TokenIdent,
					Bytes: []byte(`world`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 10, Line: 1, Column: 11},
						End:   hcl.Pos{Byte: 15, Line: 1, Column: 16},
					},
				},
				{
					Type:  TokenTemplateSeqEnd,
					Bytes: []byte(`}`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 15, Line: 1, Column: 16},
						End:   hcl.Pos{Byte: 16, Line: 1, Column: 17},
					},
				},
				{
					Type:  TokenCQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 16, Line: 1, Column: 17},
						End:   hcl.Pos{Byte: 17, Line: 1, Column: 18},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 17, Line: 1, Column: 18},
						End:   hcl.Pos{Byte: 17, Line: 1, Column: 18},
					},
				},
			},
		},

		// Templates with interpolations and control sequences
		{
			`"${1}"`,
			[]Token{
				{
					Type:  TokenOQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenTemplateInterp,
					Bytes: []byte(`${`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 3, Line: 1, Column: 4},
					},
				},
				{
					Type:  TokenNumberLit,
					Bytes: []byte(`1`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 3, Line: 1, Column: 4},
						End:   hcl.Pos{Byte: 4, Line: 1, Column: 5},
					},
				},
				{
					Type:  TokenTemplateSeqEnd,
					Bytes: []byte(`}`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 4, Line: 1, Column: 5},
						End:   hcl.Pos{Byte: 5, Line: 1, Column: 6},
					},
				},
				{
					Type:  TokenCQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 5, Line: 1, Column: 6},
						End:   hcl.Pos{Byte: 6, Line: 1, Column: 7},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 6, Line: 1, Column: 7},
						End:   hcl.Pos{Byte: 6, Line: 1, Column: 7},
					},
				},
			},
		},
		{
			`"%{a}"`,
			[]Token{
				{
					Type:  TokenOQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenTemplateControl,
					Bytes: []byte(`%{`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 3, Line: 1, Column: 4},
					},
				},
				{
					Type:  TokenIdent,
					Bytes: []byte(`a`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 3, Line: 1, Column: 4},
						End:   hcl.Pos{Byte: 4, Line: 1, Column: 5},
					},
				},
				{
					Type:  TokenTemplateSeqEnd,
					Bytes: []byte(`}`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 4, Line: 1, Column: 5},
						End:   hcl.Pos{Byte: 5, Line: 1, Column: 6},
					},
				},
				{
					Type:  TokenCQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 5, Line: 1, Column: 6},
						End:   hcl.Pos{Byte: 6, Line: 1, Column: 7},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 6, Line: 1, Column: 7},
						End:   hcl.Pos{Byte: 6, Line: 1, Column: 7},
					},
				},
			},
		},
		{
			`"${{}}"`,
			[]Token{
				{
					Type:  TokenOQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenTemplateInterp,
					Bytes: []byte(`${`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 3, Line: 1, Column: 4},
					},
				},
				{
					Type:  TokenOBrace,
					Bytes: []byte(`{`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 3, Line: 1, Column: 4},
						End:   hcl.Pos{Byte: 4, Line: 1, Column: 5},
					},
				},
				{
					Type:  TokenCBrace,
					Bytes: []byte(`}`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 4, Line: 1, Column: 5},
						End:   hcl.Pos{Byte: 5, Line: 1, Column: 6},
					},
				},
				{
					Type:  TokenTemplateSeqEnd,
					Bytes: []byte(`}`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 5, Line: 1, Column: 6},
						End:   hcl.Pos{Byte: 6, Line: 1, Column: 7},
					},
				},
				{
					Type:  TokenCQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 6, Line: 1, Column: 7},
						End:   hcl.Pos{Byte: 7, Line: 1, Column: 8},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 7, Line: 1, Column: 8},
						End:   hcl.Pos{Byte: 7, Line: 1, Column: 8},
					},
				},
			},
		},
		{
			`"${""}"`,
			[]Token{
				{
					Type:  TokenOQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenTemplateInterp,
					Bytes: []byte(`${`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 3, Line: 1, Column: 4},
					},
				},
				{
					Type:  TokenOQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 3, Line: 1, Column: 4},
						End:   hcl.Pos{Byte: 4, Line: 1, Column: 5},
					},
				},
				{
					Type:  TokenCQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 4, Line: 1, Column: 5},
						End:   hcl.Pos{Byte: 5, Line: 1, Column: 6},
					},
				},
				{
					Type:  TokenTemplateSeqEnd,
					Bytes: []byte(`}`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 5, Line: 1, Column: 6},
						End:   hcl.Pos{Byte: 6, Line: 1, Column: 7},
					},
				},
				{
					Type:  TokenCQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 6, Line: 1, Column: 7},
						End:   hcl.Pos{Byte: 7, Line: 1, Column: 8},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 7, Line: 1, Column: 8},
						End:   hcl.Pos{Byte: 7, Line: 1, Column: 8},
					},
				},
			},
		},
		{
			`"${"${a}"}"`,
			[]Token{
				{
					Type:  TokenOQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenTemplateInterp,
					Bytes: []byte(`${`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 3, Line: 1, Column: 4},
					},
				},
				{
					Type:  TokenOQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 3, Line: 1, Column: 4},
						End:   hcl.Pos{Byte: 4, Line: 1, Column: 5},
					},
				},
				{
					Type:  TokenTemplateInterp,
					Bytes: []byte(`${`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 4, Line: 1, Column: 5},
						End:   hcl.Pos{Byte: 6, Line: 1, Column: 7},
					},
				},
				{
					Type:  TokenIdent,
					Bytes: []byte(`a`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 6, Line: 1, Column: 7},
						End:   hcl.Pos{Byte: 7, Line: 1, Column: 8},
					},
				},
				{
					Type:  TokenTemplateSeqEnd,
					Bytes: []byte(`}`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 7, Line: 1, Column: 8},
						End:   hcl.Pos{Byte: 8, Line: 1, Column: 9},
					},
				},
				{
					Type:  TokenCQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 8, Line: 1, Column: 9},
						End:   hcl.Pos{Byte: 9, Line: 1, Column: 10},
					},
				},
				{
					Type:  TokenTemplateSeqEnd,
					Bytes: []byte(`}`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 9, Line: 1, Column: 10},
						End:   hcl.Pos{Byte: 10, Line: 1, Column: 11},
					},
				},
				{
					Type:  TokenCQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 10, Line: 1, Column: 11},
						End:   hcl.Pos{Byte: 11, Line: 1, Column: 12},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 11, Line: 1, Column: 12},
						End:   hcl.Pos{Byte: 11, Line: 1, Column: 12},
					},
				},
			},
		},
		{
			`"${"${a} foo"}"`,
			[]Token{
				{
					Type:  TokenOQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenTemplateInterp,
					Bytes: []byte(`${`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 3, Line: 1, Column: 4},
					},
				},
				{
					Type:  TokenOQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 3, Line: 1, Column: 4},
						End:   hcl.Pos{Byte: 4, Line: 1, Column: 5},
					},
				},
				{
					Type:  TokenTemplateInterp,
					Bytes: []byte(`${`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 4, Line: 1, Column: 5},
						End:   hcl.Pos{Byte: 6, Line: 1, Column: 7},
					},
				},
				{
					Type:  TokenIdent,
					Bytes: []byte(`a`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 6, Line: 1, Column: 7},
						End:   hcl.Pos{Byte: 7, Line: 1, Column: 8},
					},
				},
				{
					Type:  TokenTemplateSeqEnd,
					Bytes: []byte(`}`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 7, Line: 1, Column: 8},
						End:   hcl.Pos{Byte: 8, Line: 1, Column: 9},
					},
				},
				{
					Type:  TokenQuotedLit,
					Bytes: []byte(` foo`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 8, Line: 1, Column: 9},
						End:   hcl.Pos{Byte: 12, Line: 1, Column: 13},
					},
				},
				{
					Type:  TokenCQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 12, Line: 1, Column: 13},
						End:   hcl.Pos{Byte: 13, Line: 1, Column: 14},
					},
				},
				{
					Type:  TokenTemplateSeqEnd,
					Bytes: []byte(`}`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 13, Line: 1, Column: 14},
						End:   hcl.Pos{Byte: 14, Line: 1, Column: 15},
					},
				},
				{
					Type:  TokenCQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 14, Line: 1, Column: 15},
						End:   hcl.Pos{Byte: 15, Line: 1, Column: 16},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 15, Line: 1, Column: 16},
						End:   hcl.Pos{Byte: 15, Line: 1, Column: 16},
					},
				},
			},
		},

		// Heredoc Templates
		{
			`<<EOT
hello world
EOT
`,
			[]Token{
				{
					Type:  TokenOHeredoc,
					Bytes: []byte("<<EOT\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 6, Line: 2, Column: 1},
					},
				},
				{
					Type:  TokenStringLit,
					Bytes: []byte("hello world\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 6, Line: 2, Column: 1},
						End:   hcl.Pos{Byte: 18, Line: 3, Column: 1},
					},
				},
				{
					Type:  TokenCHeredoc,
					Bytes: []byte("EOT"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 18, Line: 3, Column: 1},
						End:   hcl.Pos{Byte: 21, Line: 3, Column: 4},
					},
				},
				{
					Type:  TokenNewline,
					Bytes: []byte("\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 21, Line: 3, Column: 4},
						End:   hcl.Pos{Byte: 22, Line: 4, Column: 1},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 22, Line: 4, Column: 1},
						End:   hcl.Pos{Byte: 22, Line: 4, Column: 1},
					},
				},
			},
		},
		{
			"<<EOT\r\nhello world\r\nEOT\r\n", // intentional windows-style line endings
			[]Token{
				{
					Type:  TokenOHeredoc,
					Bytes: []byte("<<EOT\r\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 7, Line: 2, Column: 1},
					},
				},
				{
					Type:  TokenStringLit,
					Bytes: []byte("hello world\r\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 7, Line: 2, Column: 1},
						End:   hcl.Pos{Byte: 20, Line: 3, Column: 1},
					},
				},
				{
					Type:  TokenCHeredoc,
					Bytes: []byte("EOT"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 20, Line: 3, Column: 1},
						End:   hcl.Pos{Byte: 23, Line: 3, Column: 4},
					},
				},
				{
					Type:  TokenNewline,
					Bytes: []byte("\r\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 23, Line: 3, Column: 4},
						End:   hcl.Pos{Byte: 25, Line: 4, Column: 1},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 25, Line: 4, Column: 1},
						End:   hcl.Pos{Byte: 25, Line: 4, Column: 1},
					},
				},
			},
		},
		{
			`<<EOT
hello ${name}
EOT
`,
			[]Token{
				{
					Type:  TokenOHeredoc,
					Bytes: []byte("<<EOT\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 6, Line: 2, Column: 1},
					},
				},
				{
					Type:  TokenStringLit,
					Bytes: []byte("hello "),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 6, Line: 2, Column: 1},
						End:   hcl.Pos{Byte: 12, Line: 2, Column: 7},
					},
				},
				{
					Type:  TokenTemplateInterp,
					Bytes: []byte("${"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 12, Line: 2, Column: 7},
						End:   hcl.Pos{Byte: 14, Line: 2, Column: 9},
					},
				},
				{
					Type:  TokenIdent,
					Bytes: []byte("name"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 14, Line: 2, Column: 9},
						End:   hcl.Pos{Byte: 18, Line: 2, Column: 13},
					},
				},
				{
					Type:  TokenTemplateSeqEnd,
					Bytes: []byte("}"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 18, Line: 2, Column: 13},
						End:   hcl.Pos{Byte: 19, Line: 2, Column: 14},
					},
				},
				{
					Type:  TokenStringLit,
					Bytes: []byte("\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 19, Line: 2, Column: 14},
						End:   hcl.Pos{Byte: 20, Line: 3, Column: 1},
					},
				},
				{
					Type:  TokenCHeredoc,
					Bytes: []byte("EOT"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 20, Line: 3, Column: 1},
						End:   hcl.Pos{Byte: 23, Line: 3, Column: 4},
					},
				},
				{
					Type:  TokenNewline,
					Bytes: []byte("\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 23, Line: 3, Column: 4},
						End:   hcl.Pos{Byte: 24, Line: 4, Column: 1},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 24, Line: 4, Column: 1},
						End:   hcl.Pos{Byte: 24, Line: 4, Column: 1},
					},
				},
			},
		},
		{
			`<<EOT
${name}EOT
EOT
`,
			[]Token{
				{
					Type:  TokenOHeredoc,
					Bytes: []byte("<<EOT\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 6, Line: 2, Column: 1},
					},
				},
				{
					Type:  TokenTemplateInterp,
					Bytes: []byte("${"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 6, Line: 2, Column: 1},
						End:   hcl.Pos{Byte: 8, Line: 2, Column: 3},
					},
				},
				{
					Type:  TokenIdent,
					Bytes: []byte("name"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 8, Line: 2, Column: 3},
						End:   hcl.Pos{Byte: 12, Line: 2, Column: 7},
					},
				},
				{
					Type:  TokenTemplateSeqEnd,
					Bytes: []byte("}"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 12, Line: 2, Column: 7},
						End:   hcl.Pos{Byte: 13, Line: 2, Column: 8},
					},
				},
				{
					Type:  TokenStringLit,
					Bytes: []byte("EOT\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 13, Line: 2, Column: 8},
						End:   hcl.Pos{Byte: 17, Line: 3, Column: 1},
					},
				},
				{
					Type:  TokenCHeredoc,
					Bytes: []byte("EOT"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 17, Line: 3, Column: 1},
						End:   hcl.Pos{Byte: 20, Line: 3, Column: 4},
					},
				},
				{
					Type:  TokenNewline,
					Bytes: []byte("\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 20, Line: 3, Column: 4},
						End:   hcl.Pos{Byte: 21, Line: 4, Column: 1},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 21, Line: 4, Column: 1},
						End:   hcl.Pos{Byte: 21, Line: 4, Column: 1},
					},
				},
			},
		},
		{
			`<<EOF
${<<-EOF
hello
EOF
}
EOF
`,
			[]Token{
				{
					Type:  TokenOHeredoc,
					Bytes: []byte("<<EOF\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 6, Line: 2, Column: 1},
					},
				},
				{
					Type:  TokenTemplateInterp,
					Bytes: []byte("${"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 6, Line: 2, Column: 1},
						End:   hcl.Pos{Byte: 8, Line: 2, Column: 3},
					},
				},
				{
					Type:  TokenOHeredoc,
					Bytes: []byte("<<-EOF\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 8, Line: 2, Column: 3},
						End:   hcl.Pos{Byte: 15, Line: 3, Column: 1},
					},
				},
				{
					Type:  TokenStringLit,
					Bytes: []byte("hello\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 15, Line: 3, Column: 1},
						End:   hcl.Pos{Byte: 21, Line: 4, Column: 1},
					},
				},
				{
					Type:  TokenCHeredoc,
					Bytes: []byte("EOF"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 21, Line: 4, Column: 1},
						End:   hcl.Pos{Byte: 24, Line: 4, Column: 4},
					},
				},
				{
					Type:  TokenNewline,
					Bytes: []byte("\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 24, Line: 4, Column: 4},
						End:   hcl.Pos{Byte: 25, Line: 5, Column: 1},
					},
				},
				{
					Type:  TokenTemplateSeqEnd,
					Bytes: []byte("}"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 25, Line: 5, Column: 1},
						End:   hcl.Pos{Byte: 26, Line: 5, Column: 2},
					},
				},
				{
					Type:  TokenStringLit,
					Bytes: []byte("\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 26, Line: 5, Column: 2},
						End:   hcl.Pos{Byte: 27, Line: 6, Column: 1},
					},
				},
				{
					Type:  TokenCHeredoc,
					Bytes: []byte("EOF"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 27, Line: 6, Column: 1},
						End:   hcl.Pos{Byte: 30, Line: 6, Column: 4},
					},
				},
				{
					Type:  TokenNewline,
					Bytes: []byte("\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 30, Line: 6, Column: 4},
						End:   hcl.Pos{Byte: 31, Line: 7, Column: 1},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 31, Line: 7, Column: 1},
						End:   hcl.Pos{Byte: 31, Line: 7, Column: 1},
					},
				},
			},
		},

		// Combinations
		{
			` (1 + 2) * 3 `,
			[]Token{
				{
					Type:  TokenOParen,
					Bytes: []byte(`(`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 2, Line: 1, Column: 3},
					},
				},
				{
					Type:  TokenNumberLit,
					Bytes: []byte(`1`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 2, Line: 1, Column: 3},
						End:   hcl.Pos{Byte: 3, Line: 1, Column: 4},
					},
				},
				{
					Type:  TokenPlus,
					Bytes: []byte(`+`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 4, Line: 1, Column: 5},
						End:   hcl.Pos{Byte: 5, Line: 1, Column: 6},
					},
				},
				{
					Type:  TokenNumberLit,
					Bytes: []byte(`2`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 6, Line: 1, Column: 7},
						End:   hcl.Pos{Byte: 7, Line: 1, Column: 8},
					},
				},
				{
					Type:  TokenCParen,
					Bytes: []byte(`)`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 7, Line: 1, Column: 8},
						End:   hcl.Pos{Byte: 8, Line: 1, Column: 9},
					},
				},
				{
					Type:  TokenStar,
					Bytes: []byte(`*`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 9, Line: 1, Column: 10},
						End:   hcl.Pos{Byte: 10, Line: 1, Column: 11},
					},
				},
				{
					Type:  TokenNumberLit,
					Bytes: []byte(`3`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 11, Line: 1, Column: 12},
						End:   hcl.Pos{Byte: 12, Line: 1, Column: 13},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 13, Line: 1, Column: 14},
						End:   hcl.Pos{Byte: 13, Line: 1, Column: 14},
					},
				},
			},
		},
		{
			`9%8`,
			[]Token{
				{
					Type:  TokenNumberLit,
					Bytes: []byte(`9`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenPercent,
					Bytes: []byte(`%`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 2, Line: 1, Column: 3},
					},
				},
				{
					Type:  TokenNumberLit,
					Bytes: []byte(`8`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 2, Line: 1, Column: 3},
						End:   hcl.Pos{Byte: 3, Line: 1, Column: 4},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte(``),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 3, Line: 1, Column: 4},
						End:   hcl.Pos{Byte: 3, Line: 1, Column: 4},
					},
				},
			},
		},
		{
			"\na = 1\n",
			[]Token{
				{
					Type:  TokenNewline,
					Bytes: []byte("\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 2, Column: 1},
					},
				},
				{
					Type:  TokenIdent,
					Bytes: []byte("a"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 2, Column: 1},
						End:   hcl.Pos{Byte: 2, Line: 2, Column: 2},
					},
				},
				{
					Type:  TokenEqual,
					Bytes: []byte("="),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 3, Line: 2, Column: 3},
						End:   hcl.Pos{Byte: 4, Line: 2, Column: 4},
					},
				},
				{
					Type:  TokenNumberLit,
					Bytes: []byte("1"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 5, Line: 2, Column: 5},
						End:   hcl.Pos{Byte: 6, Line: 2, Column: 6},
					},
				},
				{
					Type:  TokenNewline,
					Bytes: []byte("\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 6, Line: 2, Column: 6},
						End:   hcl.Pos{Byte: 7, Line: 3, Column: 1},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 7, Line: 3, Column: 1},
						End:   hcl.Pos{Byte: 7, Line: 3, Column: 1},
					},
				},
			},
		},

		// Comments
		{
			"# hello\n",
			[]Token{
				{
					Type:  TokenComment,
					Bytes: []byte("# hello\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 8, Line: 2, Column: 1},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 8, Line: 2, Column: 1},
						End:   hcl.Pos{Byte: 8, Line: 2, Column: 1},
					},
				},
			},
		},
		{
			"// hello\n",
			[]Token{
				{
					Type:  TokenComment,
					Bytes: []byte("// hello\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 9, Line: 2, Column: 1},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 9, Line: 2, Column: 1},
						End:   hcl.Pos{Byte: 9, Line: 2, Column: 1},
					},
				},
			},
		},
		{
			"/* hello */",
			[]Token{
				{
					Type:  TokenComment,
					Bytes: []byte("/* hello */"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 11, Line: 1, Column: 12},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 11, Line: 1, Column: 12},
						End:   hcl.Pos{Byte: 11, Line: 1, Column: 12},
					},
				},
			},
		},
		{
			"/* hello */ howdy /* hey */",
			[]Token{
				{
					Type:  TokenComment,
					Bytes: []byte("/* hello */"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 11, Line: 1, Column: 12},
					},
				},
				{
					Type:  TokenIdent,
					Bytes: []byte("howdy"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 12, Line: 1, Column: 13},
						End:   hcl.Pos{Byte: 17, Line: 1, Column: 18},
					},
				},
				{
					Type:  TokenComment,
					Bytes: []byte("/* hey */"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 18, Line: 1, Column: 19},
						End:   hcl.Pos{Byte: 27, Line: 1, Column: 28},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 27, Line: 1, Column: 28},
						End:   hcl.Pos{Byte: 27, Line: 1, Column: 28},
					},
				},
			},
		},

		// Invalid things
		{
			`🌻`,
			[]Token{
				{
					Type:  TokenInvalid,
					Bytes: []byte(`🌻`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 4, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 4, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 4, Line: 1, Column: 2},
					},
				},
			},
		},
		{
			`|`,
			[]Token{
				{
					Type:  TokenBitwiseOr,
					Bytes: []byte(`|`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
			},
		},
		{
			"\x80", // UTF-8 continuation without an introducer
			[]Token{
				{
					Type:  TokenBadUTF8,
					Bytes: []byte{0x80},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
			},
		},
		{
			" \x80\x80", // UTF-8 continuation without an introducer
			[]Token{
				{
					Type:  TokenBadUTF8,
					Bytes: []byte{0x80},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 2, Line: 1, Column: 3},
					},
				},
				{
					Type:  TokenBadUTF8,
					Bytes: []byte{0x80},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 2, Line: 1, Column: 3},
						End:   hcl.Pos{Byte: 3, Line: 1, Column: 4},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 3, Line: 1, Column: 4},
						End:   hcl.Pos{Byte: 3, Line: 1, Column: 4},
					},
				},
			},
		},
		{
			"\t\t",
			[]Token{
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 2, Line: 1, Column: 3},
						End:   hcl.Pos{Byte: 2, Line: 1, Column: 3},
					},
				},
			},
		},

		// Misc combinations that have come up in bug reports, etc.
		{
			"locals {\n  is_percent = percent_sign == \"%\" ? true : false\n}\n",
			[]Token{
				{
					Type:  TokenIdent,
					Bytes: []byte(`locals`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 6, Line: 1, Column: 7},
					},
				},
				{
					Type:  TokenOBrace,
					Bytes: []byte{'{'},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 7, Line: 1, Column: 8},
						End:   hcl.Pos{Byte: 8, Line: 1, Column: 9},
					},
				},
				{
					Type:  TokenNewline,
					Bytes: []byte{'\n'},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 8, Line: 1, Column: 9},
						End:   hcl.Pos{Byte: 9, Line: 2, Column: 1},
					},
				},
				{
					Type:  TokenIdent,
					Bytes: []byte(`is_percent`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 11, Line: 2, Column: 3},
						End:   hcl.Pos{Byte: 21, Line: 2, Column: 13},
					},
				},
				{
					Type:  TokenEqual,
					Bytes: []byte(`=`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 22, Line: 2, Column: 14},
						End:   hcl.Pos{Byte: 23, Line: 2, Column: 15},
					},
				},
				{
					Type:  TokenIdent,
					Bytes: []byte(`percent_sign`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 24, Line: 2, Column: 16},
						End:   hcl.Pos{Byte: 36, Line: 2, Column: 28},
					},
				},
				{
					Type:  TokenEqualOp,
					Bytes: []byte(`==`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 37, Line: 2, Column: 29},
						End:   hcl.Pos{Byte: 39, Line: 2, Column: 31},
					},
				},
				{
					Type:  TokenOQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 40, Line: 2, Column: 32},
						End:   hcl.Pos{Byte: 41, Line: 2, Column: 33},
					},
				},
				{
					Type:  TokenQuotedLit,
					Bytes: []byte(`%`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 41, Line: 2, Column: 33},
						End:   hcl.Pos{Byte: 42, Line: 2, Column: 34},
					},
				},
				{
					Type:  TokenCQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 42, Line: 2, Column: 34},
						End:   hcl.Pos{Byte: 43, Line: 2, Column: 35},
					},
				},
				{
					Type:  TokenQuestion,
					Bytes: []byte(`?`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 44, Line: 2, Column: 36},
						End:   hcl.Pos{Byte: 45, Line: 2, Column: 37},
					},
				},
				{
					Type:  TokenIdent,
					Bytes: []byte(`true`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 46, Line: 2, Column: 38},
						End:   hcl.Pos{Byte: 50, Line: 2, Column: 42},
					},
				},
				{
					Type:  TokenColon,
					Bytes: []byte(`:`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 51, Line: 2, Column: 43},
						End:   hcl.Pos{Byte: 52, Line: 2, Column: 44},
					},
				},
				{
					Type:  TokenIdent,
					Bytes: []byte(`false`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 53, Line: 2, Column: 45},
						End:   hcl.Pos{Byte: 58, Line: 2, Column: 50},
					},
				},
				{
					Type:  TokenNewline,
					Bytes: []byte{'\n'},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 58, Line: 2, Column: 50},
						End:   hcl.Pos{Byte: 59, Line: 3, Column: 1},
					},
				},
				{
					Type:  TokenCBrace,
					Bytes: []byte{'}'},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 59, Line: 3, Column: 1},
						End:   hcl.Pos{Byte: 60, Line: 3, Column: 2},
					},
				},
				{
					Type:  TokenNewline,
					Bytes: []byte{'\n'},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 60, Line: 3, Column: 2},
						End:   hcl.Pos{Byte: 61, Line: 4, Column: 1},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 61, Line: 4, Column: 1},
						End:   hcl.Pos{Byte: 61, Line: 4, Column: 1},
					},
				},
			},
		},
	}

	prettyConfig := &pretty.Config{
		Diffable:          true,
		IncludeUnexported: true,
		PrintStringers:    true,
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			got := scanTokens([]byte(test.input), "", hcl.Pos{Byte: 0, Line: 1, Column: 1}, scanNormal)

			if !reflect.DeepEqual(got, test.want) {
				diff := prettyConfig.Compare(test.want, got)
				t.Errorf(
					"wrong result\ninput: %s\ndiff:  %s",
					test.input, diff,
				)
			}

			// "pretty" diff output is not helpful for all differences, so
			// we'll also print out a list of specific differences.
			for _, problem := range deep.Equal(got, test.want) {
				t.Error(problem)
			}

		})
	}
}

func TestScanTokens_template(t *testing.T) {
	tests := []struct {
		input string
		want  []Token
	}{
		// Empty input
		{
			``,
			[]Token{
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 0, Line: 1, Column: 1},
					},
				},
			},
		},

		// Simple literals
		{
			` hello `,
			[]Token{
				{
					Type:  TokenStringLit,
					Bytes: []byte(` hello `),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 7, Line: 1, Column: 8},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 7, Line: 1, Column: 8},
						End:   hcl.Pos{Byte: 7, Line: 1, Column: 8},
					},
				},
			},
		},
		{
			"\nhello\n",
			[]Token{
				{
					Type:  TokenStringLit,
					Bytes: []byte("\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 2, Column: 1},
					},
				},
				{
					Type:  TokenStringLit,
					Bytes: []byte("hello\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 2, Column: 1},
						End:   hcl.Pos{Byte: 7, Line: 3, Column: 1},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 7, Line: 3, Column: 1},
						End:   hcl.Pos{Byte: 7, Line: 3, Column: 1},
					},
				},
			},
		},
		{
			"hello ${foo} hello",
			[]Token{
				{
					Type:  TokenStringLit,
					Bytes: []byte("hello "),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 6, Line: 1, Column: 7},
					},
				},
				{
					Type:  TokenTemplateInterp,
					Bytes: []byte("${"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 6, Line: 1, Column: 7},
						End:   hcl.Pos{Byte: 8, Line: 1, Column: 9},
					},
				},
				{
					Type:  TokenIdent,
					Bytes: []byte("foo"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 8, Line: 1, Column: 9},
						End:   hcl.Pos{Byte: 11, Line: 1, Column: 12},
					},
				},
				{
					Type:  TokenTemplateSeqEnd,
					Bytes: []byte("}"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 11, Line: 1, Column: 12},
						End:   hcl.Pos{Byte: 12, Line: 1, Column: 13},
					},
				},
				{
					Type:  TokenStringLit,
					Bytes: []byte(" hello"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 12, Line: 1, Column: 13},
						End:   hcl.Pos{Byte: 18, Line: 1, Column: 19},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 18, Line: 1, Column: 19},
						End:   hcl.Pos{Byte: 18, Line: 1, Column: 19},
					},
				},
			},
		},
		{
			"hello ${~foo~} hello",
			[]Token{
				{
					Type:  TokenStringLit,
					Bytes: []byte("hello "),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 6, Line: 1, Column: 7},
					},
				},
				{
					Type:  TokenTemplateInterp,
					Bytes: []byte("${~"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 6, Line: 1, Column: 7},
						End:   hcl.Pos{Byte: 9, Line: 1, Column: 10},
					},
				},
				{
					Type:  TokenIdent,
					Bytes: []byte("foo"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 9, Line: 1, Column: 10},
						End:   hcl.Pos{Byte: 12, Line: 1, Column: 13},
					},
				},
				{
					Type:  TokenTemplateSeqEnd,
					Bytes: []byte("~}"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 12, Line: 1, Column: 13},
						End:   hcl.Pos{Byte: 14, Line: 1, Column: 15},
					},
				},
				{
					Type:  TokenStringLit,
					Bytes: []byte(" hello"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 14, Line: 1, Column: 15},
						End:   hcl.Pos{Byte: 20, Line: 1, Column: 21},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 20, Line: 1, Column: 21},
						End:   hcl.Pos{Byte: 20, Line: 1, Column: 21},
					},
				},
			},
		},
	}

	prettyConfig := &pretty.Config{
		Diffable:          true,
		IncludeUnexported: true,
		PrintStringers:    true,
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			got := scanTokens([]byte(test.input), "", hcl.Pos{Byte: 0, Line: 1, Column: 1}, scanTemplate)

			if !reflect.DeepEqual(got, test.want) {
				diff := prettyConfig.Compare(test.want, got)
				t.Errorf(
					"wrong result\ninput: %s\ndiff:  %s",
					test.input, diff,
				)
			}
		})
	}
}
