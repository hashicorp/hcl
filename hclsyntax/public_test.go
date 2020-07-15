package hclsyntax

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/zclconf/go-cty/cty"
)

func TestValidIdentifier(t *testing.T) {
	tests := []struct {
		Input string
		Want  bool
	}{
		{"", false},
		{"hello", true},
		{"hello.world", false},
		{"hello ", false},
		{" hello", false},
		{"hello\n", false},
		{"hello world", false},
		{"aws_instance", true},
		{"aws.instance", false},
		{"foo-bar", true},
		{"foo--bar", true},
		{"foo_", true},
		{"foo-", true},
		{"_foobar", true},
		{"-foobar", false},
		{"blah1", true},
		{"blah1blah", true},
		{"1blah1blah", false},
		{"héllo", true}, // combining acute accent
		{"Χαίρετε", true},
		{"звать", true},
		{"今日は", true},
		{"\x80", false},  // UTF-8 continuation without an introducer
		{"a\x80", false}, // UTF-8 continuation after a non-introducer
	}

	for _, test := range tests {
		t.Run(test.Input, func(t *testing.T) {
			got := ValidIdentifier(test.Input)
			if got != test.Want {
				t.Errorf("wrong result %#v; want %#v", got, test.Want)
			}
		})
	}
}

func TestParseBlockFromTokens_withoutNewline(t *testing.T) {
	_, diags := ParseBlockFromTokens(testBlockTokensWithoutNewline)
	if len(diags) != 1 {
		t.Fatalf("Expected exactly 1 diagnostic, %d given", len(diags))
	}
}

func TestParseBlockFromTokens_block(t *testing.T) {
	b, diags := ParseBlockFromTokens(testBlockTokensWithNewline)
	if len(diags) > 0 {
		t.Fatal(diags)
	}
	expectedBlock := &Block{
		Type:   "blocktype",
		Labels: []string{"onelabel"},
		Body: &Body{
			Attributes: Attributes{
				"attr": &Attribute{
					Name: "attr",
					Expr: &LiteralValueExpr{
						Val: cty.NumberIntVal(42),
					},
				},
			},
			Blocks: Blocks{},
		},
	}
	opts := cmp.Options{
		cmpopts.IgnoreUnexported(Body{}),
		cmpopts.IgnoreUnexported(cty.Value{}),
	}
	opts = append(opts, optsIgnoreRanges...)
	if diff := cmp.Diff(expectedBlock, b, opts); diff != "" {
		t.Fatalf("Blocks don't match:\n%s", diff)
	}
}

func TestParseBlockFromTokens_invalid(t *testing.T) {
	_, diags := ParseBlockFromTokens(invalidTokens)
	if len(diags) != 1 {
		t.Fatalf("Expected exactly 1 diagnostic, %d given", len(diags))
	}
}

func TestParseBlockFromTokens_invalidBlock(t *testing.T) {
	_, diags := ParseBlockFromTokens(invalidBlockTokens)
	if len(diags) != 1 {
		t.Fatalf("Expected exactly 1 diagnostic, %d given", len(diags))
	}
}

func TestParseBlockFromTokens_attr(t *testing.T) {
	_, diags := ParseBlockFromTokens(testAttributeTokensValid)
	if len(diags) != 1 {
		t.Fatalf("Expected exactly 1 diagnostic, given:\n%#v", diags)
	}
}

func TestParseAttributeFromTokens_attr(t *testing.T) {
	b, diags := ParseAttributeFromTokens(testAttributeTokensValid)
	if len(diags) > 0 {
		t.Fatal(diags)
	}
	expectedAttribute := &Attribute{
		Name: "attr",
		Expr: &LiteralValueExpr{
			Val: cty.NumberIntVal(79),
		},
	}
	opts := cmp.Options{
		cmpopts.IgnoreFields(Token{}, "Range"),
		cmpopts.IgnoreUnexported(Attribute{}),
		cmpopts.IgnoreUnexported(cty.Value{}),
	}
	if diff := cmp.Diff(expectedAttribute, b, opts); diff != "" {
		t.Fatalf("Blocks don't match:\n%s", diff)
	}
}

