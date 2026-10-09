// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package hclwrite

import (
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func assertPanic(t *testing.T, desc string, f func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatalf("%s: expected panic for invalid identifier, got none", desc)
		}
	}()
	f()
}

func TestAppendNewBlockInvalidIdentifierPanics(t *testing.T) {
	f := NewEmptyFile()
	body := f.Body()
	evilType := "resource \"aws_instance\" \"x\" {\n  user_data = \"evil\"\n}\nresource \"noop"
	assertPanic(t, "AppendNewBlock with injection string", func() {
		body.AppendNewBlock(evilType, []string{"legit"})
	})
	assertPanic(t, "AppendNewBlock with empty string", func() {
		body.AppendNewBlock("", []string{"legit"})
	})
	assertPanic(t, "AppendNewBlock with whitespace", func() {
		body.AppendNewBlock("invalid type", []string{"legit"})
	})
}

func TestNewBlockInvalidIdentifierPanics(t *testing.T) {
	assertPanic(t, "NewBlock with spaces", func() {
		NewBlock("bad block", nil)
	})
	assertPanic(t, "NewBlock with digit prefix", func() {
		NewBlock("123bad", nil)
	})
}

func TestSetTypeInvalidIdentifierPanics(t *testing.T) {
	block := NewBlock("resource", []string{"legit"})
	assertPanic(t, "SetType with newline injection", func() {
		block.SetType("resource \"x\" {\n}\nresource \"evil")
	})
	assertPanic(t, "SetType with operator", func() {
		block.SetType("a+b")
	})
}

func TestSetAttributeValueInvalidIdentifierPanics(t *testing.T) {
	f := NewEmptyFile()
	body := f.Body()
	evilName := "ok\"\n}\nresource \"evil\" \"y"
	assertPanic(t, "SetAttributeValue with quote injection", func() {
		body.SetAttributeValue(evilName, cty.StringVal("x"))
	})
	assertPanic(t, "SetAttributeValue with empty name", func() {
		body.SetAttributeValue("", cty.StringVal("x"))
	})
}

func TestAttributeSetNameInvalidIdentifierPanics(t *testing.T) {
	f := NewEmptyFile()
	body := f.Body()
	attr := body.SetAttributeValue("initial", cty.StringVal("val"))
	assertPanic(t, "setName with invalid name", func() {
		attr.setName("bad name")
	})
}

func TestRenameAttributeInvalidIdentifierPanics(t *testing.T) {
	f := NewEmptyFile()
	body := f.Body()
	body.SetAttributeValue("initial", cty.StringVal("val"))
	assertPanic(t, "RenameAttribute with invalid name", func() {
		body.RenameAttribute("initial", "bad name")
	})
}

func TestSetAttributeRawInvalidIdentifierPanics(t *testing.T) {
	f := NewEmptyFile()
	body := f.Body()
	assertPanic(t, "SetAttributeRaw with invalid name", func() {
		body.SetAttributeRaw("bad attr", Tokens{
			{
				Type:  TokensForIdentifier("val")[0].Type,
				Bytes: []byte("val"),
			},
		})
	})
}

func TestTokensForIdentifierInvalidPanics(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"arithmetic expression", "1 + 1"},
		{"starts with digit", "0bad"},
		{"contains dot", "foo.bar"},
		{"contains hyphen and spaces", "foo - bar"},
		{"newline character", "foo\nbar"},
		{"curly brace", "foo{bar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertPanic(t, tt.name, func() {
				TokensForIdentifier(tt.input)
			})
		})
	}
}

func TestValidIdentifiersAccepted(t *testing.T) {
	validNames := []string{
		"simple",
		"_leading_underscore",
		"camelCaseName",
		"PascalCaseName",
		"with_underscores_123",
		"kebab-case-name",
		"has-hyphens-and-numbers-1",
	}

	for _, name := range validNames {
		t.Run(name, func(t *testing.T) {
			toks := TokensForIdentifier(name)
			if len(toks) != 1 {
				t.Fatalf("expected 1 token for %q, got %d", name, len(toks))
			}
			if string(toks[0].Bytes) != name {
				t.Fatalf("expected token bytes %q, got %q", name, string(toks[0].Bytes))
			}
		})
	}

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
