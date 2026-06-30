// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package hclwrite

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

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
func TestObjectConsExprSetItemValue(t *testing.T) {
	tests := []struct {
		src      string
		attrName string
		key      string
		ctyValue cty.Value
		want     Tokens
	}{
		{
			`a = {
				hat = "derby", (cat) = "calico" }` + "\n",
			"a",
			"hat",
			cty.StringVal("bowler"),
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
		t.Run(fmt.Sprintf("%s = %s in %s", test.attrName, test.ctyValue.AsString(), test.src), func(t *testing.T) {
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
			expr.SetItemValue(test.key, test.ctyValue)

			got := f.BuildTokens(nil)
			format(got)
			if !reflect.DeepEqual(got, test.want) {
				diff := cmp.Diff(test.want, got)
				t.Errorf("wrong result\ngot:  %s\nwant: %s\ndiff:\n%s", spew.Sdump(got), spew.Sdump(test.want), diff)
			}
		})
	}
}

func TestObjectConsExprSetItemTraversal(t *testing.T) {
	tests := []struct {
		src       string
		attrName  string
		key       string
		traversal hcl.Traversal
		want      Tokens
	}{
		{
			`a = {
				hat = "derby", (cat) = "calico" }` + "\n",
			"a",
			"hat",
			hcl.Traversal{
				hcl.TraverseRoot{Name: `var.fancy_hat`},
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
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte("var.fancy_hat"),
					SpacesBefore: 1,
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
		t.Run(fmt.Sprintf("%s = %s in %s", test.attrName, string(TokensForTraversal(test.traversal).Bytes()), test.src), func(t *testing.T) {
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
			expr.SetItemTraversal(test.key, test.traversal)

			got := f.BuildTokens(nil)
			format(got)
			if !reflect.DeepEqual(got, test.want) {
				diff := cmp.Diff(test.want, got)
				t.Errorf("wrong result\ngot:  %s\nwant: %s\ndiff:\n%s", spew.Sdump(got), spew.Sdump(test.want), diff)
			}
		})
	}
}