func TestParseAttributeFromTokens_invalid(t *testing.T) {
	_, diags := ParseAttributeFromTokens(invalidTokens)
	if len(diags) != 1 {
		t.Fatalf("Expected exactly 1 diagnostic, %d given", len(diags))
	}
}

func TestParseAttributeFromTokens_block(t *testing.T) {
	_, diags := ParseAttributeFromTokens(testBlockTokensWithNewline)
	if len(diags) != 1 {
		t.Fatalf("Expected exactly 1 diagnostic, given:\n%#v", diags)
	}
}

var optsIgnoreRanges = []cmp.Option{
	cmpopts.IgnoreFields(Token{}, "Range"),
	cmpopts.IgnoreFields(Attribute{}, "SrcRange", "NameRange", "EqualsRange"),
	cmpopts.IgnoreFields(Block{}, "TypeRange", "LabelRanges", "OpenBraceRange", "CloseBraceRange"),
	cmpopts.IgnoreFields(LiteralValueExpr{}, "SrcRange"),
	cmpopts.IgnoreFields(Body{}, "SrcRange", "EndRange"),
}

var testAttributeTokensValid = Tokens{
	{Type: TokenIdent, Bytes: []byte("attr")},
	{Type: TokenEqual, Bytes: []byte("=")},
	{Type: TokenNumberLit, Bytes: []byte("79")},
	{Type: TokenNewline, Bytes: []byte("\n")},
}

var testBlockTokensWithNewline = Tokens{
	{Type: TokenIdent, Bytes: []byte("blocktype")},
	{Type: TokenOQuote, Bytes: []byte(`"`)},
	{Type: TokenQuotedLit, Bytes: []byte("onelabel")},
	{Type: TokenCQuote, Bytes: []byte(`"`)},
	{Type: TokenOBrace, Bytes: []byte("{")},
	{Type: TokenNewline, Bytes: []byte("\n")},
	{Type: TokenIdent, Bytes: []byte("attr")},
	{Type: TokenEqual, Bytes: []byte("=")},
	{Type: TokenNumberLit, Bytes: []byte("42")},
	{Type: TokenNewline, Bytes: []byte("\n")},
	{Type: TokenCBrace, Bytes: []byte("}")},
	{Type: TokenNewline, Bytes: []byte("\n")},
}

var testBlockTokensWithoutNewline = Tokens{
	{Type: TokenIdent, Bytes: []byte("blocktype")},
	{Type: TokenOQuote, Bytes: []byte(`"`)},
	{Type: TokenQuotedLit, Bytes: []byte("onelabel")},
	{Type: TokenCQuote, Bytes: []byte(`"`)},
	{Type: TokenOBrace, Bytes: []byte("{")},
	{Type: TokenNewline, Bytes: []byte("\n")},
	{Type: TokenIdent, Bytes: []byte("attr")},
	{Type: TokenEqual, Bytes: []byte("=")},
	{Type: TokenNumberLit, Bytes: []byte("42")},
	{Type: TokenNewline, Bytes: []byte("\n")},
	{Type: TokenCBrace, Bytes: []byte("}")},
}

var invalidBlockTokens = Tokens{
	{Type: TokenIdent, Bytes: []byte("variable")},
	{Type: TokenOQuote, Bytes: []byte(`"`)},
	{Type: TokenQuotedLit, Bytes: []byte("name")},
	{Type: TokenCQuote, Bytes: []byte(`"`)},
	{Type: TokenOBrace, Bytes: []byte("{")},
	{Type: TokenNewline, Bytes: []byte("\n")},
	{Type: TokenIdent, Bytes: []byte("default")},
	{Type: TokenEqual, Bytes: []byte("=")},
	{Type: TokenOQuote, Bytes: []byte(`"`)},
	{Type: TokenQuotedNewline, Bytes: []byte("\n")},
	{Type: TokenQuotedLit, Bytes: []byte("}")}, // TODO: Fix the tokenizer so this comes back as TokenCBrace
	{Type: TokenNewline, Bytes: []byte("\n")},
}

var invalidTokens = Tokens{
	{Type: TokenNewline, Bytes: []byte("\n")},
}
