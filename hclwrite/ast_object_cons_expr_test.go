// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package hclwrite

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func TestObjectConsExprValueFor(t *testing.T) {
	ident := newIdentifier(TokensForIdentifier("hat")[0])
	key := newObjectConsKeyExpr(newNode(ident))

	valueExpr := NewExpressionRaw(TokensForValue(cty.StringVal("fez")))
	value := newObjectConsValue()
	value.expr = value.children.Append(valueExpr)

	item := newObjectConsItem()
	item.key = item.children.Append(key)
	item.value = item.children.Append(value)

	object := newObjectConsExpr()
	object.items.Add(object.children.Append(item))

	var actual strings.Builder
	object.ValueFor("hat").expr.BuildTokens(nil).WriteTo(&actual)
	expected := `"fez"`
	if diff := cmp.Diff(actual.String(), expected); len(diff) > 0 {
		t.Error(diff)
	}
}

func TestObjectConsExprSetItemRaw(t *testing.T) {
	tests := []struct {
		src      string
		attrName string
		key      string
		tokens   Tokens
		want     Tokens
	}{
		{
			`a = {
				hat = "derby", (cat) = "calico" }` + "\n",
			"a",
			"hat",
			Tokens{
				{
					Type:         hclsyntax.TokenOQuote,
					Bytes:        []byte(`"`),
					SpacesBefore: 1,
				},
				{
					Type:  hclsyntax.TokenQuotedLit,
					Bytes: []byte(`bowler`),
				},
				{
					Type:  hclsyntax.TokenCQuote,
					Bytes: []byte(`"`),
				},
			},
			Tokens{
				{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte{'a'},
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte{'='},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenOBrace,
					Bytes:        []byte{'{'},
					SpacesBefore: 1,
				},
				{
					Type:  hclsyntax.TokenNewline,
					Bytes: []byte("\n"),
				},
				{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte("hat"),
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte{'='},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenOQuote,
					Bytes:        []byte{'"'},
					SpacesBefore: 1,
				},
				{
					Type:  hclsyntax.TokenQuotedLit,
					Bytes: []byte(`bowler`),
				},
				{
					Type:  hclsyntax.TokenCQuote,
					Bytes: []byte{'"'},
				},
				{
					Type:  hclsyntax.TokenComma,
					Bytes: []byte{','},
				},
				{
					Type:         hclsyntax.TokenOParen,
					Bytes:        []byte{'('},
					SpacesBefore: 1,
				},
				{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte("cat"),
				},
				{
					Type:  hclsyntax.TokenCParen,
					Bytes: []byte{')'},
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte{'='},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenOQuote,
					Bytes:        []byte{'"'},
					SpacesBefore: 1,
				},
				{
					Type:  hclsyntax.TokenQuotedLit,
					Bytes: []byte("calico"),
				},
				{
					Type:  hclsyntax.TokenCQuote,
					Bytes: []byte{'"'},
				},
				{
					Type:         hclsyntax.TokenCBrace,
					Bytes:        []byte{'}'},
					SpacesBefore: 1,
				},
				{
					Type:  hclsyntax.TokenNewline,
					Bytes: []byte("\n"),
				},
				{
					Type:         hclsyntax.TokenEOF,
					Bytes:        []byte{},
					SpacesBefore: 0,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s = %s in %s", test.attrName, test.tokens.Bytes(), test.src), func(t *testing.T) {
			f, diags := ParseConfig([]byte(test.src), "", hcl.Pos{Line: 1, Column: 1})
			if len(diags) != 0 {
				for _, diag := range diags {
					t.Logf("- %s", diag.Error())
				}
				t.Fatalf("unexpected diagnostics")
			}

			attr := f.Body().GetAttribute(test.attrName)
			if attr == nil {
				t.Fatal("attr nil")
			}
			expr := attr.Expr().AsObjectConsExpr()
			expr.SetItemRaw(test.key, test.tokens)

			got := f.BuildTokens(nil)
			format(got)
			if !reflect.DeepEqual(got, test.want) {
				diff := cmp.Diff(test.want, got)
				t.Errorf("wrong result\ngot:  %s\nwant: %s\ndiff:\n%s", spew.Sdump(got), spew.Sdump(test.want), diff)
			}
		})
	}
}
