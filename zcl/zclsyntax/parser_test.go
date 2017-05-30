package zclsyntax

import (
	"reflect"
	"testing"

	"github.com/kylelemons/godebug/pretty"
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
			`block {}`,
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
					End:   zcl.Pos{Line: 1, Column: 9, Byte: 8},
				},
				EndRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 9, Byte: 8},
					End:   zcl.Pos{Line: 1, Column: 9, Byte: 8},
				},
			},
		},

		{
			`block "foo" {}`,
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
					End:   zcl.Pos{Line: 1, Column: 15, Byte: 14},
				},
				EndRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 15, Byte: 14},
					End:   zcl.Pos{Line: 1, Column: 15, Byte: 14},
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
