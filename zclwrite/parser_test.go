package zclwrite

import (
	"reflect"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/zclconf/go-zcl/zcl/zclsyntax"
)

func TestLexConfig(t *testing.T) {
	tests := []struct {
		input string
		want  Tokens
	}{
		{
			`a  b `,
			Tokens{
				{
					Type:         zclsyntax.TokenIdent,
					Bytes:        []byte(`a`),
					SpacesBefore: 0,
				},
				{
					Type:         zclsyntax.TokenIdent,
					Bytes:        []byte(`b`),
					SpacesBefore: 2,
				},
				{
					Type:         zclsyntax.TokenEOF,
					Bytes:        []byte{},
					SpacesBefore: 1,
				},
			},
		},
		{
			`
foo "bar" "baz" {
    pizza = " cheese "
}
`,
			Tokens{
				{
					Type:         zclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
				{
					Type:         zclsyntax.TokenIdent,
					Bytes:        []byte(`foo`),
					SpacesBefore: 0,
				},
				{
					Type:         zclsyntax.TokenOQuote,
					Bytes:        []byte(`"`),
					SpacesBefore: 1,
				},
				{
					Type:         zclsyntax.TokenStringLit,
					Bytes:        []byte(`bar`),
					SpacesBefore: 0,
				},
				{
					Type:         zclsyntax.TokenCQuote,
					Bytes:        []byte(`"`),
					SpacesBefore: 0,
				},
				{
					Type:         zclsyntax.TokenOQuote,
					Bytes:        []byte(`"`),
					SpacesBefore: 1,
				},
				{
					Type:         zclsyntax.TokenStringLit,
					Bytes:        []byte(`baz`),
					SpacesBefore: 0,
				},
				{
					Type:         zclsyntax.TokenCQuote,
					Bytes:        []byte(`"`),
					SpacesBefore: 0,
				},
				{
					Type:         zclsyntax.TokenOBrace,
					Bytes:        []byte(`{`),
					SpacesBefore: 1,
				},
				{
					Type:         zclsyntax.TokenNewline,
					Bytes:        []byte("\n"),
					SpacesBefore: 0,
				},
				{
					Type:         zclsyntax.TokenIdent,
					Bytes:        []byte(`pizza`),
					SpacesBefore: 4,
				},
				{
					Type:         zclsyntax.TokenEqual,
					Bytes:        []byte(`=`),
					SpacesBefore: 1,
				},
				{
					Type:         zclsyntax.TokenOQuote,
					Bytes:        []byte(`"`),
					SpacesBefore: 1,
				},
				{
					Type:         zclsyntax.TokenStringLit,
					Bytes:        []byte(` cheese `),
					SpacesBefore: 0,
				},
				{
					Type:         zclsyntax.TokenCQuote,
					Bytes:        []byte(`"`),
					SpacesBefore: 0,
				},
				{
					Type:         zclsyntax.TokenNewline,
					Bytes:        []byte("\n"),
					SpacesBefore: 0,
				},
				{
					Type:         zclsyntax.TokenCBrace,
					Bytes:        []byte(`}`),
					SpacesBefore: 0,
				},
				{
					Type:         zclsyntax.TokenNewline,
					Bytes:        []byte("\n"),
					SpacesBefore: 0,
				},
				{
					Type:         zclsyntax.TokenEOF,
					Bytes:        []byte{},
					SpacesBefore: 0,
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
			got := lexConfig([]byte(test.input))

			if !reflect.DeepEqual(got, test.want) {
				diff := prettyConfig.Compare(test.want, got)
				t.Errorf(
					"wrong result\ninput: %s\ndiff:  %s", test.input, diff,
				)
			}
		})
	}
}
