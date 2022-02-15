package hclwrite

import (
	"bytes"
	"math/big"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func TestTokensForValue(t *testing.T) {
	tests := []struct {
		Val  cty.Value
		Want Tokens
	}{
		{
			cty.NullVal(cty.DynamicPseudoType),
			Tokens{
				{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte(`null`),
				},
			},
		},
		{
			cty.True,
			Tokens{
				{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte(`true`),
				},
			},
		},
		{
			cty.False,
			Tokens{
				{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte(`false`),
				},
			},
		},
		{
			cty.NumberIntVal(0),
			Tokens{
				{
					Type:  hclsyntax.TokenNumberLit,
					Bytes: []byte(`0`),
				},
			},
		},
		{
			cty.NumberFloatVal(0.5),
			Tokens{
				{
					Type:  hclsyntax.TokenNumberLit,
					Bytes: []byte(`0.5`),
				},
			},
		},
		{
			cty.NumberVal(big.NewFloat(0).SetPrec(512).Mul(big.NewFloat(40000000), big.NewFloat(2000000))),
			Tokens{
				{
					Type:  hclsyntax.TokenNumberLit,
					Bytes: []byte(`80000000000000`),
				},
			},
		},
		{
			cty.StringVal(""),
			Tokens{
				{
					Type:  hclsyntax.TokenOQuote,
					Bytes: []byte(`"`),
				},
				{
					Type:  hclsyntax.TokenCQuote,
					Bytes: []byte(`"`),
				},
			},
		},
		{
			cty.StringVal("foo"),
			Tokens{
				{
					Type:  hclsyntax.TokenOQuote,
					Bytes: []byte(`"`),
				},
				{
					Type:  hclsyntax.TokenQuotedLit,
					Bytes: []byte(`foo`),
				},
				{
					Type:  hclsyntax.TokenCQuote,
					Bytes: []byte(`"`),
				},
			},
		},
		{
			cty.StringVal(`"foo"`),
			Tokens{
				{
					Type:  hclsyntax.TokenOQuote,
					Bytes: []byte(`"`),
				},
				{
					Type:  hclsyntax.TokenQuotedLit,
					Bytes: []byte(`\"foo\"`),
				},
				{
					Type:  hclsyntax.TokenCQuote,
					Bytes: []byte(`"`),
				},
			},
		},
		{
			cty.StringVal("hello\nworld\n"),
			Tokens{
				{
					Type:  hclsyntax.TokenOQuote,
					Bytes: []byte(`"`),
				},
				{
					Type:  hclsyntax.TokenQuotedLit,
					Bytes: []byte(`hello\nworld\n`),
				},
				{
					Type:  hclsyntax.TokenCQuote,
					Bytes: []byte(`"`),
				},
			},
		},
		{
			cty.StringVal("hello\r\nworld\r\n"),
			Tokens{
				{
					Type:  hclsyntax.TokenOQuote,
					Bytes: []byte(`"`),
				},
				{
					Type:  hclsyntax.TokenQuotedLit,
					Bytes: []byte(`hello\r\nworld\r\n`),
				},
				{
					Type:  hclsyntax.TokenCQuote,
					Bytes: []byte(`"`),
				},
			},
		},
		{
			cty.StringVal(`what\what`),
			Tokens{
				{
					Type:  hclsyntax.TokenOQuote,
					Bytes: []byte(`"`),
				},
				{
					Type:  hclsyntax.TokenQuotedLit,
					Bytes: []byte(`what\\what`),
				},
				{
					Type:  hclsyntax.TokenCQuote,
					Bytes: []byte(`"`),
				},
			},
		},
		{
			cty.StringVal("𝄞"),
			Tokens{
				{
					Type:  hclsyntax.TokenOQuote,
					Bytes: []byte(`"`),
				},
				{
					Type:  hclsyntax.TokenQuotedLit,
					Bytes: []byte("𝄞"),
				},
				{
					Type:  hclsyntax.TokenCQuote,
					Bytes: []byte(`"`),
				},
			},
		},
		{
			cty.StringVal("👩🏾"),
			Tokens{
				{
					Type:  hclsyntax.TokenOQuote,
					Bytes: []byte(`"`),
				},
				{
					Type:  hclsyntax.TokenQuotedLit,
					Bytes: []byte(`👩🏾`),
				},
				{
					Type:  hclsyntax.TokenCQuote,
					Bytes: []byte(`"`),
				},
			},
		},
		{
			cty.EmptyTupleVal,
			Tokens{
				{
					Type:  hclsyntax.TokenOBrack,
					Bytes: []byte(`[`),
				},
				{
					Type:  hclsyntax.TokenCBrack,
					Bytes: []byte(`]`),
				},
			},
		},
		{
			cty.TupleVal([]cty.Value{cty.EmptyTupleVal}),
			Tokens{
				{
					Type:  hclsyntax.TokenOBrack,
					Bytes: []byte(`[`),
				},
				{
					Type:  hclsyntax.TokenOBrack,
					Bytes: []byte(`[`),
				},
				{
					Type:  hclsyntax.TokenCBrack,
					Bytes: []byte(`]`),
				},
				{
					Type:  hclsyntax.TokenCBrack,
					Bytes: []byte(`]`),
				},
			},
		},
		{
			cty.ListValEmpty(cty.String),
			Tokens{
				{
					Type:  hclsyntax.TokenOBrack,
					Bytes: []byte(`[`),
				},
				{
					Type:  hclsyntax.TokenCBrack,
					Bytes: []byte(`]`),
				},
			},
		},
		{
			cty.SetValEmpty(cty.Bool),
			Tokens{
				{
					Type:  hclsyntax.TokenOBrack,
					Bytes: []byte(`[`),
				},
				{
					Type:  hclsyntax.TokenCBrack,
					Bytes: []byte(`]`),
				},
			},
		},
		{
			cty.TupleVal([]cty.Value{cty.True}),
			Tokens{
				{
					Type:  hclsyntax.TokenOBrack,
					Bytes: []byte(`[`),
				},
				{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte(`true`),
				},
				{
					Type:  hclsyntax.TokenCBrack,
					Bytes: []byte(`]`),
				},
			},
		},
		{
			cty.TupleVal([]cty.Value{cty.True, cty.NumberIntVal(0)}),
			Tokens{
				{
					Type:  hclsyntax.TokenOBrack,
					Bytes: []byte(`[`),
				},
				{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte(`true`),
				},
				{
					Type:  hclsyntax.TokenComma,
					Bytes: []byte(`,`),
				},
				{
					Type:         hclsyntax.TokenNumberLit,
					Bytes:        []byte(`0`),
					SpacesBefore: 1,
				},
				{
					Type:  hclsyntax.TokenCBrack,
					Bytes: []byte(`]`),
				},
			},
		},
		{
			cty.EmptyObjectVal,
			Tokens{
				{
					Type:  hclsyntax.TokenOBrace,
					Bytes: []byte(`{`),
				},
				{
					Type:  hclsyntax.TokenCBrace,
					Bytes: []byte(`}`),
				},
			},
		},
		{
			cty.MapValEmpty(cty.Bool),
			Tokens{
				{
					Type:  hclsyntax.TokenOBrace,
					Bytes: []byte(`{`),
				},
				{
					Type:  hclsyntax.TokenCBrace,
					Bytes: []byte(`}`),
				},
			},
		},
		{
			cty.ObjectVal(map[string]cty.Value{
				"foo": cty.True,
			}),
			Tokens{
				{
					Type:  hclsyntax.TokenOBrace,
					Bytes: []byte(`{`),
				},
				{
					Type:  hclsyntax.TokenNewline,
					Bytes: []byte("\n"),
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte(`foo`),
					SpacesBefore: 2,
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte(`=`),
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte(`true`),
					SpacesBefore: 1,
				},
				{
					Type:  hclsyntax.TokenNewline,
					Bytes: []byte("\n"),
				},
				{
					Type:  hclsyntax.TokenCBrace,
					Bytes: []byte(`}`),
				},
			},
		},
		{
			cty.ObjectVal(map[string]cty.Value{
				"foo": cty.True,
				"bar": cty.NumberIntVal(0),
			}),
			Tokens{
				{
					Type:  hclsyntax.TokenOBrace,
					Bytes: []byte(`{`),
				},
				{
					Type:  hclsyntax.TokenNewline,
					Bytes: []byte("\n"),
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte(`bar`),
					SpacesBefore: 2,
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte(`=`),
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNumberLit,
					Bytes:        []byte(`0`),
					SpacesBefore: 1,
				},
				{
					Type:  hclsyntax.TokenNewline,
					Bytes: []byte("\n"),
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte(`foo`),
					SpacesBefore: 2,
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte(`=`),
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte(`true`),
					SpacesBefore: 1,
				},
				{
					Type:  hclsyntax.TokenNewline,
					Bytes: []byte("\n"),
				},
				{
					Type:  hclsyntax.TokenCBrace,
					Bytes: []byte(`}`),
				},
			},
		},
		{
			cty.ObjectVal(map[string]cty.Value{
				"foo bar": cty.True,
			}),
			Tokens{
				{
					Type:  hclsyntax.TokenOBrace,
					Bytes: []byte(`{`),
				},
				{
					Type:  hclsyntax.TokenNewline,
					Bytes: []byte("\n"),
				},
				{
					Type:         hclsyntax.TokenOQuote,
					Bytes:        []byte(`"`),
					SpacesBefore: 2,
				},
				{
					Type:  hclsyntax.TokenQuotedLit,
					Bytes: []byte(`foo bar`),
				},
				{
					Type:  hclsyntax.TokenCQuote,
					Bytes: []byte(`"`),
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte(`=`),
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte(`true`),
					SpacesBefore: 1,
				},
				{
					Type:  hclsyntax.TokenNewline,
					Bytes: []byte("\n"),
				},
				{
					Type:  hclsyntax.TokenCBrace,
					Bytes: []byte(`}`),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Val.GoString(), func(t *testing.T) {
			got := TokensForValue(test.Val)

			if !cmp.Equal(got, test.Want) {
				diff := cmp.Diff(got, test.Want, cmp.Comparer(func(a, b []byte) bool {
					return bytes.Equal(a, b)
				}))
				var gotBuf, wantBuf bytes.Buffer
				got.WriteTo(&gotBuf)
				test.Want.WriteTo(&wantBuf)
				t.Errorf(
					"wrong result\nvalue: %#v\ngot:   %s\nwant:  %s\ndiff:  %s",
					test.Val, gotBuf.String(), wantBuf.String(), diff,
				)
			}
		})
	}
}

func TestTokensForTraversal(t *testing.T) {
	tests := []struct {
		Val  hcl.Traversal
		Want Tokens
	}{
		{
			hcl.Traversal{
				hcl.TraverseRoot{Name: "root"},
				hcl.TraverseAttr{Name: "attr"},
				hcl.TraverseIndex{Key: cty.StringVal("index")},
			},
			Tokens{
				{Type: hclsyntax.TokenIdent, Bytes: []byte("root")},
				{Type: hclsyntax.TokenDot, Bytes: []byte(".")},
				{Type: hclsyntax.TokenIdent, Bytes: []byte("attr")},
				{Type: hclsyntax.TokenOBrack, Bytes: []byte{'['}},
				{Type: hclsyntax.TokenOQuote, Bytes: []byte(`"`)},
				{Type: hclsyntax.TokenQuotedLit, Bytes: []byte("index")},
				{Type: hclsyntax.TokenCQuote, Bytes: []byte(`"`)},
				{Type: hclsyntax.TokenCBrack, Bytes: []byte{']'}},
			},
		},
	}

	for _, test := range tests {
		got := TokensForTraversal(test.Val)

		if !cmp.Equal(got, test.Want) {
			diff := cmp.Diff(got, test.Want, cmp.Comparer(func(a, b []byte) bool {
				return bytes.Equal(a, b)
			}))
			var gotBuf, wantBuf bytes.Buffer
			got.WriteTo(&gotBuf)
			test.Want.WriteTo(&wantBuf)
			t.Errorf(
				"wrong result\nvalue: %#v\ngot:   %s\nwant:  %s\ndiff:  %s",
				test.Val, gotBuf.String(), wantBuf.String(), diff,
			)
		}
	}
}

func TestTokensForTuple(t *testing.T) {
	tests := map[string]struct {
		Val  []Tokens
		Want Tokens
	}{
		"no elements": {
			nil,
			Tokens{
				{Type: hclsyntax.TokenOBrack, Bytes: []byte{'['}},
				{Type: hclsyntax.TokenCBrack, Bytes: []byte{']'}},
			},
		},
		"one element": {
			[]Tokens{
				TokensForValue(cty.StringVal("foo")),
			},
			Tokens{
				{Type: hclsyntax.TokenOBrack, Bytes: []byte{'['}},
				{Type: hclsyntax.TokenOQuote, Bytes: []byte(`"`)},
				{Type: hclsyntax.TokenQuotedLit, Bytes: []byte("foo")},
				{Type: hclsyntax.TokenCQuote, Bytes: []byte(`"`)},
				{Type: hclsyntax.TokenCBrack, Bytes: []byte{']'}},
			},
		},
		"two elements": {
			[]Tokens{
				TokensForTraversal(hcl.Traversal{
					hcl.TraverseRoot{Name: "root"},
					hcl.TraverseAttr{Name: "attr"},
				}),
				TokensForValue(cty.StringVal("foo")),
			},
			Tokens{
				{Type: hclsyntax.TokenOBrack, Bytes: []byte{'['}},
				{Type: hclsyntax.TokenIdent, Bytes: []byte("root")},
				{Type: hclsyntax.TokenDot, Bytes: []byte(".")},
				{Type: hclsyntax.TokenIdent, Bytes: []byte("attr")},
				{Type: hclsyntax.TokenComma, Bytes: []byte{','}},
				{Type: hclsyntax.TokenOQuote, Bytes: []byte(`"`), SpacesBefore: 1},
				{Type: hclsyntax.TokenQuotedLit, Bytes: []byte("foo")},
				{Type: hclsyntax.TokenCQuote, Bytes: []byte(`"`)},
				{Type: hclsyntax.TokenCBrack, Bytes: []byte{']'}},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := TokensForTuple(test.Val)

			if !cmp.Equal(got, test.Want) {
				diff := cmp.Diff(got, test.Want, cmp.Comparer(func(a, b []byte) bool {
					return bytes.Equal(a, b)
				}))
				var gotBuf, wantBuf bytes.Buffer
				got.WriteTo(&gotBuf)
				test.Want.WriteTo(&wantBuf)
				t.Errorf(
					"wrong result\nvalue: %#v\ngot:   %s\nwant:  %s\ndiff:  %s",
					test.Val, gotBuf.String(), wantBuf.String(), diff,
				)
			}
		})
	}
}

func TestTokensForObject(t *testing.T) {
	tests := map[string]struct {
		Val  []ObjectAttrTokens
		Want Tokens
	}{
		"no attributes": {
			nil,
			Tokens{
				{Type: hclsyntax.TokenOBrace, Bytes: []byte{'{'}},
				{Type: hclsyntax.TokenCBrace, Bytes: []byte{'}'}},
			},
		},
		"one attribute": {
			[]ObjectAttrTokens{
				{
					Name: TokensForTraversal(hcl.Traversal{
						hcl.TraverseRoot{Name: "bar"},
					}),
					Value: TokensForValue(cty.StringVal("baz")),
				},
			},
			Tokens{
				{Type: hclsyntax.TokenOBrace, Bytes: []byte{'{'}},
				{Type: hclsyntax.TokenNewline, Bytes: []byte{'\n'}},
				{Type: hclsyntax.TokenIdent, Bytes: []byte("bar"), SpacesBefore: 2},
				{Type: hclsyntax.TokenEqual, Bytes: []byte{'='}, SpacesBefore: 1},
				{Type: hclsyntax.TokenOQuote, Bytes: []byte(`"`), SpacesBefore: 1},
				{Type: hclsyntax.TokenQuotedLit, Bytes: []byte("baz")},
				{Type: hclsyntax.TokenCQuote, Bytes: []byte(`"`)},
				{Type: hclsyntax.TokenNewline, Bytes: []byte{'\n'}},
				{Type: hclsyntax.TokenCBrace, Bytes: []byte{'}'}},
			},
		},
		"two attributes": {
			[]ObjectAttrTokens{
				{
					Name: TokensForTraversal(hcl.Traversal{
						hcl.TraverseRoot{Name: "foo"},
					}),
					Value: TokensForTraversal(hcl.Traversal{
						hcl.TraverseRoot{Name: "root"},
						hcl.TraverseAttr{Name: "attr"},
					}),
				},
				{
					Name: TokensForTraversal(hcl.Traversal{
						hcl.TraverseRoot{Name: "bar"},
					}),
					Value: TokensForValue(cty.StringVal("baz")),
				},
			},
			Tokens{
				{Type: hclsyntax.TokenOBrace, Bytes: []byte{'{'}},
				{Type: hclsyntax.TokenNewline, Bytes: []byte{'\n'}},
				{Type: hclsyntax.TokenIdent, Bytes: []byte("foo"), SpacesBefore: 2},
				{Type: hclsyntax.TokenEqual, Bytes: []byte{'='}, SpacesBefore: 1},
				{Type: hclsyntax.TokenIdent, Bytes: []byte("root"), SpacesBefore: 1},
				{Type: hclsyntax.TokenDot, Bytes: []byte(".")},
				{Type: hclsyntax.TokenIdent, Bytes: []byte("attr")},
				{Type: hclsyntax.TokenNewline, Bytes: []byte{'\n'}},
				{Type: hclsyntax.TokenIdent, Bytes: []byte("bar"), SpacesBefore: 2},
				{Type: hclsyntax.TokenEqual, Bytes: []byte{'='}, SpacesBefore: 1},
				{Type: hclsyntax.TokenOQuote, Bytes: []byte(`"`), SpacesBefore: 1},
				{Type: hclsyntax.TokenQuotedLit, Bytes: []byte("baz")},
				{Type: hclsyntax.TokenCQuote, Bytes: []byte(`"`)},
				{Type: hclsyntax.TokenNewline, Bytes: []byte{'\n'}},
				{Type: hclsyntax.TokenCBrace, Bytes: []byte{'}'}},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := TokensForObject(test.Val)

			if !cmp.Equal(got, test.Want) {
				diff := cmp.Diff(got, test.Want, cmp.Comparer(func(a, b []byte) bool {
					return bytes.Equal(a, b)
				}))
				var gotBuf, wantBuf bytes.Buffer
				got.WriteTo(&gotBuf)
				test.Want.WriteTo(&wantBuf)
				t.Errorf(
					"wrong result\nvalue: %#v\ngot:   %s\nwant:  %s\ndiff:  %s",
					test.Val, gotBuf.String(), wantBuf.String(), diff,
				)
			}
		})
	}
}

func TestTokensForFunctionCall(t *testing.T) {
	tests := map[string]struct {
		FuncName string
		Val      []Tokens
		Want     Tokens
	}{
		"no arguments": {
			"uuid",
			nil,
			Tokens{
				{Type: hclsyntax.TokenIdent, Bytes: []byte("uuid")},
				{Type: hclsyntax.TokenOParen, Bytes: []byte{'('}},
				{Type: hclsyntax.TokenCParen, Bytes: []byte(")")},
			},
		},
		"one argument": {
			"strlen",
			[]Tokens{
				TokensForValue(cty.StringVal("hello")),
			},
			Tokens{
				{Type: hclsyntax.TokenIdent, Bytes: []byte("strlen")},
				{Type: hclsyntax.TokenOParen, Bytes: []byte{'('}},
				{Type: hclsyntax.TokenOQuote, Bytes: []byte(`"`)},
				{Type: hclsyntax.TokenQuotedLit, Bytes: []byte("hello")},
				{Type: hclsyntax.TokenCQuote, Bytes: []byte(`"`)},
				{Type: hclsyntax.TokenCParen, Bytes: []byte(")")},
			},
		},
		"two arguments": {
			"list",
			[]Tokens{
				TokensForIdentifier("string"),
				TokensForIdentifier("int"),
			},
			Tokens{
				{Type: hclsyntax.TokenIdent, Bytes: []byte("list")},
				{Type: hclsyntax.TokenOParen, Bytes: []byte{'('}},
				{Type: hclsyntax.TokenIdent, Bytes: []byte("string")},
				{Type: hclsyntax.TokenComma, Bytes: []byte(",")},
				{Type: hclsyntax.TokenIdent, Bytes: []byte("int"), SpacesBefore: 1},
				{Type: hclsyntax.TokenCParen, Bytes: []byte(")")},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := TokensForFunctionCall(test.FuncName, test.Val...)

			if !cmp.Equal(got, test.Want) {
				diff := cmp.Diff(got, test.Want, cmp.Comparer(func(a, b []byte) bool {
					return bytes.Equal(a, b)
				}))
				var gotBuf, wantBuf bytes.Buffer
				got.WriteTo(&gotBuf)
				test.Want.WriteTo(&wantBuf)
				t.Errorf(
					"wrong result\nvalue: %#v\ngot:   %s\nwant:  %s\ndiff:  %s",
					test.Val, gotBuf.String(), wantBuf.String(), diff,
				)
			}
		})
	}
}

func TestTokenGenerateConsistency(t *testing.T) {

	bytesComparer := cmp.Comparer(func(a, b []byte) bool {
		return bytes.Equal(a, b)
	})

	// This test verifies that different ways of generating equivalent token
	// sequences all generate identical tokens, to help us keep them all in
	// sync under future maintanence.

	t.Run("tuple constructor", func(t *testing.T) {
		tests := map[string]struct {
			elems []cty.Value
		}{
			"no elements": {
				nil,
			},
			"one element": {
				[]cty.Value{
					cty.StringVal("hello"),
				},
			},
			"two elements": {
				[]cty.Value{
					cty.StringVal("hello"),
					cty.StringVal("world"),
				},
			},
		}

		for name, test := range tests {
			t.Run(name, func(t *testing.T) {
				var listVal cty.Value
				if len(test.elems) > 0 {
					listVal = cty.ListVal(test.elems)
				} else {
					listVal = cty.ListValEmpty(cty.DynamicPseudoType)
				}
				fromListValue := TokensForValue(listVal)
				fromTupleValue := TokensForValue(cty.TupleVal(test.elems))
				elemTokens := make([]Tokens, len(test.elems))
				for i, v := range test.elems {
					elemTokens[i] = TokensForValue(v)
				}
				fromTupleTokens := TokensForTuple(elemTokens)

				if diff := cmp.Diff(fromListValue, fromTupleTokens, bytesComparer); diff != "" {
					t.Errorf("inconsistency between TokensForValue(list) and TokensForTuple\n%s", diff)
				}
				if diff := cmp.Diff(fromTupleValue, fromTupleTokens, bytesComparer); diff != "" {
					t.Errorf("inconsistency between TokensForValue(tuple) and TokensForTuple\n%s", diff)
				}

			})
		}
	})

	t.Run("object constructor", func(t *testing.T) {
		tests := map[string]struct {
			attrs map[string]cty.Value
		}{
			"no elements": {
				nil,
			},
			"one element": {
				map[string]cty.Value{
					"greeting": cty.StringVal("hello"),
				},
			},
			"two elements": {
				map[string]cty.Value{
					"greeting1": cty.StringVal("hello"),
					"greeting2": cty.StringVal("world"),
				},
			},
		}

		for name, test := range tests {
			t.Run(name, func(t *testing.T) {
				var mapVal cty.Value
				if len(test.attrs) > 0 {
					mapVal = cty.MapVal(test.attrs)
				} else {
					mapVal = cty.MapValEmpty(cty.DynamicPseudoType)
				}
				fromMapValue := TokensForValue(mapVal)
				fromObjectValue := TokensForValue(cty.ObjectVal(test.attrs))
				attrTokens := make([]ObjectAttrTokens, 0, len(test.attrs))

				// TokensForValue always writes the keys/attributes in cty's
				// standard iteration order, but TokensForObject gives the
				// caller direct control of the ordering. The result is
				// therefore consistent only if the given attributes are
				// pre-sorted into the same iteration order, which is a lexical
				// sort by attribute name.
				keys := make([]string, 0, len(test.attrs))
				for k := range test.attrs {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					v := test.attrs[k]
					attrTokens = append(attrTokens, ObjectAttrTokens{
						Name:  TokensForIdentifier(k),
						Value: TokensForValue(v),
					})
				}
				fromObjectTokens := TokensForObject(attrTokens)

				if diff := cmp.Diff(fromMapValue, fromObjectTokens, bytesComparer); diff != "" {
					t.Errorf("inconsistency between TokensForValue(map) and TokensForObject\n%s", diff)
				}
				if diff := cmp.Diff(fromObjectValue, fromObjectTokens, bytesComparer); diff != "" {
					t.Errorf("inconsistency between TokensForValue(object) and TokensForObject\n%s", diff)
				}
			})
		}
	})
}
