package hclwrite

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func TestBodyGetAttribute(t *testing.T) {
	tests := []struct {
		src  string
		name string
		want Tokens
	}{
		{
			"",
			"a",
			nil,
		},
		{
			"a = 1\n",
			"a",
			Tokens{
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte{'a'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte{'='},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNumberLit,
					Bytes:        []byte{'1'},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
			},
		},
		{
			"a = 1\nb = 1\nc = 1\n",
			"a",
			Tokens{
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte{'a'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte{'='},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNumberLit,
					Bytes:        []byte{'1'},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
			},
		},
		{
			"a = 1\nb = 2\nc = 3\n",
			"b",
			Tokens{
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte{'b'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte{'='},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNumberLit,
					Bytes:        []byte{'2'},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
			},
		},
		{
			"a = 1\nb = 2\nc = 3\n",
			"c",
			Tokens{
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte{'c'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte{'='},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNumberLit,
					Bytes:        []byte{'3'},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
			},
		},
		{
			"a = 1\n# b is a b\nb = 2\nc = 3\n",
			"b",
			Tokens{
				{
					// Recognized as a lead comment and so attached to the attribute
					Type:         hclsyntax.TokenComment,
					Bytes:        []byte("# b is a b\n"),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte{'b'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte{'='},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNumberLit,
					Bytes:        []byte{'2'},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
			},
		},
		{
			"a = 1\n# not attached to a or b\n\nb = 2\nc = 3\n",
			"b",
			Tokens{
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte{'b'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte{'='},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNumberLit,
					Bytes:        []byte{'2'},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s in %s", test.name, test.src), func(t *testing.T) {
			f, diags := ParseConfig([]byte(test.src), "", hcl.Pos{Line: 1, Column: 1})
			if len(diags) != 0 {
				for _, diag := range diags {
					t.Logf("- %s", diag.Error())
				}
				t.Fatalf("unexpected diagnostics")
			}

			attr := f.Body().GetAttribute(test.name)
			if attr == nil {
				if test.want != nil {
					t.Fatal("attribute not found, but want it to exist")
				}
			} else {
				if test.want == nil {
					t.Fatal("attribute found, but expecting not found")
				}

				got := attr.BuildTokens(nil)
				if !reflect.DeepEqual(got, test.want) {
					t.Errorf("wrong result\ngot:  %s\nwant: %s", spew.Sdump(got), spew.Sdump(test.want))
				}
			}
		})
	}
}
func TestBodySetAttributeValue(t *testing.T) {
	tests := []struct {
		src  string
		name string
		val  cty.Value
		want Tokens
	}{
		{
			"",
			"a",
			cty.True,
			Tokens{
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte{'a'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte{'='},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte("true"),
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEOF,
					Bytes:        []byte{},
					SpacesBefore: 0,
				},
			},
		},
		{
			"b = false\n",
			"a",
			cty.True,
			Tokens{
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte{'b'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte{'='},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte("false"),
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte{'a'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte{'='},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte("true"),
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEOF,
					Bytes:        []byte{},
					SpacesBefore: 0,
				},
			},
		},
		{
			"a = false\n",
			"a",
			cty.True,
			Tokens{
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte{'a'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte{'='},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte("true"),
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEOF,
					Bytes:        []byte{},
					SpacesBefore: 0,
				},
			},
		},
		{
			"a = 1\nb = false\n",
			"a",
			cty.True,
			Tokens{
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte{'a'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte{'='},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte("true"),
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte{'b'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte{'='},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte("false"),
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
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
		t.Run(fmt.Sprintf("%s = %#v in %s", test.name, test.val, test.src), func(t *testing.T) {
			f, diags := ParseConfig([]byte(test.src), "", hcl.Pos{Line: 1, Column: 1})
			if len(diags) != 0 {
				for _, diag := range diags {
					t.Logf("- %s", diag.Error())
				}
				t.Fatalf("unexpected diagnostics")
			}

			f.Body().SetAttributeValue(test.name, test.val)
			got := f.BuildTokens(nil)
			format(got)
			if !reflect.DeepEqual(got, test.want) {
				diff := cmp.Diff(test.want, got)
				t.Errorf("wrong result\ngot:  %s\nwant: %s\ndiff:\n%s", spew.Sdump(got), spew.Sdump(test.want), diff)
			}
		})
	}
}
