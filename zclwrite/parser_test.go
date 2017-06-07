package zclwrite

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/kylelemons/godebug/pretty"
	"github.com/zclconf/go-zcl/zcl"
	"github.com/zclconf/go-zcl/zcl/zclsyntax"
)

func TestParse(t *testing.T) {
	tests := []struct {
		src  string
		want *Body
	}{
		{
			"",
			&Body{
				Items:     nil,
				AllTokens: nil,
			},
		},
		{
			"a = 1",
			&Body{
				Items: []Node{
					&Unstructured{
						AllTokens: &TokenSeq{Tokens{
							{
								Type:         zclsyntax.TokenIdent,
								Bytes:        []byte(`a`),
								SpacesBefore: 0,
							},
							{
								Type:         zclsyntax.TokenEqual,
								Bytes:        []byte(`=`),
								SpacesBefore: 1,
							},
							{
								Type:         zclsyntax.TokenNumberLit,
								Bytes:        []byte(`1`),
								SpacesBefore: 1,
							},
						}},
					},
				},
				AllTokens: &TokenSeq{
					&TokenSeq{
						Tokens{
							{
								Type:         zclsyntax.TokenIdent,
								Bytes:        []byte(`a`),
								SpacesBefore: 0,
							},
							{
								Type:         zclsyntax.TokenEqual,
								Bytes:        []byte(`=`),
								SpacesBefore: 1,
							},
							{
								Type:         zclsyntax.TokenNumberLit,
								Bytes:        []byte(`1`),
								SpacesBefore: 1,
							},
						},
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
		t.Run(test.src, func(t *testing.T) {
			file, diags := parse([]byte(test.src), "", zcl.Pos{Line: 1, Column: 1})
			if len(diags) > 0 {
				for _, diag := range diags {
					t.Logf(" - %s", diag.Error())
				}
				t.Fatalf("unexpected diagnostics")
			}

			got := file.Body

			if !reflect.DeepEqual(got, test.want) {
				diff := prettyConfig.Compare(got, test.want)
				if diff != "" {
					t.Errorf(
						"wrong result\ninput: %s\ndiff:  %s",
						test.src,
						diff,
					)
				} else {
					t.Errorf(
						"wrong result\ninput: %s\ngot:   %s\nwant:  %s",
						test.src,
						spew.Sdump(got),
						spew.Sdump(test.want),
					)
				}
			}
		})
	}
}

func TestPartitionTokens(t *testing.T) {
	tests := []struct {
		tokens    zclsyntax.Tokens
		rng       zcl.Range
		wantStart int
		wantEnd   int
	}{
		{
			zclsyntax.Tokens{},
			zcl.Range{
				Start: zcl.Pos{Byte: 0},
				End:   zcl.Pos{Byte: 0},
			},
			0,
			0,
		},
		{
			zclsyntax.Tokens{
				{
					Type: zclsyntax.TokenIdent,
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 0},
						End:   zcl.Pos{Byte: 4},
					},
				},
			},
			zcl.Range{
				Start: zcl.Pos{Byte: 0},
				End:   zcl.Pos{Byte: 4},
			},
			0,
			1,
		},
		{
			zclsyntax.Tokens{
				{
					Type: zclsyntax.TokenIdent,
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 0},
						End:   zcl.Pos{Byte: 4},
					},
				},
				{
					Type: zclsyntax.TokenIdent,
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 4},
						End:   zcl.Pos{Byte: 8},
					},
				},
				{
					Type: zclsyntax.TokenIdent,
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 8},
						End:   zcl.Pos{Byte: 12},
					},
				},
			},
			zcl.Range{
				Start: zcl.Pos{Byte: 4},
				End:   zcl.Pos{Byte: 8},
			},
			1,
			2,
		},
		{
			zclsyntax.Tokens{
				{
					Type: zclsyntax.TokenIdent,
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 0},
						End:   zcl.Pos{Byte: 4},
					},
				},
				{
					Type: zclsyntax.TokenIdent,
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 4},
						End:   zcl.Pos{Byte: 8},
					},
				},
				{
					Type: zclsyntax.TokenIdent,
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 8},
						End:   zcl.Pos{Byte: 12},
					},
				},
			},
			zcl.Range{
				Start: zcl.Pos{Byte: 0},
				End:   zcl.Pos{Byte: 8},
			},
			0,
			2,
		},
		{
			zclsyntax.Tokens{
				{
					Type: zclsyntax.TokenIdent,
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 0},
						End:   zcl.Pos{Byte: 4},
					},
				},
				{
					Type: zclsyntax.TokenIdent,
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 4},
						End:   zcl.Pos{Byte: 8},
					},
				},
				{
					Type: zclsyntax.TokenIdent,
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 8},
						End:   zcl.Pos{Byte: 12},
					},
				},
			},
			zcl.Range{
				Start: zcl.Pos{Byte: 4},
				End:   zcl.Pos{Byte: 12},
			},
			1,
			3,
		},
	}

	prettyConfig := &pretty.Config{
		Diffable:          true,
		IncludeUnexported: true,
		PrintStringers:    true,
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			gotStart, gotEnd := partitionTokens(test.tokens, test.rng)

			if gotStart != test.wantStart || gotEnd != test.wantEnd {
				t.Errorf(
					"wrong result\ntokens: %s\nrange: %#v\ngot:   %d, %d\nwant:  %d, %d",
					prettyConfig.Sprint(test.tokens), test.rng,
					gotStart, test.wantStart,
					gotEnd, test.wantEnd,
				)
			}
		})
	}
}

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
					Type:         zclsyntax.TokenQuotedLit,
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
					Type:         zclsyntax.TokenQuotedLit,
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
					Type:         zclsyntax.TokenQuotedLit,
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
