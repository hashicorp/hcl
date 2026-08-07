// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package hclwrite

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func TestTupleConsExprClear(t *testing.T) {
	tests := []struct {
		testName    string
		src         string
		countBefore int
		attrName    string
		want        Tokens
	}{
		{
			testName: `empty tuple`,
			src:      `a = [ ]`,
			attrName: `a`,
			want: Tokens{
				{
					Type:  hclsyntax.TokenOBrack,
					Bytes: []byte{'['},
				},
				{
					Type:  hclsyntax.TokenCBrack,
					Bytes: []byte{']'},
				},
			},
		},
		{
			testName: `non-empty tuple`,
			src:      `a = ["derby", "calico", null]`,
			attrName: `a`,
			want: Tokens{
				{
					Type:  hclsyntax.TokenOBrack,
					Bytes: []byte{'['},
				},
				{
					Type:  hclsyntax.TokenCBrack,
					Bytes: []byte{']'},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			f, diags := ParseConfig([]byte(test.src), "", hcl.InitialPos)
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

			expr := attr.Expr().AsTupleConsExpr()
			expr.Clear()

			got := expr.BuildTokens(nil)
			format(got)

			count := len(expr.Items())
			if count != 0 {
				t.Errorf("expected 0 items; got %d items", count)
			}
			if !reflect.DeepEqual(got, test.want) {
				diff := cmp.Diff(test.want, got)
				t.Errorf("wrong result\ngot:  %s\nwant: %s\ndiff:\n%s", spew.Sdump(got), spew.Sdump(test.want), diff)
			}
		})
	}
}
func TestTupleConsExprItems(t *testing.T) {
	tests := []struct {
		testName string
		src      string
		attrName string
		want     []Tokens // One Tokens value per tuple item.
	}{
		{
			testName: `simple tuple`,
			src:      `a = ["derby", "calico", null]`,
			attrName: `a`,
			want: []Tokens{
				{
					{
						Type:  hclsyntax.TokenOQuote,
						Bytes: []byte{'"'},
					},
					{
						Type:  hclsyntax.TokenQuotedLit,
						Bytes: []byte("derby"),
					},
					{
						Type:  hclsyntax.TokenCQuote,
						Bytes: []byte{'"'},
					},
				},
				{
					{
						Type:  hclsyntax.TokenOQuote,
						Bytes: []byte{'"'},
					},
					{
						Type:  hclsyntax.TokenQuotedLit,
						Bytes: []byte("calico"),
					},
					{
						Type:  hclsyntax.TokenCQuote,
						Bytes: []byte{'"'},
					},
				},
				{
					{
						Type:  hclsyntax.TokenIdent,
						Bytes: []byte("null"),
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			f, diags := ParseConfig([]byte(test.src), "", hcl.InitialPos)
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

			expr := attr.Expr().AsTupleConsExpr()
			items := expr.Items()

			got := make([]Tokens, len(items))
			for i := range items {
				got[i] = items[i].BuildTokens(nil)
				format(got[i])
			}

			if !reflect.DeepEqual(got, test.want) {
				diff := cmp.Diff(test.want, got)
				t.Errorf("wrong result\ngot:  %s\nwant: %s\ndiff:\n%s", spew.Sdump(got), spew.Sdump(test.want), diff)
			}
		})
	}
}

func TestTupleConsExprAddRaw(t *testing.T) {
	tests := []struct {
		testName string
		src      string
		attrName string
		tokens   Tokens // Tokens constituting a new tuple item.
		want     Tokens // Tokens for the tuple.
	}{
		{
			testName: `simple tuple`,
			src:      `a = ["derby", "calico"]`,
			attrName: `a`,
			tokens: Tokens{
				{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte("var"),
				},
				{
					Type:  hclsyntax.TokenDot,
					Bytes: []byte{'.'},
				},
				{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte("bat"),
				},
			},
			want: Tokens{
				{
					Type:  hclsyntax.TokenOBrack,
					Bytes: []byte{'['},
				},
				{
					Type:  hclsyntax.TokenOQuote,
					Bytes: []byte{'"'},
				},
				{
					Type:  hclsyntax.TokenQuotedLit,
					Bytes: []byte("derby"),
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
					Type:  hclsyntax.TokenComma,
					Bytes: []byte{','},
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte("var"),
					SpacesBefore: 1,
				},
				{
					Type:  hclsyntax.TokenDot,
					Bytes: []byte{'.'},
				},
				{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte("bat"),
				},
				{
					Type:  hclsyntax.TokenCBrack,
					Bytes: []byte{']'},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			f, diags := ParseConfig([]byte(test.src), "", hcl.InitialPos)
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

			expr := attr.Expr().AsTupleConsExpr()

			// Append twice, to observe set semantics.
			expr.AddRaw(test.tokens)
			expr.AddRaw(test.tokens)

			got := expr.BuildTokens(nil)
			format(got)

			if !reflect.DeepEqual(got, test.want) {
				diff := cmp.Diff(test.want, got)
				t.Errorf("wrong result\ngot:  %s\nwant: %s\ndiff:\n%s", spew.Sdump(got), spew.Sdump(test.want), diff)
			}
		})
	}
}

func TestTupleConsExprAddValue(t *testing.T) {
	tests := []struct {
		testName string
		src      string
		attrName string
		val      cty.Value // Value desired for a new tuple item.
		want     Tokens    // Tokens for the tuple.
	}{
		{
			testName: `simple tuple`,
			src:      `a = ["derby", "calico"]`,
			attrName: `a`,
			val:      cty.StringVal(`fruit bat`),
			want: Tokens{
				{
					Type:  hclsyntax.TokenOBrack,
					Bytes: []byte{'['},
				},
				{
					Type:  hclsyntax.TokenOQuote,
					Bytes: []byte{'"'},
				},
				{
					Type:  hclsyntax.TokenQuotedLit,
					Bytes: []byte("derby"),
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
					Type:  hclsyntax.TokenComma,
					Bytes: []byte{','},
				},
				{
					Type:         hclsyntax.TokenOQuote,
					Bytes:        []byte{'"'},
					SpacesBefore: 1,
				},
				{
					Type:  hclsyntax.TokenQuotedLit,
					Bytes: []byte("fruit bat"),
				},
				{
					Type:  hclsyntax.TokenCQuote,
					Bytes: []byte{'"'},
				},
				{
					Type:  hclsyntax.TokenCBrack,
					Bytes: []byte{']'},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			f, diags := ParseConfig([]byte(test.src), "", hcl.InitialPos)
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

			expr := attr.Expr().AsTupleConsExpr()

			// Append twice, to observe set semantics.
			expr.AddValue(test.val)
			expr.AddValue(test.val)

			got := expr.BuildTokens(nil)
			format(got)

			if !reflect.DeepEqual(got, test.want) {
				diff := cmp.Diff(test.want, got)
				t.Errorf("wrong result\ngot:  %s\nwant: %s\ndiff:\n%s", spew.Sdump(got), spew.Sdump(test.want), diff)
			}
		})
	}
}
func TestTupleConsExprAppendRaw(t *testing.T) {
	tests := []struct {
		testName string
		src      string
		attrName string
		tokens   Tokens // Tokens constituting a new tuple item.
		want     Tokens // Tokens for the tuple.
	}{
		{
			testName: `simple tuple`,
			src:      `a = ["derby", "calico"]`,
			attrName: `a`,
			tokens: Tokens{
				{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte("var"),
				},
				{
					Type:  hclsyntax.TokenDot,
					Bytes: []byte{'.'},
				},
				{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte("bat"),
				},
			},
			want: Tokens{
				{
					Type:  hclsyntax.TokenOBrack,
					Bytes: []byte{'['},
				},
				{
					Type:  hclsyntax.TokenOQuote,
					Bytes: []byte{'"'},
				},
				{
					Type:  hclsyntax.TokenQuotedLit,
					Bytes: []byte("derby"),
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
					Type:  hclsyntax.TokenComma,
					Bytes: []byte{','},
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte("var"),
					SpacesBefore: 1,
				},
				{
					Type:  hclsyntax.TokenDot,
					Bytes: []byte{'.'},
				},
				{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte("bat"),
				},
				{
					Type:  hclsyntax.TokenComma,
					Bytes: []byte{','},
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte("var"),
					SpacesBefore: 1,
				},
				{
					Type:  hclsyntax.TokenDot,
					Bytes: []byte{'.'},
				},
				{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte("bat"),
				},
				{
					Type:  hclsyntax.TokenCBrack,
					Bytes: []byte{']'},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			f, diags := ParseConfig([]byte(test.src), "", hcl.InitialPos)
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

			expr := attr.Expr().AsTupleConsExpr()

			// Append twice, to observe list (not set) semantics.
			expr.AppendRaw(test.tokens)
			expr.AppendRaw(test.tokens)

			got := expr.BuildTokens(nil)
			format(got)

			if !reflect.DeepEqual(got, test.want) {
				diff := cmp.Diff(test.want, got)
				t.Errorf("wrong result\ngot:  %s\nwant: %s\ndiff:\n%s", spew.Sdump(got), spew.Sdump(test.want), diff)
			}
		})
	}
}

func TestTupleConsExprAppendValue(t *testing.T) {
	tests := []struct {
		testName string
		src      string
		attrName string
		val      cty.Value // Value desired for a new tuple item.
		want     Tokens    // Tokens for the tuple.
	}{
		{
			testName: `simple tuple`,
			src:      `a = ["derby", "calico"]`,
			attrName: `a`,
			val:      cty.StringVal(`fruit bat`),
			want: Tokens{
				{
					Type:  hclsyntax.TokenOBrack,
					Bytes: []byte{'['},
				},
				{
					Type:  hclsyntax.TokenOQuote,
					Bytes: []byte{'"'},
				},
				{
					Type:  hclsyntax.TokenQuotedLit,
					Bytes: []byte("derby"),
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
					Type:  hclsyntax.TokenComma,
					Bytes: []byte{','},
				},
				{
					Type:         hclsyntax.TokenOQuote,
					Bytes:        []byte{'"'},
					SpacesBefore: 1,
				},
				{
					Type:  hclsyntax.TokenQuotedLit,
					Bytes: []byte("fruit bat"),
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
					Type:         hclsyntax.TokenOQuote,
					Bytes:        []byte{'"'},
					SpacesBefore: 1,
				},
				{
					Type:  hclsyntax.TokenQuotedLit,
					Bytes: []byte("fruit bat"),
				},
				{
					Type:  hclsyntax.TokenCQuote,
					Bytes: []byte{'"'},
				},
				{
					Type:  hclsyntax.TokenCBrack,
					Bytes: []byte{']'},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			f, diags := ParseConfig([]byte(test.src), "", hcl.InitialPos)
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

			expr := attr.Expr().AsTupleConsExpr()

			// Append twice, to observe list (not set) semantics.
			expr.AppendValue(test.val)
			expr.AppendValue(test.val)

			got := expr.BuildTokens(nil)
			format(got)

			if !reflect.DeepEqual(got, test.want) {
				diff := cmp.Diff(test.want, got)
				t.Errorf("wrong result\ngot:  %s\nwant: %s\ndiff:\n%s", spew.Sdump(got), spew.Sdump(test.want), diff)
			}
		})
	}
}
