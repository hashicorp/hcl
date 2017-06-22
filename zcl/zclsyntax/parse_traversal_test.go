package zclsyntax

import (
	"testing"

	"reflect"

	"github.com/davecgh/go-spew/spew"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-zcl/zcl"
)

func TestParseTraversalAbs(t *testing.T) {
	tests := []struct {
		src       string
		want      zcl.Traversal
		diagCount int
	}{
		{
			"",
			nil,
			1, // variable name required
		},
		{
			"foo",
			zcl.Traversal{
				zcl.TraverseRoot{
					Name: "foo",
					SrcRange: zcl.Range{
						Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:   zcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
				},
			},
			0,
		},
		{
			"foo.bar.baz",
			zcl.Traversal{
				zcl.TraverseRoot{
					Name: "foo",
					SrcRange: zcl.Range{
						Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:   zcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
				},
				zcl.TraverseAttr{
					Name: "bar",
					SrcRange: zcl.Range{
						Start: zcl.Pos{Line: 1, Column: 4, Byte: 3},
						End:   zcl.Pos{Line: 1, Column: 8, Byte: 7},
					},
				},
				zcl.TraverseAttr{
					Name: "baz",
					SrcRange: zcl.Range{
						Start: zcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:   zcl.Pos{Line: 1, Column: 12, Byte: 11},
					},
				},
			},
			0,
		},
		{
			"foo[1]",
			zcl.Traversal{
				zcl.TraverseRoot{
					Name: "foo",
					SrcRange: zcl.Range{
						Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:   zcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
				},
				zcl.TraverseIndex{
					Key: cty.NumberIntVal(1),
					SrcRange: zcl.Range{
						Start: zcl.Pos{Line: 1, Column: 4, Byte: 3},
						End:   zcl.Pos{Line: 1, Column: 7, Byte: 6},
					},
				},
			},
			0,
		},
		{
			"foo[1][2]",
			zcl.Traversal{
				zcl.TraverseRoot{
					Name: "foo",
					SrcRange: zcl.Range{
						Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:   zcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
				},
				zcl.TraverseIndex{
					Key: cty.NumberIntVal(1),
					SrcRange: zcl.Range{
						Start: zcl.Pos{Line: 1, Column: 4, Byte: 3},
						End:   zcl.Pos{Line: 1, Column: 7, Byte: 6},
					},
				},
				zcl.TraverseIndex{
					Key: cty.NumberIntVal(2),
					SrcRange: zcl.Range{
						Start: zcl.Pos{Line: 1, Column: 7, Byte: 6},
						End:   zcl.Pos{Line: 1, Column: 10, Byte: 9},
					},
				},
			},
			0,
		},
		{
			"foo[1].bar",
			zcl.Traversal{
				zcl.TraverseRoot{
					Name: "foo",
					SrcRange: zcl.Range{
						Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:   zcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
				},
				zcl.TraverseIndex{
					Key: cty.NumberIntVal(1),
					SrcRange: zcl.Range{
						Start: zcl.Pos{Line: 1, Column: 4, Byte: 3},
						End:   zcl.Pos{Line: 1, Column: 7, Byte: 6},
					},
				},
				zcl.TraverseAttr{
					Name: "bar",
					SrcRange: zcl.Range{
						Start: zcl.Pos{Line: 1, Column: 7, Byte: 6},
						End:   zcl.Pos{Line: 1, Column: 11, Byte: 10},
					},
				},
			},
			0,
		},
		{
			"foo.",
			zcl.Traversal{
				zcl.TraverseRoot{
					Name: "foo",
					SrcRange: zcl.Range{
						Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:   zcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
				},
			},
			1, // attribute name required
		},
		{
			"foo[",
			zcl.Traversal{
				zcl.TraverseRoot{
					Name: "foo",
					SrcRange: zcl.Range{
						Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:   zcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
				},
			},
			1, // index required
		},
		{
			"foo[index]",
			zcl.Traversal{
				zcl.TraverseRoot{
					Name: "foo",
					SrcRange: zcl.Range{
						Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:   zcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
				},
			},
			1, // index must be literal
		},
		{
			"foo[0",
			zcl.Traversal{
				zcl.TraverseRoot{
					Name: "foo",
					SrcRange: zcl.Range{
						Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:   zcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
				},
				zcl.TraverseIndex{
					Key: cty.NumberIntVal(0),
					SrcRange: zcl.Range{
						Start: zcl.Pos{Line: 1, Column: 4, Byte: 3},
						End:   zcl.Pos{Line: 1, Column: 6, Byte: 5},
					},
				},
			},
			1, // missing close bracket
		},
		{
			"foo 0",
			zcl.Traversal{
				zcl.TraverseRoot{
					Name: "foo",
					SrcRange: zcl.Range{
						Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:   zcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
				},
			},
			1, // extra junk after traversal
		},
	}

	for _, test := range tests {
		t.Run(test.src, func(t *testing.T) {
			got, diags := ParseTraversalAbs([]byte(test.src), "", zcl.Pos{Line: 1, Column: 1})
			if len(diags) != test.diagCount {
				for _, diag := range diags {
					t.Logf(" - %s", diag.Error())
				}
				t.Errorf("wrong number of diagnostics %d; want %d", len(diags), test.diagCount)
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("wrong result\nsrc:  %s\ngot:  %s\nwant: %s", test.src, spew.Sdump(got), spew.Sdump(test.want))
			}
		})
	}
}
