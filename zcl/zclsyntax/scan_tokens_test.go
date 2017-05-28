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
