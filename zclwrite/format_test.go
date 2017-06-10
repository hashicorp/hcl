package zclwrite

import (
	"fmt"
	"testing"

	"reflect"

	"github.com/zclconf/go-zcl/zcl/zclsyntax"
)

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
