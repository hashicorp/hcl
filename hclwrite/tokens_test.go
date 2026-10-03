package hclwrite

import (
	"testing"

	"github.com/zclconf/go-cty/cty"
)

// newIdentToken wrapped any Go string as a raw TokenIdent with no
// validation. AppendNewBlock's typeName and SetAttributeValue's name both
// go through it directly, so a name containing HCL syntax (braces,
// newlines, quotes) was emitted verbatim and the package silently produced
// a document with extra blocks or attributes the caller never asked for,
// rather than reporting the bad name.

func mustPanic(t *testing.T, name string, f func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatalf("%s: expected a panic for an invalid HCL identifier, got none", name)
		}
	}()
	f()
}

func TestBlockTypeNameInjectionPanics(t *testing.T) {
	f := NewEmptyFile()
	body := f.Body()
	evilType := "resource \"aws_instance\" \"x\" {\n  user_data = \"evil\"\n}\nresource \"noop"
	mustPanic(t, "AppendNewBlock", func() {
		body.AppendNewBlock(evilType, []string{"legit"})
	})
}

func TestSetTypeInjectionPanics(t *testing.T) {
	block := NewBlock("resource", []string{"legit"})
	mustPanic(t, "SetType", func() {
		block.SetType("resource \"x\" {\n}\nresource \"evil")
	})
}

func TestAttributeNameInjectionPanics(t *testing.T) {
	f := NewEmptyFile()
	body := f.Body()
	evilName := "ok\"\n}\nresource \"evil\" \"y"
	mustPanic(t, "SetAttributeValue", func() {
		body.SetAttributeValue(evilName, cty.StringVal("x"))
	})
}

func TestTokensForIdentifierInjectionPanics(t *testing.T) {
	mustPanic(t, "TokensForIdentifier", func() {
		TokensForIdentifier("1 + 1")
	})
}

func TestValidNamesStillWork(t *testing.T) {
	f := NewEmptyFile()
	body := f.Body()
	block := body.AppendNewBlock("resource", []string{"aws_instance", "x"})
	block.Body().SetAttributeValue("ami", cty.StringVal("ok"))

	got := string(f.Bytes())
	want := "resource \"aws_instance\" \"x\" {\n  ami = \"ok\"\n}\n"
	if got != want {
		t.Fatalf("unexpected output:\ngot:  %q\nwant: %q", got, want)
	}
}

// Labels were checked separately: they already go through
// TokensForValue(cty.StringVal(label)), which quotes and escapes the
// string properly, so a label can't break out of its quoted position.
// Confirmed locally - not re-asserted here since it was never broken.
