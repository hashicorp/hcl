package hclsyntax

import (
	"testing"

	"github.com/go-test/deep"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/zclconf/go-cty/cty"
)

func init() {
	deep.MaxDepth = 999
}

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
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 1, Column: 1, Byte: 0},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 1, Column: 1, Byte: 0},
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

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 7, Byte: 6},
								End:   hcl.Pos{Line: 1, Column: 9, Byte: 8},
							},
							EndRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 9, Byte: 8},
								End:   hcl.Pos{Line: 1, Column: 9, Byte: 8},
							},
						},

						TypeRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
						},
						LabelRanges: nil,
						OpenBraceRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 7, Byte: 6},
							End:   hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
						CloseBraceRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:   hcl.Pos{Line: 1, Column: 9, Byte: 8},
						},
					},
				},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 9},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 9},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 9},
				},
			},
		},
		{
			"block {}",
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

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 7, Byte: 6},
								End:   hcl.Pos{Line: 1, Column: 9, Byte: 8},
							},
							EndRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 9, Byte: 8},
								End:   hcl.Pos{Line: 1, Column: 9, Byte: 8},
							},
						},

						TypeRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
						},
						LabelRanges: nil,
						OpenBraceRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 7, Byte: 6},
							End:   hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
						CloseBraceRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:   hcl.Pos{Line: 1, Column: 9, Byte: 8},
						},
					},
				},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 1, Column: 9, Byte: 8},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 9, Byte: 8},
					End:   hcl.Pos{Line: 1, Column: 9, Byte: 8},
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

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 7, Byte: 6},
								End:   hcl.Pos{Line: 1, Column: 9, Byte: 8},
							},
							EndRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 9, Byte: 8},
								End:   hcl.Pos{Line: 1, Column: 9, Byte: 8},
							},
						},

						TypeRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
						},
						LabelRanges: nil,
						OpenBraceRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 7, Byte: 6},
							End:   hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
						CloseBraceRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:   hcl.Pos{Line: 1, Column: 9, Byte: 8},
						},
					},
				},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 17},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 17},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 17},
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

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 13, Byte: 12},
								End:   hcl.Pos{Line: 1, Column: 15, Byte: 14},
							},
							EndRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 15, Byte: 14},
								End:   hcl.Pos{Line: 1, Column: 15, Byte: 14},
							},
						},

						TypeRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
						},
						LabelRanges: []hcl.Range{
							{
								Start: hcl.Pos{Line: 1, Column: 7, Byte: 6},
								End:   hcl.Pos{Line: 1, Column: 12, Byte: 11},
							},
						},
						OpenBraceRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 13, Byte: 12},
							End:   hcl.Pos{Line: 1, Column: 14, Byte: 13},
						},
						CloseBraceRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 14, Byte: 13},
							End:   hcl.Pos{Line: 1, Column: 15, Byte: 14},
						},
					},
				},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 15},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 15},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 15},
				},
			},
		},
		{
			"block foo {}\n",
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

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 11, Byte: 10},
								End:   hcl.Pos{Line: 1, Column: 13, Byte: 12},
							},
							EndRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 13, Byte: 12},
								End:   hcl.Pos{Line: 1, Column: 13, Byte: 12},
							},
						},

						TypeRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
						},
						LabelRanges: []hcl.Range{
							{
								Start: hcl.Pos{Line: 1, Column: 7, Byte: 6},
								End:   hcl.Pos{Line: 1, Column: 10, Byte: 9},
							},
						},
						OpenBraceRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 11, Byte: 10},
							End:   hcl.Pos{Line: 1, Column: 12, Byte: 11},
						},
						CloseBraceRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 12, Byte: 11},
							End:   hcl.Pos{Line: 1, Column: 13, Byte: 12},
						},
					},
				},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 13},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 13},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 13},
				},
			},
		},
		{
			"block \"invalid ${not_allowed_here} foo\" {}\n",
			1, // Invalid string literal; Template sequences are not allowed in this string.
			&Body{
				Attributes: Attributes{},
				Blocks: Blocks{
					&Block{
						Type:   "block",
						Labels: []string{"invalid ${ ... } foo"}, // invalid interpolation gets replaced with a placeholder here
						Body: &Body{
							Attributes: Attributes{},
							Blocks:     Blocks{},

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 41, Byte: 40},
								End:   hcl.Pos{Line: 1, Column: 43, Byte: 42},
							},
							EndRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 43, Byte: 42},
								End:   hcl.Pos{Line: 1, Column: 43, Byte: 42},
							},
						},

						TypeRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
						},
						LabelRanges: []hcl.Range{
							{
								Start: hcl.Pos{Line: 1, Column: 7, Byte: 6},
								End:   hcl.Pos{Line: 1, Column: 40, Byte: 39},
							},
						},
						OpenBraceRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 41, Byte: 40},
							End:   hcl.Pos{Line: 1, Column: 42, Byte: 41},
						},
						CloseBraceRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 42, Byte: 41},
							End:   hcl.Pos{Line: 1, Column: 43, Byte: 42},
						},
					},
				},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 43},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 43},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 43},
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
						Body: &Body{
							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 2, Column: 1, Byte: 1},
								End:   hcl.Pos{Line: 2, Column: 6, Byte: 6},
							},
							EndRange: hcl.Range{
								Start: hcl.Pos{Line: 2, Column: 1, Byte: 1},
								End:   hcl.Pos{Line: 2, Column: 6, Byte: 6},
							},
						},

						TypeRange: hcl.Range{
							Start: hcl.Pos{Line: 2, Column: 1, Byte: 1},
							End:   hcl.Pos{Line: 2, Column: 6, Byte: 6},
						},
						LabelRanges: []hcl.Range{
							{
								Start: hcl.Pos{Line: 2, Column: 7, Byte: 7},
								End:   hcl.Pos{Line: 2, Column: 16, Byte: 16},
							},
						},

						// Since we failed parsing before we got to the
						// braces, the type range is used as a placeholder
						// for these.
						OpenBraceRange: hcl.Range{
							Start: hcl.Pos{Line: 2, Column: 1, Byte: 1},
							End:   hcl.Pos{Line: 2, Column: 6, Byte: 6},
						},
						CloseBraceRange: hcl.Range{
							Start: hcl.Pos{Line: 2, Column: 1, Byte: 1},
							End:   hcl.Pos{Line: 2, Column: 6, Byte: 6},
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

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 3, Column: 15, Byte: 38},
								End:   hcl.Pos{Line: 3, Column: 17, Byte: 40},
							},
							EndRange: hcl.Range{
								Start: hcl.Pos{Line: 3, Column: 17, Byte: 40},
								End:   hcl.Pos{Line: 3, Column: 17, Byte: 40},
							},
						},

						TypeRange: hcl.Range{
							Start: hcl.Pos{Line: 3, Column: 1, Byte: 24},
							End:   hcl.Pos{Line: 3, Column: 6, Byte: 29},
						},
						LabelRanges: []hcl.Range{
							{
								Start: hcl.Pos{Line: 3, Column: 7, Byte: 30},
								End:   hcl.Pos{Line: 3, Column: 14, Byte: 37},
							},
						},
						OpenBraceRange: hcl.Range{
							Start: hcl.Pos{Line: 3, Column: 15, Byte: 38},
							End:   hcl.Pos{Line: 3, Column: 16, Byte: 39},
						},
						CloseBraceRange: hcl.Range{
							Start: hcl.Pos{Line: 3, Column: 16, Byte: 39},
							End:   hcl.Pos{Line: 3, Column: 17, Byte: 40},
						},
					},
				},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 4, Column: 1, Byte: 41},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 4, Column: 1, Byte: 41},
					End:   hcl.Pos{Line: 4, Column: 1, Byte: 41},
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
						Body: &Body{
							Attributes: map[string]*Attribute{},
							Blocks:     []*Block{},

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 13, Byte: 12},
								End:   hcl.Pos{Line: 1, Column: 15, Byte: 14},
							},
							EndRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 15, Byte: 14},
								End:   hcl.Pos{Line: 1, Column: 15, Byte: 14},
							},
						},

						TypeRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
						},
						LabelRanges: []hcl.Range{
							{
								Start: hcl.Pos{Line: 1, Column: 7, Byte: 6},
								End:   hcl.Pos{Line: 1, Column: 12, Byte: 11},
							},
						},
						OpenBraceRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 13, Byte: 12},
							End:   hcl.Pos{Line: 1, Column: 14, Byte: 13},
						},
						CloseBraceRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 14, Byte: 13},
							End:   hcl.Pos{Line: 1, Column: 15, Byte: 14},
						},
					},
				},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 15},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 15},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 15},
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

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 13, Byte: 12},
								End:   hcl.Pos{Line: 1, Column: 15, Byte: 14},
							},
							EndRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 15, Byte: 14},
								End:   hcl.Pos{Line: 1, Column: 15, Byte: 14},
							},
						},

						TypeRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
						},
						LabelRanges: []hcl.Range{
							{
								Start: hcl.Pos{Line: 1, Column: 7, Byte: 6},
								End:   hcl.Pos{Line: 1, Column: 12, Byte: 11},
							},
						},
						OpenBraceRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 13, Byte: 12},
							End:   hcl.Pos{Line: 1, Column: 14, Byte: 13},
						},
						CloseBraceRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 14, Byte: 13},
							End:   hcl.Pos{Line: 1, Column: 15, Byte: 14},
						},
					},
				},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 15},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 15},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 15},
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

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
							},
						},

						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 6},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 6},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 6},
				},
			},
		},
		{
			"a = 1",
			0,
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &LiteralValueExpr{
							Val: cty.NumberIntVal(1),

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
							},
						},

						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 6, Byte: 5},
					End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
				},
			},
		},
		{
			"a = \"hello ${true}\"\n",
			0,
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &TemplateExpr{
							Parts: []Expression{
								&LiteralValueExpr{
									Val: cty.StringVal("hello "),

									SrcRange: hcl.Range{
										Start: hcl.Pos{Line: 1, Column: 6, Byte: 5},
										End:   hcl.Pos{Line: 1, Column: 12, Byte: 11},
									},
								},
								&LiteralValueExpr{
									Val: cty.True,

									SrcRange: hcl.Range{
										Start: hcl.Pos{Line: 1, Column: 14, Byte: 13},
										End:   hcl.Pos{Line: 1, Column: 18, Byte: 17},
									},
								},
							},

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 20, Byte: 19},
							},
						},

						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 20, Byte: 19},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 20},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 20},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 20},
				},
			},
		},
		{
			"a = \"hello $${true}\"\n",
			0,
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &TemplateExpr{
							Parts: []Expression{
								&LiteralValueExpr{
									Val: cty.StringVal("hello ${true}"),

									SrcRange: hcl.Range{
										Start: hcl.Pos{Line: 1, Column: 6, Byte: 5},
										End:   hcl.Pos{Line: 1, Column: 20, Byte: 19},
									},
								},
							},

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 21, Byte: 20},
							},
						},

						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 21, Byte: 20},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 21},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 21},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 21},
				},
			},
		},
		{
			"a = \"hello %%{true}\"\n",
			0,
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &TemplateExpr{
							Parts: []Expression{
								&LiteralValueExpr{
									Val: cty.StringVal("hello %{true}"),

									SrcRange: hcl.Range{
										Start: hcl.Pos{Line: 1, Column: 6, Byte: 5},
										End:   hcl.Pos{Line: 1, Column: 20, Byte: 19},
									},
								},
							},

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 21, Byte: 20},
							},
						},

						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 21, Byte: 20},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 21},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 21},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 21},
				},
			},
		},
		{
			"a = \"hello $$\"\n",
			0,
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &TemplateExpr{
							Parts: []Expression{
								&LiteralValueExpr{
									Val: cty.StringVal("hello $"),

									SrcRange: hcl.Range{
										Start: hcl.Pos{Line: 1, Column: 6, Byte: 5},
										End:   hcl.Pos{Line: 1, Column: 13, Byte: 12},
									},
								},
								// This parses oddly due to how the scanner
								// handles escaping of the $ sequence, but it's
								// functionally equivalent to a single literal.
								&LiteralValueExpr{
									Val: cty.StringVal("$"),

									SrcRange: hcl.Range{
										Start: hcl.Pos{Line: 1, Column: 13, Byte: 12},
										End:   hcl.Pos{Line: 1, Column: 14, Byte: 13},
									},
								},
							},

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 15, Byte: 14},
							},
						},

						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 15, Byte: 14},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 15},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 15},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 15},
				},
			},
		},
		{
			"a = \"hello $\"\n",
			0, // unterminated template interpolation sequence
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &TemplateExpr{
							Parts: []Expression{
								&LiteralValueExpr{
									Val: cty.StringVal("hello $"),

									SrcRange: hcl.Range{
										Start: hcl.Pos{Line: 1, Column: 6, Byte: 5},
										End:   hcl.Pos{Line: 1, Column: 13, Byte: 12},
									},
								},
							},

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 14, Byte: 13},
							},
						},

						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 14, Byte: 13},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 14},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 14},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 14},
				},
			},
		},
		{
			"a = \"hello %%\"\n",
			0,
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &TemplateExpr{
							Parts: []Expression{
								&LiteralValueExpr{
									Val: cty.StringVal("hello %"),

									SrcRange: hcl.Range{
										Start: hcl.Pos{Line: 1, Column: 6, Byte: 5},
										End:   hcl.Pos{Line: 1, Column: 13, Byte: 12},
									},
								},
								// This parses oddly due to how the scanner
								// handles escaping of the $ sequence, but it's
								// functionally equivalent to a single literal.
								&LiteralValueExpr{
									Val: cty.StringVal("%"),

									SrcRange: hcl.Range{
										Start: hcl.Pos{Line: 1, Column: 13, Byte: 12},
										End:   hcl.Pos{Line: 1, Column: 14, Byte: 13},
									},
								},
							},

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 15, Byte: 14},
							},
						},

						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 15, Byte: 14},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 15},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 15},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 15},
				},
			},
		},
		{
			"a = \"hello %\"\n",
			0, // unterminated template control sequence
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &TemplateExpr{
							Parts: []Expression{
								&LiteralValueExpr{
									Val: cty.StringVal("hello %"),

									SrcRange: hcl.Range{
										Start: hcl.Pos{Line: 1, Column: 6, Byte: 5},
										End:   hcl.Pos{Line: 1, Column: 13, Byte: 12},
									},
								},
							},

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 14, Byte: 13},
							},
						},

						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 14, Byte: 13},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 14},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 14},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 14},
				},
			},
		},
		{
			"a = \"hello!\"\n",
			0,
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &TemplateExpr{
							Parts: []Expression{
								&LiteralValueExpr{
									Val: cty.StringVal("hello!"),

									SrcRange: hcl.Range{
										Start: hcl.Pos{Line: 1, Column: 6, Byte: 5},
										End:   hcl.Pos{Line: 1, Column: 12, Byte: 11},
									},
								},
							},

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 13, Byte: 12},
							},
						},

						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 13, Byte: 12},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 13},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 13},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 13},
				},
			},
		},
		{
			"a = \"\\u2022\"\n",
			0,
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &TemplateExpr{
							Parts: []Expression{
								&LiteralValueExpr{
									Val: cty.StringVal("\u2022"),

									SrcRange: hcl.Range{
										Start: hcl.Pos{Line: 1, Column: 6, Byte: 5},
										End:   hcl.Pos{Line: 1, Column: 12, Byte: 11},
									},
								},
							},

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 13, Byte: 12},
							},
						},

						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 13, Byte: 12},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 13},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 13},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 13},
				},
			},
		},
		{
			"a = \"\\uu2022\"\n",
			1, // \u must be followed by four hex digits
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &TemplateExpr{
							Parts: []Expression{
								&LiteralValueExpr{
									Val: cty.StringVal("\\uu2022"),

									SrcRange: hcl.Range{
										Start: hcl.Pos{Line: 1, Column: 6, Byte: 5},
										End:   hcl.Pos{Line: 1, Column: 13, Byte: 12},
									},
								},
							},

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 14, Byte: 13},
							},
						},

						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 14, Byte: 13},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 14},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 14},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 14},
				},
			},
		},
		{
			"a = \"\\U0001d11e\"\n",
			0,
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &TemplateExpr{
							Parts: []Expression{
								&LiteralValueExpr{
									Val: cty.StringVal("\U0001d11e"),

									SrcRange: hcl.Range{
										Start: hcl.Pos{Line: 1, Column: 6, Byte: 5},
										End:   hcl.Pos{Line: 1, Column: 16, Byte: 15},
									},
								},
							},

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 17, Byte: 16},
							},
						},

						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 17, Byte: 16},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 17},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 17},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 17},
				},
			},
		},
		{
			"a = \"\\u0001d11e\"\n",
			0, // This is valid, but probably not what the user intended :(
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &TemplateExpr{
							Parts: []Expression{
								&LiteralValueExpr{
									// Only the first four digits were used for the
									// escape sequence, so the remaining four just
									// get echoed out literally.
									Val: cty.StringVal("\u0001d11e"),

									SrcRange: hcl.Range{
										Start: hcl.Pos{Line: 1, Column: 6, Byte: 5},
										End:   hcl.Pos{Line: 1, Column: 16, Byte: 15},
									},
								},
							},

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 17, Byte: 16},
							},
						},

						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 17, Byte: 16},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 17},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 17},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 17},
				},
			},
		},
		{
			"a = \"\\U2022\"\n",
			1, // Invalid escape sequence, since we need eight hex digits for \U
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &TemplateExpr{
							Parts: []Expression{
								&LiteralValueExpr{
									Val: cty.StringVal("\\U2022"),

									SrcRange: hcl.Range{
										Start: hcl.Pos{Line: 1, Column: 6, Byte: 5},
										End:   hcl.Pos{Line: 1, Column: 12, Byte: 11},
									},
								},
							},

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 13, Byte: 12},
							},
						},

						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 13, Byte: 12},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 13},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 13},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 13},
				},
			},
		},
		{
			"a = \"\\u20m2\"\n",
			1, // Invalid escape sequence
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &TemplateExpr{
							Parts: []Expression{
								&LiteralValueExpr{
									Val: cty.StringVal("\\u20m2"),

									SrcRange: hcl.Range{
										Start: hcl.Pos{Line: 1, Column: 6, Byte: 5},
										End:   hcl.Pos{Line: 1, Column: 12, Byte: 11},
									},
								},
							},

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 13, Byte: 12},
							},
						},

						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 13, Byte: 12},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 13},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 13},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 13},
				},
			},
		},
		{
			"a = \"\\U00300000\"\n",
			1, // Invalid unicode character (can't encode in UTF-8)
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &TemplateExpr{
							Parts: []Expression{
								&LiteralValueExpr{
									Val: cty.StringVal("\\U00300000"),

									SrcRange: hcl.Range{
										Start: hcl.Pos{Line: 1, Column: 6, Byte: 5},
										End:   hcl.Pos{Line: 1, Column: 16, Byte: 15},
									},
								},
							},

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 17, Byte: 16},
							},
						},

						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 17, Byte: 16},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 17},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 17},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 17},
				},
			},
		},
		{
			"a = \"\\Ub2705550\"\n",
			1, // Invalid unicode character (can't encode in UTF-8)
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &TemplateExpr{
							Parts: []Expression{
								&LiteralValueExpr{
									Val: cty.StringVal("\\Ub2705550"),

									SrcRange: hcl.Range{
										Start: hcl.Pos{Line: 1, Column: 6, Byte: 5},
										End:   hcl.Pos{Line: 1, Column: 16, Byte: 15},
									},
								},
							},

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 17, Byte: 16},
							},
						},

						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 17, Byte: 16},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 17},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 17},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 17},
				},
			},
		},
		{
			"a = <<EOT\nHello\nEOT\nb = \"Hi\"",
			0,
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &TemplateExpr{
							Parts: []Expression{
								&LiteralValueExpr{
									Val: cty.StringVal("Hello\n"),

									SrcRange: hcl.Range{
										Start: hcl.Pos{Line: 2, Column: 1, Byte: 10},
										End:   hcl.Pos{Line: 3, Column: 1, Byte: 16},
									},
								},
							},

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 3, Column: 4, Byte: 19},
							},
						},

						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 3, Column: 4, Byte: 19},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
					"b": {
						Name: "b",
						Expr: &TemplateExpr{
							Parts: []Expression{
								&LiteralValueExpr{
									Val: cty.StringVal("Hi"),

									SrcRange: hcl.Range{
										Start: hcl.Pos{Line: 4, Column: 6, Byte: 25},
										End:   hcl.Pos{Line: 4, Column: 8, Byte: 27},
									},
								},
							},

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 4, Column: 5, Byte: 24},
								End:   hcl.Pos{Line: 4, Column: 9, Byte: 28},
							},
						},

						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 4, Column: 1, Byte: 20},
							End:   hcl.Pos{Line: 4, Column: 9, Byte: 28},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 4, Column: 1, Byte: 20},
							End:   hcl.Pos{Line: 4, Column: 2, Byte: 21},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 4, Column: 3, Byte: 22},
							End:   hcl.Pos{Line: 4, Column: 4, Byte: 23},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 4, Column: 9, Byte: 28},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 4, Column: 9, Byte: 28},
					End:   hcl.Pos{Line: 4, Column: 9, Byte: 28},
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
							Traversal: hcl.Traversal{
								hcl.TraverseRoot{
									Name: "foo",

									SrcRange: hcl.Range{
										Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
										End:   hcl.Pos{Line: 1, Column: 8, Byte: 7},
									},
								},
								hcl.TraverseAttr{
									Name: "bar",

									SrcRange: hcl.Range{
										Start: hcl.Pos{Line: 1, Column: 8, Byte: 7},
										End:   hcl.Pos{Line: 1, Column: 12, Byte: 11},
									},
								},
							},

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 12, Byte: 11},
							},
						},

						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 12, Byte: 11},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 12},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 12},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 12},
				},
			},
		},
		{
			"a = foo.0.1.baz\n",
			1, // Chaining legacy index syntax is not supported
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &ScopeTraversalExpr{
							Traversal: hcl.Traversal{
								hcl.TraverseRoot{
									Name: "foo",

									SrcRange: hcl.Range{
										Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
										End:   hcl.Pos{Line: 1, Column: 8, Byte: 7},
									},
								},
								hcl.TraverseIndex{
									Key: cty.DynamicVal,

									SrcRange: hcl.Range{
										Start: hcl.Pos{Line: 1, Column: 8, Byte: 7},
										End:   hcl.Pos{Line: 1, Column: 12, Byte: 11},
									},
								},
								hcl.TraverseAttr{
									Name: "baz",

									SrcRange: hcl.Range{
										Start: hcl.Pos{Line: 1, Column: 12, Byte: 11},
										End:   hcl.Pos{Line: 1, Column: 16, Byte: 15},
									},
								},
							},

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 16, Byte: 15},
							},
						},

						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 16, Byte: 15},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 16},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 16},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 16},
				},
			},
		},
		{
			"a = \"${var.public_subnets[count.index]}\"\n",
			0,
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &TemplateWrapExpr{
							Wrapped: &IndexExpr{
								Collection: &ScopeTraversalExpr{
									Traversal: hcl.Traversal{
										hcl.TraverseRoot{
											Name: "var",

											SrcRange: hcl.Range{
												Start: hcl.Pos{Line: 1, Column: 8, Byte: 7},
												End:   hcl.Pos{Line: 1, Column: 11, Byte: 10},
											},
										},
										hcl.TraverseAttr{
											Name: "public_subnets",

											SrcRange: hcl.Range{
												Start: hcl.Pos{Line: 1, Column: 11, Byte: 10},
												End:   hcl.Pos{Line: 1, Column: 26, Byte: 25},
											},
										},
									},

									SrcRange: hcl.Range{
										Start: hcl.Pos{Line: 1, Column: 8, Byte: 7},
										End:   hcl.Pos{Line: 1, Column: 26, Byte: 25},
									},
								},
								Key: &ScopeTraversalExpr{
									Traversal: hcl.Traversal{
										hcl.TraverseRoot{
											Name: "count",

											SrcRange: hcl.Range{
												Start: hcl.Pos{Line: 1, Column: 27, Byte: 26},
												End:   hcl.Pos{Line: 1, Column: 32, Byte: 31},
											},
										},
										hcl.TraverseAttr{
											Name: "index",

											SrcRange: hcl.Range{
												Start: hcl.Pos{Line: 1, Column: 32, Byte: 31},
												End:   hcl.Pos{Line: 1, Column: 38, Byte: 37},
											},
										},
									},

									SrcRange: hcl.Range{
										Start: hcl.Pos{Line: 1, Column: 27, Byte: 26},
										End:   hcl.Pos{Line: 1, Column: 38, Byte: 37},
									},
								},
								SrcRange: hcl.Range{
									Start: hcl.Pos{Line: 1, Column: 26, Byte: 25},
									End:   hcl.Pos{Line: 1, Column: 39, Byte: 38},
								},
								OpenRange: hcl.Range{
									Start: hcl.Pos{Line: 1, Column: 26, Byte: 25},
									End:   hcl.Pos{Line: 1, Column: 27, Byte: 26},
								},
							},
							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 41, Byte: 40},
							},
						},
						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 41, Byte: 40},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 41},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 41},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 41},
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

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
							},
						},

						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 21},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 21},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 21},
				},
			},
		},

		{
			"a = [for k, v in foo: v if true]\n",
			0,
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &ForExpr{
							KeyVar: "k",
							ValVar: "v",

							CollExpr: &ScopeTraversalExpr{
								Traversal: hcl.Traversal{
									hcl.TraverseRoot{
										Name: "foo",
										SrcRange: hcl.Range{
											Start: hcl.Pos{Line: 1, Column: 18, Byte: 17},
											End:   hcl.Pos{Line: 1, Column: 21, Byte: 20},
										},
									},
								},
								SrcRange: hcl.Range{
									Start: hcl.Pos{Line: 1, Column: 18, Byte: 17},
									End:   hcl.Pos{Line: 1, Column: 21, Byte: 20},
								},
							},
							ValExpr: &ScopeTraversalExpr{
								Traversal: hcl.Traversal{
									hcl.TraverseRoot{
										Name: "v",
										SrcRange: hcl.Range{
											Start: hcl.Pos{Line: 1, Column: 23, Byte: 22},
											End:   hcl.Pos{Line: 1, Column: 24, Byte: 23},
										},
									},
								},
								SrcRange: hcl.Range{
									Start: hcl.Pos{Line: 1, Column: 23, Byte: 22},
									End:   hcl.Pos{Line: 1, Column: 24, Byte: 23},
								},
							},
							CondExpr: &LiteralValueExpr{
								Val: cty.True,
								SrcRange: hcl.Range{
									Start: hcl.Pos{Line: 1, Column: 28, Byte: 27},
									End:   hcl.Pos{Line: 1, Column: 32, Byte: 31},
								},
							},

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 33, Byte: 32},
							},
							OpenRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
							},
							CloseRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 32, Byte: 31},
								End:   hcl.Pos{Line: 1, Column: 33, Byte: 32},
							},
						},

						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 33, Byte: 32},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 33},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 33},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 33},
				},
			},
		},
		{
			"a = [for k, v in foo: k => v... if true]\n",
			2, // can't use => or ... in a tuple for
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &ForExpr{
							KeyVar: "k",
							ValVar: "v",

							CollExpr: &ScopeTraversalExpr{
								Traversal: hcl.Traversal{
									hcl.TraverseRoot{
										Name: "foo",
										SrcRange: hcl.Range{
											Start: hcl.Pos{Line: 1, Column: 18, Byte: 17},
											End:   hcl.Pos{Line: 1, Column: 21, Byte: 20},
										},
									},
								},
								SrcRange: hcl.Range{
									Start: hcl.Pos{Line: 1, Column: 18, Byte: 17},
									End:   hcl.Pos{Line: 1, Column: 21, Byte: 20},
								},
							},
							KeyExpr: &ScopeTraversalExpr{
								Traversal: hcl.Traversal{
									hcl.TraverseRoot{
										Name: "k",
										SrcRange: hcl.Range{
											Start: hcl.Pos{Line: 1, Column: 23, Byte: 22},
											End:   hcl.Pos{Line: 1, Column: 24, Byte: 23},
										},
									},
								},
								SrcRange: hcl.Range{
									Start: hcl.Pos{Line: 1, Column: 23, Byte: 22},
									End:   hcl.Pos{Line: 1, Column: 24, Byte: 23},
								},
							},
							ValExpr: &ScopeTraversalExpr{
								Traversal: hcl.Traversal{
									hcl.TraverseRoot{
										Name: "v",
										SrcRange: hcl.Range{
											Start: hcl.Pos{Line: 1, Column: 28, Byte: 27},
											End:   hcl.Pos{Line: 1, Column: 29, Byte: 28},
										},
									},
								},
								SrcRange: hcl.Range{
									Start: hcl.Pos{Line: 1, Column: 28, Byte: 27},
									End:   hcl.Pos{Line: 1, Column: 29, Byte: 28},
								},
							},
							CondExpr: &LiteralValueExpr{
								Val: cty.True,
								SrcRange: hcl.Range{
									Start: hcl.Pos{Line: 1, Column: 36, Byte: 35},
									End:   hcl.Pos{Line: 1, Column: 40, Byte: 39},
								},
							},
							Group: true,

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 41, Byte: 40},
							},
							OpenRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
							},
							CloseRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 40, Byte: 39},
								End:   hcl.Pos{Line: 1, Column: 41, Byte: 40},
							},
						},

						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 41, Byte: 40},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 41},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 41},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 41},
				},
			},
		},

		{
			`	`,
			0, // the tab character is treated as a single whitespace character
			&Body{
				Attributes: Attributes{},
				Blocks:     Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 2, Byte: 1},
					End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 2, Byte: 1},
					End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
				},
			},
		},
		{
			`\x81`,
			2, // invalid UTF-8, and body item is required here
			&Body{
				Attributes: Attributes{},
				Blocks:     Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 2, Byte: 1},
					End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
				},
			},
		},
		{
			"a = 1,",
			1,
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &LiteralValueExpr{
							Val: cty.NumberIntVal(1),

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
							},
						},

						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 1, Column: 7, Byte: 6},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 7, Byte: 6},
					End:   hcl.Pos{Line: 1, Column: 7, Byte: 6},
				},
			},
		},
		{
			"a = `str`",
			2, // Invalid character and expression
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &LiteralValueExpr{
							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
							},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 1, Column: 10, Byte: 9},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 10, Byte: 9},
					End:   hcl.Pos{Line: 1, Column: 10, Byte: 9},
				},
			},
		},
		{
			`a = 'str'`,
			2, // Invalid character and expression
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &LiteralValueExpr{
							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
							},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 1, Column: 10, Byte: 9},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 10, Byte: 9},
					End:   hcl.Pos{Line: 1, Column: 10, Byte: 9},
				},
			},
		},
		{
			"a = sort(data.first.ref.attr)[count.index]\n",
			0,
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &IndexExpr{
							Collection: &FunctionCallExpr{
								Name: "sort",
								Args: []Expression{
									&ScopeTraversalExpr{
										Traversal: hcl.Traversal{
											hcl.TraverseRoot{
												Name: "data",
												SrcRange: hcl.Range{
													Filename: "",
													Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
													End:      hcl.Pos{Line: 1, Column: 14, Byte: 13},
												},
											},
											hcl.TraverseAttr{
												Name: "first",
												SrcRange: hcl.Range{
													Filename: "",
													Start:    hcl.Pos{Line: 1, Column: 14, Byte: 13},
													End:      hcl.Pos{Line: 1, Column: 20, Byte: 19},
												},
											},
											hcl.TraverseAttr{
												Name: "ref",
												SrcRange: hcl.Range{
													Filename: "",
													Start:    hcl.Pos{Line: 1, Column: 20, Byte: 19},
													End:      hcl.Pos{Line: 1, Column: 24, Byte: 23},
												},
											},
											hcl.TraverseAttr{
												Name: "attr",
												SrcRange: hcl.Range{
													Filename: "",
													Start:    hcl.Pos{Line: 1, Column: 24, Byte: 23},
													End:      hcl.Pos{Line: 1, Column: 29, Byte: 28},
												},
											},
										},
										SrcRange: hcl.Range{
											Filename: "",
											Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
											End:      hcl.Pos{Line: 1, Column: 29, Byte: 28},
										},
									},
								},
								ExpandFinal: false,
								NameRange: hcl.Range{
									Filename: "",
									Start:    hcl.Pos{Line: 1, Column: 5, Byte: 4},
									End:      hcl.Pos{Line: 1, Column: 9, Byte: 8},
								},
								OpenParenRange: hcl.Range{
									Filename: "",
									Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
									End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
								},
								CloseParenRange: hcl.Range{
									Filename: "",
									Start:    hcl.Pos{Line: 1, Column: 29, Byte: 28},
									End:      hcl.Pos{Line: 1, Column: 30, Byte: 29},
								},
							},
							Key: &ScopeTraversalExpr{
								Traversal: hcl.Traversal{
									hcl.TraverseRoot{
										Name: "count",
										SrcRange: hcl.Range{
											Filename: "",
											Start:    hcl.Pos{Line: 1, Column: 31, Byte: 30},
											End:      hcl.Pos{Line: 1, Column: 36, Byte: 35},
										},
									},
									hcl.TraverseAttr{
										Name: "index",
										SrcRange: hcl.Range{
											Filename: "",
											Start:    hcl.Pos{Line: 1, Column: 36, Byte: 35},
											End:      hcl.Pos{Line: 1, Column: 42, Byte: 41},
										},
									},
								},
								SrcRange: hcl.Range{
									Filename: "",
									Start:    hcl.Pos{Line: 1, Column: 31, Byte: 30},
									End:      hcl.Pos{Line: 1, Column: 42, Byte: 41},
								},
							},
							SrcRange: hcl.Range{
								Filename: "",
								Start:    hcl.Pos{Line: 1, Column: 30, Byte: 29},
								End:      hcl.Pos{Line: 1, Column: 43, Byte: 42},
							},
							OpenRange: hcl.Range{
								Filename: "",
								Start:    hcl.Pos{Line: 1, Column: 30, Byte: 29},
								End:      hcl.Pos{Line: 1, Column: 31, Byte: 30},
							},
						},
						SrcRange: hcl.Range{
							Filename: "",
							Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:      hcl.Pos{Line: 1, Column: 43, Byte: 42},
						},
						NameRange: hcl.Range{
							Filename: "",
							Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:      hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Filename: "",
							Start:    hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:      hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 43},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 43},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 43},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			file, diags := ParseConfig([]byte(test.input), "", hcl.Pos{Byte: 0, Line: 1, Column: 1})
			if len(diags) != test.diagCount {
				t.Errorf("wrong number of diagnostics %d; want %d", len(diags), test.diagCount)
				for _, diag := range diags {
					t.Logf(" - %s", diag.Error())
				}
			}

			got := file.Body

			if diff := deep.Equal(got, test.want); diff != nil {
				for _, problem := range diff {
					t.Errorf(problem)
				}
			}
		})
	}
}
