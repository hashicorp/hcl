package hclpack

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/hcl2/hcl"
)

func TestPackNativeFile(t *testing.T) {
	src := `
foo = "bar"
baz = "boz"

child {
  a = b + c
}

another_child "foo" "bar" {}
`

	got, diags := PackNativeFile([]byte(src), "", hcl.Pos{Line: 1, Column: 1})
	for _, diag := range diags {
		t.Errorf("unexpected diagnostic: %s", diag.Error())
	}

	want := &Body{
		Attributes: map[string]Attribute{
			"baz": {
				Expr: Expression{
					Source:     []byte(`"boz"`),
					SourceType: ExprNative,
					Range_: hcl.Range{
						Start: hcl.Pos{Line: 3, Column: 7, Byte: 19},
						End:   hcl.Pos{Line: 3, Column: 12, Byte: 24},
					},
					StartRange_: hcl.Range{
						Start: hcl.Pos{Line: 3, Column: 8, Byte: 20},
						End:   hcl.Pos{Line: 3, Column: 11, Byte: 23},
					},
				},
				Range: hcl.Range{
					Start: hcl.Pos{Line: 3, Column: 1, Byte: 13},
					End:   hcl.Pos{Line: 3, Column: 12, Byte: 24},
				},
				NameRange: hcl.Range{
					Start: hcl.Pos{Line: 3, Column: 1, Byte: 13},
					End:   hcl.Pos{Line: 3, Column: 4, Byte: 16},
				},
			},
			"foo": {
				Expr: Expression{
					Source:     []byte(`"bar"`),
					SourceType: ExprNative,
					Range_: hcl.Range{
						Start: hcl.Pos{Line: 2, Column: 7, Byte: 7},
						End:   hcl.Pos{Line: 2, Column: 12, Byte: 12},
					},
					StartRange_: hcl.Range{
						Start: hcl.Pos{Line: 2, Column: 8, Byte: 8},
						End:   hcl.Pos{Line: 2, Column: 11, Byte: 11},
					},
				},
				Range: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 1},
					End:   hcl.Pos{Line: 2, Column: 12, Byte: 12},
				},
				NameRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 1},
					End:   hcl.Pos{Line: 2, Column: 4, Byte: 4},
				},
			},
		},
		ChildBlocks: []Block{
			{
				Type: "child",
				Body: Body{
					Attributes: map[string]Attribute{
						"a": {
							Expr: Expression{
								Source:     []byte(`b + c`),
								SourceType: ExprNative,
								Range_: hcl.Range{
									Start: hcl.Pos{Line: 6, Column: 7, Byte: 40},
									End:   hcl.Pos{Line: 6, Column: 12, Byte: 45},
								},
								StartRange_: hcl.Range{
									Start: hcl.Pos{Line: 6, Column: 7, Byte: 40},
									End:   hcl.Pos{Line: 6, Column: 8, Byte: 41},
								},
							},
							Range: hcl.Range{
								Start: hcl.Pos{Line: 6, Column: 3, Byte: 36},
								End:   hcl.Pos{Line: 6, Column: 12, Byte: 45},
							},
							NameRange: hcl.Range{
								Start: hcl.Pos{Line: 6, Column: 3, Byte: 36},
								End:   hcl.Pos{Line: 6, Column: 4, Byte: 37},
							},
						},
					},
					MissingItemRange_: hcl.Range{
						Start: hcl.Pos{Line: 7, Column: 2, Byte: 47},
						End:   hcl.Pos{Line: 7, Column: 2, Byte: 47},
					},
				},
				DefRange: hcl.Range{
					Start: hcl.Pos{Line: 5, Column: 1, Byte: 26},
					End:   hcl.Pos{Line: 5, Column: 6, Byte: 31},
				},
				TypeRange: hcl.Range{
					Start: hcl.Pos{Line: 5, Column: 1, Byte: 26},
					End:   hcl.Pos{Line: 5, Column: 6, Byte: 31},
				},
			},
			{
				Type:   "another_child",
				Labels: []string{"foo", "bar"},
				Body: Body{
					MissingItemRange_: hcl.Range{
						Start: hcl.Pos{Line: 9, Column: 29, Byte: 77},
						End:   hcl.Pos{Line: 9, Column: 29, Byte: 77},
					},
				},
				DefRange: hcl.Range{
					Start: hcl.Pos{Line: 9, Column: 1, Byte: 49},
					End:   hcl.Pos{Line: 9, Column: 26, Byte: 74},
				},
				TypeRange: hcl.Range{
					Start: hcl.Pos{Line: 9, Column: 1, Byte: 49},
					End:   hcl.Pos{Line: 9, Column: 14, Byte: 62},
				},
				LabelRanges: []hcl.Range{
					hcl.Range{
						Start: hcl.Pos{Line: 9, Column: 15, Byte: 63},
						End:   hcl.Pos{Line: 9, Column: 20, Byte: 68},
					},
					hcl.Range{
						Start: hcl.Pos{Line: 9, Column: 21, Byte: 69},
						End:   hcl.Pos{Line: 9, Column: 26, Byte: 74},
					},
				},
			},
		},
		MissingItemRange_: hcl.Range{
			Start: hcl.Pos{Line: 10, Column: 1, Byte: 78},
			End:   hcl.Pos{Line: 10, Column: 1, Byte: 78},
		},
	}

	if !cmp.Equal(want, got) {
		bytesAsString := func(s []byte) string {
			return string(s)
		}
		posAsString := func(pos hcl.Pos) string {
			return fmt.Sprintf("%#v", pos)
		}
		t.Errorf("wrong result\n%s", cmp.Diff(
			want, got,
			cmp.Transformer("bytesAsString", bytesAsString),
			cmp.Transformer("posAsString", posAsString),
		))
	}
}
