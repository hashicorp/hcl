package zclsyntax

import (
	"reflect"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/zclconf/go-zcl/zcl"
)

func TestScanTokens(t *testing.T) {
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
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   zcl.Pos{Byte: 0, Line: 1, Column: 1},
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
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   zcl.Pos{Byte: 1, Line: 1, Column: 2},
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
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   zcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   zcl.Pos{Byte: 1, Line: 1, Column: 2},
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
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   zcl.Pos{Byte: 2, Line: 1, Column: 3},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 2, Line: 1, Column: 3},
						End:   zcl.Pos{Byte: 2, Line: 1, Column: 3},
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
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   zcl.Pos{Byte: 4, Line: 1, Column: 5},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 4, Line: 1, Column: 5},
						End:   zcl.Pos{Byte: 4, Line: 1, Column: 5},
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
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   zcl.Pos{Byte: 3, Line: 1, Column: 4},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 3, Line: 1, Column: 4},
						End:   zcl.Pos{Byte: 3, Line: 1, Column: 4},
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
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   zcl.Pos{Byte: 4, Line: 1, Column: 5},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 4, Line: 1, Column: 5},
						End:   zcl.Pos{Byte: 4, Line: 1, Column: 5},
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
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   zcl.Pos{Byte: 5, Line: 1, Column: 6},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 5, Line: 1, Column: 6},
						End:   zcl.Pos{Byte: 5, Line: 1, Column: 6},
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
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   zcl.Pos{Byte: 5, Line: 1, Column: 6},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 5, Line: 1, Column: 6},
						End:   zcl.Pos{Byte: 5, Line: 1, Column: 6},
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
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   zcl.Pos{Byte: 7, Line: 1, Column: 6},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 7, Line: 1, Column: 6},
						End:   zcl.Pos{Byte: 7, Line: 1, Column: 6},
					},
				},
			},
		},

		// Invalid things
		{
			`|`,
			[]Token{
				{
					Type:  TokenInvalid,
					Bytes: []byte(`|`),
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   zcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   zcl.Pos{Byte: 1, Line: 1, Column: 2},
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
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   zcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   zcl.Pos{Byte: 1, Line: 1, Column: 2},
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
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   zcl.Pos{Byte: 2, Line: 1, Column: 3},
					},
				},
				{
					Type:  TokenBadUTF8,
					Bytes: []byte{0x80},
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 2, Line: 1, Column: 3},
						End:   zcl.Pos{Byte: 3, Line: 1, Column: 4},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 3, Line: 1, Column: 4},
						End:   zcl.Pos{Byte: 3, Line: 1, Column: 4},
					},
				},
			},
		},
		{
			"\t\t",
			[]Token{
				{
					Type:  TokenTabs,
					Bytes: []byte{0x09, 0x09},
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   zcl.Pos{Byte: 2, Line: 1, Column: 3},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 2, Line: 1, Column: 3},
						End:   zcl.Pos{Byte: 2, Line: 1, Column: 3},
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
			got := scanTokens([]byte(test.input), "", zcl.Pos{Byte: 0, Line: 1, Column: 1})

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
