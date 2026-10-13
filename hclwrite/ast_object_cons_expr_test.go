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
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte("bat"),
					SpacesBefore: 2,
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
					Type:  hclsyntax.TokenNewline,
					Bytes: []byte("\n"),
				},
				{
					Type:  hclsyntax.TokenCBrace,
					Bytes: []byte{'}'},
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
		_, o := parseTestObjectA(t, "a = { hat = \"derby\", cat = \"calico\" }\n")
		attached := o.ValueExprFor("hat").Expr()

		assertPanics(t, func() { o.SetItem("cat", attached) })
	})

	t.Run("new expression attached by a previous SetItem", func(t *testing.T) {
		_, o := parseTestObjectA(t, "a = { hat = \"derby\", cat = \"calico\" }\n")
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

// setObjectItem parses src, applies fn to the object-construct expression
// assigned to attribute "a", and returns the resulting source. It fails the
// test if the result is not valid HCL.
func setObjectItem(t *testing.T, src string, fn func(*ObjectConsExpr)) string {
	t.Helper()

	f, diags := ParseConfig([]byte(src), "", hcl.InitialPos)
	if diags.HasErrors() {
		t.Fatalf("unexpected parse errors: %s", diags)
	}
	obj := f.Body().GetAttribute("a").Expr().AsObjectConsExpr()
	if obj == nil {
		t.Fatalf("attribute a is not an object-construct expression")
	}

	fn(obj)

	got := f.Bytes()
	if _, diags := hclsyntax.ParseConfig(got, "", hcl.InitialPos); diags.HasErrors() {
		t.Fatalf("result is not valid HCL: %s\n%s", diags, got)
	}
	return string(got)
}

func TestObjectConsExprItemFor_quotedKey(t *testing.T) {
	setObjectItem(t, `a = { "hat" = "derby" }`+"\n", func(o *ObjectConsExpr) {
		if o.ItemFor("hat") == nil {
			t.Error(`ItemFor("hat") did not find the quoted key "hat"`)
		}
	})
}

// Keys that aren't valid identifiers must be written as quoted strings.
func TestObjectConsExprSetItem_nonIdentifierKey(t *testing.T) {
	tests := map[string]string{
		"space":       "a hat",
		"dot":         "hat.band",
		"leading num": "1hat",
	}

	for name, key := range tests {
		t.Run(name, func(t *testing.T) {
			// setObjectItem fails the test if the output doesn't parse.
			got := setObjectItem(t, "a = {\n  cat = \"calico\"\n}\n", func(o *ObjectConsExpr) {
				o.SetItemValue(key, cty.StringVal("bowler"))

				// A key that was just set must be findable again.
				if o.ItemFor(key) == nil {
					t.Errorf("ItemFor(%q) did not find the item just set", key)
				}
			})

			// Format aligns the "=" signs, which varies with key length.
			want := string(Format([]byte("a = {\n  cat = \"calico\"\n  \"" + key + "\" = \"bowler\"\n}\n")))
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("wrong result (-want +got):\n%s", diff)
			}
		})
	}
}

// parseTestObjectA parses src and returns the file along with the
// object-construct expression assigned to attribute "a".
func parseTestObjectA(t *testing.T, src string) (*File, *ObjectConsExpr) {
	t.Helper()
	f, diags := ParseConfig([]byte(src), "", hcl.InitialPos)
	if diags.HasErrors() {
		t.Fatalf("unexpected parse errors: %s", diags)
	}
	return f, f.Body().GetAttribute("a").Expr().AsObjectConsExpr()
}

// Shows why the missing panic matters: after reusing an expression, both items
// share it, so changing one item silently changes the other too.
func TestObjectConsExprSetItem_attachedExprAliasing(t *testing.T) {
	f, o := parseTestObjectA(t, "a = { hat = var.derby, cat = \"calico\" }\n")

	func() {
		defer func() { _ = recover() }()
		o.SetItem("cat", o.ValueExprFor("hat").Expr())
	}()

	o.ValueExprFor("hat").Expr().RenameVariablePrefix([]string{"var"}, []string{"local"})

	want := "a = { hat = local.derby, cat = \"calico\" }\n"
	if got := string(f.Bytes()); got != want {
		t.Errorf("renaming hat's variable also changed cat\ngot:  %swant: %s", got, want)
	}
}

// Lower-level demonstration of the root cause: attaching an Expression never
// records its parent, so inTree.assertUnattached can never fire.
func TestInTreeAssertUnattached_afterAppend(t *testing.T) {
	expr := NewExpressionLiteral(cty.StringVal("bowler"))
	parent := &nodes{}
	parent.Append(expr)

	if expr.parent == nil {
		t.Error("parent was not recorded when the expression was attached")
	}
	assertPanics(t, func() { expr.assertUnattached() })
}

// New items should follow the object's existing layout, as a person would
// write them. In a multi-line object each new item goes on its own indented
// line with the closing brace on its own line; a single-line object stays on
// one line, with comma separators. Any existing trailing comma is left where
// it was. Format can't repair these cases, so SetItem has to get them right.
func TestObjectConsExprSetItem_layout(t *testing.T) {
	tests := map[string]struct {
		src  string
		set  []string // keys to add, each set to "bowler"
		want string
	}{
		"multi-line object": {
			src: `
a = {
  hat = "derby"
}
`,
			set: []string{"dog"},
			want: `
a = {
  hat = "derby"
  dog = "bowler"
}
`,
		},
		// PR #811 currently produces:
		//   a = {
		//   dog = "bowler" }
		"empty object": {
			src: `
a = {}
`,
			set: []string{"dog"},
			want: `
a = { dog = "bowler" }
`,
		},
		// PR #811 currently produces:
		//   a = {
		//     dog = "bowler"
		//   cat = "bowler" }
		"empty object, two new items": {
			src: `
a = {}
`,
			set: []string{"dog", "cat"},
			want: `
a = { dog = "bowler", cat = "bowler" }
`,
		},
		// PR #811 currently produces:
		//   a = { hat = "derby"
		//   dog = "bowler" }
		"single-line object": {
			src: `
a = { hat = "derby" }
`,
			set: []string{"dog"},
			want: `
a = { hat = "derby", dog = "bowler" }
`,
		},
		"single-line object, two new items": {
			src: `
a = { hat = "derby" }
`,
			set: []string{"dog", "cat"},
			want: `
a = { hat = "derby", dog = "bowler", cat = "bowler" }
`,
		},
		"single-line object with trailing comma": {
			src: `
a = { hat = "derby", }
`,
			set: []string{"dog"},
			want: `
a = { hat = "derby", dog = "bowler" }
`,
		},
		"multi-line object with trailing comma": {
			src: `
a = {
  hat = "derby",
}
`,
			set: []string{"dog"},
			want: `
a = {
  hat = "derby",
  dog = "bowler"
}
`,
		},
		"multi-line object with items sharing a line": {
			src: `
a = {
  hat = "derby", cat = "calico"
}
`,
			set: []string{"dog"},
			want: `
a = {
  hat = "derby", cat = "calico"
  dog = "bowler"
}
`,
		},
		"multi-line object with brace after last item": {
			src: `
a = {
  hat = "derby" }
`,
			set: []string{"dog"},
			want: `
a = {
  hat = "derby"
  dog = "bowler"
}
`,
		},
		"comment before closing brace": {
			src: `
a = {
  hat = "derby"

  # end of a
}
`,
			set: []string{"dog", "cat"},
			want: `
a = {
  hat = "derby"
  dog = "bowler"
  cat = "bowler"

  # end of a
}
`,
		},
		// Format puts a line starting with an inline comment followed by the
		// closing brace at the outer indentation level.
		"inline comment before closing brace": {
			src: `
a = {
  hat = "derby"
/* end of a */ }
`,
			set: []string{"dog"},
			want: `
a = {
  hat = "derby"
  dog = "bowler"
/* end of a */ }
`,
		},
		"line comment after last item": {
			src: `
a = {
  hat = "derby", # a hat
  # end of a
}
`,
			set: []string{"dog"},
			want: `
a = {
  hat = "derby", # a hat
  dog = "bowler"
  # end of a
}
`,
		},
		"comment in empty multi-line object": {
			src: `
a = {
  # nothing yet
}
`,
			set: []string{"dog"},
			want: `
a = {
  dog = "bowler"
  # nothing yet
}
`,
		},
		"nested object": {
			src: `
a = {
  hat = {
    style = "derby"
  }
}
`,
			set: []string{"dog"},
			want: `
a = {
  hat = {
    style = "derby"
  }
  dog = "bowler"
}
`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			src := strings.TrimPrefix(tc.src, "\n")
			want := strings.TrimPrefix(tc.want, "\n")

			f, diags := ParseConfig([]byte(src), "", hcl.InitialPos)
			if diags.HasErrors() {
				t.Fatalf("unexpected parse errors: %s", diags)
			}
			o := f.Body().GetAttribute("a").Expr().AsObjectConsExpr()
			for _, key := range tc.set {
				o.SetItemValue(key, cty.StringVal("bowler"))
			}
			got := string(f.Bytes())

			// The result must always be valid HCL...
			if _, diags := hclsyntax.ParseConfig([]byte(got), "", hcl.InitialPos); diags.HasErrors() {
				t.Fatalf("result is not valid HCL: %s\n%s", diags, got)
			}
			// ...laid out the way a person would write it...
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("wrong layout\n--- got:\n%s\n--- want:\n%s\n--- diff (-want +got):\n%s", got, want, diff)
			}
			// ...and already in canonical form, so callers don't need Format.
			if formatted := string(Format([]byte(got))); formatted != got {
				t.Errorf("result changes under Format:\n%s", formatted)
			}
		})
	}
}
