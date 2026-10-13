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

	//nolint:errcheck
	object.ValueExprFor("hat").expr.BuildTokens(nil).WriteTo(&actual)
	expected := `"fez"`
	if diff := cmp.Diff(actual.String(), expected); len(diff) > 0 {
		t.Error(diff)
	}
}

func TestObjectConsExprItemFor(t *testing.T) {
	tests := map[string]struct {
		src  string
		key  string
		want bool
	}{
		"identifier":         {`a = { hat = "derby" }`, "hat", true},
		"quoted":             {`a = { "hat" = "derby" }`, "hat", true},
		"quoted with escape": {`a = { "h\u0061t" = "derby" }`, "hat", true},
		"quoted with space":  {`a = { "a hat" = "derby" }`, "a hat", true},
		"quoted no match":    {`a = { "hat" = "derby" }`, "cat", false},
		"interpolated":       {`a = { "${hat}" = "derby" }`, "hat", false},
		"escaped template":   {`a = { "$${hat}" = "derby" }`, "${hat}", true},
		"parenthesized":      {`a = { (hat) = "derby" }`, "hat", false},
		"traversal":          {`a = { hat.band = "derby" }`, "hat", false},
		"later quoted item":  {`a = { cat = "calico", "hat" = "derby" }`, "hat", true},
		"missing key":        {`a = { cat = "calico" }`, "hat", false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			f, diags := ParseConfig([]byte(tc.src+"\n"), "", hcl.InitialPos)
			if diags.HasErrors() {
				t.Fatalf("unexpected parse errors: %s", diags)
			}
			obj := f.Body().GetAttribute("a").Expr().AsObjectConsExpr()

			if got := obj.ItemFor(tc.key) != nil; got != tc.want {
				t.Errorf("ItemFor(%q) found = %t, want %t", tc.key, got, tc.want)
			}
		})
	}
}

// New items with keys that need escaping must produce valid HCL whose key
// evaluates back to the original string, and must be findable again.
func TestObjectConsExprSetItem_escapedKey(t *testing.T) {
	tests := map[string]struct {
		key     string
		wantKey string
	}{
		"double quote":  {`say "hi"`, `"say \"hi\""`},
		"backslash":     {`a\b`, `"a\\b"`},
		"interpolation": {`${hat}`, `"$${hat}"`},
		"directive":     {`%{hat}`, `"%%{hat}"`},
		"empty":         {``, `""`},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			f, diags := ParseConfig([]byte("a = {}\n"), "", hcl.InitialPos)
			if diags.HasErrors() {
				t.Fatalf("unexpected parse errors: %s", diags)
			}
			obj := f.Body().GetAttribute("a").Expr().AsObjectConsExpr()
			obj.SetItemValue(tc.key, cty.StringVal("bowler"))

			if obj.ItemFor(tc.key) == nil {
				t.Errorf("ItemFor(%q) did not find the item just set", tc.key)
			}

			got := f.Bytes()
			if !strings.Contains(string(got), tc.wantKey+" = ") {
				t.Errorf("key not written as %s:\n%s", tc.wantKey, got)
			}

			native, diags := hclsyntax.ParseConfig(got, "", hcl.InitialPos)
			if diags.HasErrors() {
				t.Fatalf("result is not valid HCL: %s\n%s", diags, got)
			}
			attrs, _ := native.Body.JustAttributes()
			val, diags := attrs["a"].Expr.Value(nil)
			if diags.HasErrors() {
				t.Fatalf("unexpected errors evaluating result: %s", diags)
			}
			want := cty.ObjectVal(map[string]cty.Value{tc.key: cty.StringVal("bowler")})
			if !val.RawEquals(want) {
				t.Errorf("wrong value\ngot:  %#v\nwant: %#v", val, want)
			}
		})
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
			// Set the content of an existing item.
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
		{

			// Set the content of new item.
			`a = {
				hat = "derby", (cat) = "calico" }` + "\n",
			"a",
			"bat",
			Tokens{
				{
					Type:         hclsyntax.TokenOQuote,
					Bytes:        []byte(`"`),
					SpacesBefore: 1,
				},
				{
					Type:  hclsyntax.TokenQuotedLit,
					Bytes: []byte(`fruit bat`),
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
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte("hat"),
					SpacesBefore: 2, // TODO: y?
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
					Bytes: []byte(`derby`),
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
					Type:  hclsyntax.TokenNewline,
					Bytes: []byte("\n"),
				},
				{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte("bat"),
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
					Bytes: []byte(`fruit bat`),
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

// assertPanics fails the test if fn returns without panicking.
func assertPanics(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Error("expected a panic, but there was none")
		}
	}()
	fn()
}

// The doc comment on SetItem says it panics if the expression is already
// attached, because otherwise one node ends up in two places in the tree.
func TestObjectConsExprSetItem_attachedExprPanics(t *testing.T) {
	t.Run("expression attached by the parser", func(t *testing.T) {
		_, o := parseObjectA(t, "a = { hat = \"derby\", cat = \"calico\" }\n")
		attached := o.ValueExprFor("hat").Expr()

		assertPanics(t, func() { o.SetItem("cat", attached) })
	})

	t.Run("new expression attached by a previous SetItem", func(t *testing.T) {
		_, o := parseObjectA(t, "a = { hat = \"derby\", cat = \"calico\" }\n")
		expr := NewExpressionLiteral(cty.StringVal("bowler"))
		o.SetItem("hat", expr)

		assertPanics(t, func() { o.SetItem("cat", expr) })
	})
}

// Quoted string-literal keys are equivalent to identifier keys, so they should
// be found and replaced rather than duplicated.
func TestObjectConsExprSetItem_quotedKey(t *testing.T) {
	tests := map[string]struct {
		src  string
		want string
	}{
		"single-line": {
			src:  `a = { "hat" = "derby" }` + "\n",
			want: `a = { "hat" = "bowler" }` + "\n",
		},
		"multi-line": {
			src:  "a = {\n  \"hat\" = \"derby\"\n  cat = \"calico\"\n}\n",
			want: "a = {\n  \"hat\" = \"bowler\"\n  cat   = \"calico\"\n}\n",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := setObjectItem(t, tc.src, func(o *ObjectConsExpr) {
				o.SetItemValue("hat", cty.StringVal("bowler"))
			})
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("wrong result (-want +got):\n%s", diff)
			}
		})
	}
}
