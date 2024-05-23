package hclwrite

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func TestAttributeSetName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		src     string
		oldName string
		newName string
		want    Tokens
	}{
		{
			"old = 123",
			"old",
			"new",
			Tokens{
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte(`new`),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte{'='},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNumberLit,
					Bytes:        []byte(`123`),
					SpacesBefore: 1,
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
		test := test

		t.Run(fmt.Sprintf("%s %s in %s", test.oldName, test.newName, test.src), func(t *testing.T) {
			t.Parallel()

			f, diags := ParseConfig([]byte(test.src), "", hcl.Pos{Line: 1, Column: 1})

			if len(diags) != 0 {
				for _, diag := range diags {
					t.Logf("- %s", diag.Error())
				}
				t.Fatalf("unexpected diagnostics")
			}

			attr := f.Body().GetAttribute(test.oldName)
			attr.SetName(test.newName)
			got := f.BuildTokens(nil)
			format(got)

			if !reflect.DeepEqual(got, test.want) {
				diff := cmp.Diff(test.want, got)
				t.Errorf("wrong result\ngot:  %s\nwant: %s\ndiff:\n%s", spew.Sdump(got), spew.Sdump(test.want), diff)
			}
		})
	}
}
