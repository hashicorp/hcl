package zclsyntax

import (
	"reflect"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-zcl/zcl"
)

func TestParseConfig(t *testing.T) {
	tests := []struct {
		input     string
		diagCount int
		want      *Body
	}{
		{
			``,
			0,
			&Body{
				Attributes: Attributes{},
				Blocks:     Blocks{},
				SrcRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 1, Column: 1, Byte: 0},
				},
				EndRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 1, Column: 1, Byte: 0},
				},
			},
		},

		{
			"block {}\n",
			0,
			&Body{
				Attributes: Attributes{},
				Blocks: Blocks{
					&Block{
						Type:   "block",
						Labels: nil,
						Body: &Body{
							Attributes: Attributes{},
							Blocks:     Blocks{},

							SrcRange: zcl.Range{
								Start: zcl.Pos{Line: 1, Column: 7, Byte: 6},
								End:   zcl.Pos{Line: 1, Column: 9, Byte: 8},
							},
							EndRange: zcl.Range{
								Start: zcl.Pos{Line: 1, Column: 9, Byte: 8},
								End:   zcl.Pos{Line: 1, Column: 9, Byte: 8},
							},
						},

						TypeRange: zcl.Range{
							Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   zcl.Pos{Line: 1, Column: 6, Byte: 5},
						},
						LabelRanges: nil,
						OpenBraceRange: zcl.Range{
							Start: zcl.Pos{Line: 1, Column: 7, Byte: 6},
							End:   zcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
						CloseBraceRange: zcl.Range{
							Start: zcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:   zcl.Pos{Line: 1, Column: 9, Byte: 8},
						},
					},
				},
				SrcRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 2, Column: 1, Byte: 9},
				},
				EndRange: zcl.Range{
					Start: zcl.Pos{Line: 2, Column: 1, Byte: 9},
					End:   zcl.Pos{Line: 2, Column: 1, Byte: 9},
				},
			},
		},
		{
			"block {}block {}\n",
			1, // missing newline after block definition
			&Body{
				Attributes: Attributes{},
				Blocks: Blocks{
					&Block{
						Type:   "block",
						Labels: nil,
						Body: &Body{
							Attributes: Attributes{},
							Blocks:     Blocks{},

							SrcRange: zcl.Range{
								Start: zcl.Pos{Line: 1, Column: 7, Byte: 6},
								End:   zcl.Pos{Line: 1, Column: 9, Byte: 8},
							},
							EndRange: zcl.Range{
								Start: zcl.Pos{Line: 1, Column: 9, Byte: 8},
								End:   zcl.Pos{Line: 1, Column: 9, Byte: 8},
							},
						},

						TypeRange: zcl.Range{
							Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   zcl.Pos{Line: 1, Column: 6, Byte: 5},
						},
						LabelRanges: nil,
						OpenBraceRange: zcl.Range{
							Start: zcl.Pos{Line: 1, Column: 7, Byte: 6},
							End:   zcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
						CloseBraceRange: zcl.Range{
							Start: zcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:   zcl.Pos{Line: 1, Column: 9, Byte: 8},
						},
					},
				},
				SrcRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 2, Column: 1, Byte: 17},
				},
				EndRange: zcl.Range{
					Start: zcl.Pos{Line: 2, Column: 1, Byte: 17},
					End:   zcl.Pos{Line: 2, Column: 1, Byte: 17},
				},
			},
		},
		{
			"block \"foo\" {}\n",
			0,
			&Body{
				Attributes: Attributes{},
				Blocks: Blocks{
					&Block{
						Type:   "block",
						Labels: []string{"foo"},
						Body: &Body{
							Attributes: Attributes{},
							Blocks:     Blocks{},

							SrcRange: zcl.Range{
								Start: zcl.Pos{Line: 1, Column: 13, Byte: 12},
								End:   zcl.Pos{Line: 1, Column: 15, Byte: 14},
							},
							EndRange: zcl.Range{
								Start: zcl.Pos{Line: 1, Column: 15, Byte: 14},
								End:   zcl.Pos{Line: 1, Column: 15, Byte: 14},
							},
						},

						TypeRange: zcl.Range{
							Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   zcl.Pos{Line: 1, Column: 6, Byte: 5},
						},
						LabelRanges: []zcl.Range{
							{
								Start: zcl.Pos{Line: 1, Column: 7, Byte: 6},
								End:   zcl.Pos{Line: 1, Column: 12, Byte: 11},
							},
						},
						OpenBraceRange: zcl.Range{
							Start: zcl.Pos{Line: 1, Column: 13, Byte: 12},
							End:   zcl.Pos{Line: 1, Column: 14, Byte: 13},
						},
						CloseBraceRange: zcl.Range{
							Start: zcl.Pos{Line: 1, Column: 14, Byte: 13},
							End:   zcl.Pos{Line: 1, Column: 15, Byte: 14},
						},
					},
				},
				SrcRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 2, Column: 1, Byte: 15},
				},
				EndRange: zcl.Range{
					Start: zcl.Pos{Line: 2, Column: 1, Byte: 15},
					End:   zcl.Pos{Line: 2, Column: 1, Byte: 15},
				},
			},
		},
		{
			`
block "invalid" 1.2 {}
block "valid" {}
`,
			1,
			&Body{
				Attributes: Attributes{},
				Blocks: Blocks{
					&Block{
						Type:   "block",
						Labels: []string{"invalid"},
						Body:   nil,

						TypeRange: zcl.Range{
							Start: zcl.Pos{Line: 2, Column: 1, Byte: 1},
							End:   zcl.Pos{Line: 2, Column: 6, Byte: 6},
						},
						LabelRanges: []zcl.Range{
							{
								Start: zcl.Pos{Line: 2, Column: 7, Byte: 7},
								End:   zcl.Pos{Line: 2, Column: 16, Byte: 16},
							},
						},

						// Since we failed parsing before we got to the
						// braces, the type range is used as a placeholder
						// for these.
						OpenBraceRange: zcl.Range{
							Start: zcl.Pos{Line: 2, Column: 1, Byte: 1},
							End:   zcl.Pos{Line: 2, Column: 6, Byte: 6},
						},
						CloseBraceRange: zcl.Range{
							Start: zcl.Pos{Line: 2, Column: 1, Byte: 1},
							End:   zcl.Pos{Line: 2, Column: 6, Byte: 6},
						},
					},

					// Recovery behavior should allow us to still see this
					// second block, even though the first was invalid.
					&Block{
						Type:   "block",
						Labels: []string{"valid"},
						Body: &Body{
							Attributes: Attributes{},
							Blocks:     Blocks{},

							SrcRange: zcl.Range{
								Start: zcl.Pos{Line: 3, Column: 15, Byte: 38},
								End:   zcl.Pos{Line: 3, Column: 17, Byte: 40},
							},
							EndRange: zcl.Range{
								Start: zcl.Pos{Line: 3, Column: 17, Byte: 40},
								End:   zcl.Pos{Line: 3, Column: 17, Byte: 40},
							},
						},

						TypeRange: zcl.Range{
							Start: zcl.Pos{Line: 3, Column: 1, Byte: 24},
							End:   zcl.Pos{Line: 3, Column: 6, Byte: 29},
						},
						LabelRanges: []zcl.Range{
							{
								Start: zcl.Pos{Line: 3, Column: 7, Byte: 30},
								End:   zcl.Pos{Line: 3, Column: 14, Byte: 37},
							},
						},
						OpenBraceRange: zcl.Range{
							Start: zcl.Pos{Line: 3, Column: 15, Byte: 38},
							End:   zcl.Pos{Line: 3, Column: 16, Byte: 39},
						},
						CloseBraceRange: zcl.Range{
							Start: zcl.Pos{Line: 3, Column: 16, Byte: 39},
							End:   zcl.Pos{Line: 3, Column: 17, Byte: 40},
						},
					},
				},
				SrcRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 4, Column: 1, Byte: 41},
				},
				EndRange: zcl.Range{
					Start: zcl.Pos{Line: 4, Column: 1, Byte: 41},
					End:   zcl.Pos{Line: 4, Column: 1, Byte: 41},
				},
			},
		},
		{
			`block "f\o" {}
`,
			1, // "\o" is not a valid escape sequence
			&Body{
				Attributes: Attributes{},
				Blocks: Blocks{
					&Block{
						Type:   "block",
						Labels: []string{"fo"},
						Body:   nil,

						TypeRange: zcl.Range{
							Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   zcl.Pos{Line: 1, Column: 6, Byte: 5},
						},
						LabelRanges: []zcl.Range{
							{
								Start: zcl.Pos{Line: 1, Column: 7, Byte: 6},
								End:   zcl.Pos{Line: 1, Column: 12, Byte: 11},
							},
						},
						OpenBraceRange: zcl.Range{
							Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   zcl.Pos{Line: 1, Column: 6, Byte: 5},
						},
						CloseBraceRange: zcl.Range{
							Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   zcl.Pos{Line: 1, Column: 6, Byte: 5},
						},
					},
				},
				SrcRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 2, Column: 1, Byte: 15},
				},
				EndRange: zcl.Range{
					Start: zcl.Pos{Line: 2, Column: 1, Byte: 15},
					End:   zcl.Pos{Line: 2, Column: 1, Byte: 15},
				},
			},
		},
		{
			`block "f\n" {}
`,
			0,
			&Body{
				Attributes: Attributes{},
				Blocks: Blocks{
					&Block{
						Type:   "block",
						Labels: []string{"f\n"},
						Body: &Body{
							Attributes: Attributes{},
							Blocks:     Blocks{},

							SrcRange: zcl.Range{
								Start: zcl.Pos{Line: 1, Column: 13, Byte: 12},
								End:   zcl.Pos{Line: 1, Column: 15, Byte: 14},
							},
							EndRange: zcl.Range{
								Start: zcl.Pos{Line: 1, Column: 15, Byte: 14},
								End:   zcl.Pos{Line: 1, Column: 15, Byte: 14},
							},
						},

						TypeRange: zcl.Range{
							Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   zcl.Pos{Line: 1, Column: 6, Byte: 5},
						},
						LabelRanges: []zcl.Range{
							{
								Start: zcl.Pos{Line: 1, Column: 7, Byte: 6},
								End:   zcl.Pos{Line: 1, Column: 12, Byte: 11},
							},
						},
						OpenBraceRange: zcl.Range{
							Start: zcl.Pos{Line: 1, Column: 13, Byte: 12},
							End:   zcl.Pos{Line: 1, Column: 14, Byte: 13},
						},
						CloseBraceRange: zcl.Range{
							Start: zcl.Pos{Line: 1, Column: 14, Byte: 13},
							End:   zcl.Pos{Line: 1, Column: 15, Byte: 14},
						},
					},
				},
				SrcRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 2, Column: 1, Byte: 15},
				},
				EndRange: zcl.Range{
					Start: zcl.Pos{Line: 2, Column: 1, Byte: 15},
					End:   zcl.Pos{Line: 2, Column: 1, Byte: 15},
				},
			},
		},

		{
			"a = 1\n",
			0,
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &LiteralValueExpr{
							Val: cty.NumberIntVal(1),

							SrcRange: zcl.Range{
								Start: zcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   zcl.Pos{Line: 1, Column: 6, Byte: 5},
							},
						},

						SrcRange: zcl.Range{
							Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   zcl.Pos{Line: 1, Column: 6, Byte: 5},
						},
						NameRange: zcl.Range{
							Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   zcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: zcl.Range{
							Start: zcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   zcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 2, Column: 1, Byte: 6},
				},
				EndRange: zcl.Range{
					Start: zcl.Pos{Line: 2, Column: 1, Byte: 6},
					End:   zcl.Pos{Line: 2, Column: 1, Byte: 6},
				},
			},
		},
		{
			"a = foo.bar\n",
			0,
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &ScopeTraversalExpr{
							Traversal: zcl.Traversal{
								zcl.TraverseRoot{
									Name: "foo",

									SrcRange: zcl.Range{
										Start: zcl.Pos{Line: 1, Column: 5, Byte: 4},
										End:   zcl.Pos{Line: 1, Column: 8, Byte: 7},
									},
								},
								zcl.TraverseAttr{
									Name: "bar",

									SrcRange: zcl.Range{
										Start: zcl.Pos{Line: 1, Column: 8, Byte: 7},
										End:   zcl.Pos{Line: 1, Column: 12, Byte: 11},
									},
								},
							},

							SrcRange: zcl.Range{
								Start: zcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   zcl.Pos{Line: 1, Column: 12, Byte: 11},
							},
						},

						SrcRange: zcl.Range{
							Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   zcl.Pos{Line: 1, Column: 12, Byte: 11},
						},
						NameRange: zcl.Range{
							Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   zcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: zcl.Range{
							Start: zcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   zcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 2, Column: 1, Byte: 12},
				},
				EndRange: zcl.Range{
					Start: zcl.Pos{Line: 2, Column: 1, Byte: 12},
					End:   zcl.Pos{Line: 2, Column: 1, Byte: 12},
				},
			},
		},
		{
			"a = 1 # line comment\n",
			0,
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &LiteralValueExpr{
							Val: cty.NumberIntVal(1),

							SrcRange: zcl.Range{
								Start: zcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   zcl.Pos{Line: 1, Column: 6, Byte: 5},
							},
						},

						SrcRange: zcl.Range{
							Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   zcl.Pos{Line: 1, Column: 6, Byte: 5},
						},
						NameRange: zcl.Range{
							Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   zcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: zcl.Range{
							Start: zcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   zcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 2, Column: 1, Byte: 21},
				},
				EndRange: zcl.Range{
					Start: zcl.Pos{Line: 2, Column: 1, Byte: 21},
					End:   zcl.Pos{Line: 2, Column: 1, Byte: 21},
				},
			},
		},

		{
			`	`,
			2, // tabs not allowed, and body item is required here
			&Body{
				Attributes: Attributes{},
				Blocks:     Blocks{},
				SrcRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 1, Column: 2, Byte: 1},
				},
				EndRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 2, Byte: 1},
					End:   zcl.Pos{Line: 1, Column: 2, Byte: 1},
				},
			},
		},
		{
			`\x81`,
			2, // invalid UTF-8, and body item is required here
			&Body{
				Attributes: Attributes{},
				Blocks:     Blocks{},
				SrcRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 1, Column: 2, Byte: 1},
				},
				EndRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 2, Byte: 1},
					End:   zcl.Pos{Line: 1, Column: 2, Byte: 1},
				},
			},
		},
	}

	prettyConfig := &pretty.Config{
		Diffable:          true,
		IncludeUnexported: true,
		PrintStringers:    true,
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			file, diags := ParseConfig([]byte(test.input), "", zcl.Pos{Byte: 0, Line: 1, Column: 1})
			if len(diags) != test.diagCount {
				t.Errorf("wrong number of diagnostics %d; want %d", len(diags), test.diagCount)
				for _, diag := range diags {
					t.Logf(" - %s", diag.Error())
				}
			}

			got := file.Body

			if !reflect.DeepEqual(got, test.want) {
				diff := prettyConfig.Compare(test.want, got)
				t.Errorf("wrong result\ninput: %s\ndiff:  %s", test.input, diff)
			}
		})
	}
}
