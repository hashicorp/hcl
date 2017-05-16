package json

import (
	"reflect"
	"testing"

	"github.com/apparentlymart/go-zcl/zcl"
	"github.com/davecgh/go-spew/spew"
)

func TestParse(t *testing.T) {
	tests := []struct {
		Input     string
		Want      node
		DiagCount int
	}{
		// Simple, single-token constructs
		{
			`true`,
			&booleanVal{
				Value: true,
				SrcRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 1, Column: 5, Byte: 4},
				},
			},
			0,
		},
		{
			`false`,
			&booleanVal{
				Value: false,
				SrcRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 1, Column: 6, Byte: 5},
				},
			},
			0,
		},
		{
			`null`,
			&nullVal{
				SrcRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 1, Column: 5, Byte: 4},
				},
			},
			0,
		},
		{
			`undefined`,
			nil,
			1,
		},
		{
			`flase`,
			nil,
			1,
		},
		{
			`"hello"`,
			&stringVal{
				Value: "hello",
				SrcRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 1, Column: 8, Byte: 7},
				},
			},
			0,
		},
		{
			`"hello\nworld"`,
			&stringVal{
				Value: "hello\nworld",
				SrcRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 1, Column: 15, Byte: 14},
				},
			},
			0,
		},
		{
			`"hello \"world\""`,
			&stringVal{
				Value: `hello "world"`,
				SrcRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 1, Column: 18, Byte: 17},
				},
			},
			0,
		},
		{
			`"hello \\"`,
			&stringVal{
				Value: "hello \\",
				SrcRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 1, Column: 11, Byte: 10},
				},
			},
			0,
		},
		{
			`"hello`,
			nil,
			1,
		},
		{
			`"he\llo"`,
			nil,
			1,
		},

		// Objects
		{
			`{"hello": true}`,
			&objectVal{
				Attrs: map[string]*objectAttr{
					"hello": {
						Name: "hello",
						Value: &booleanVal{
							Value: true,
							SrcRange: zcl.Range{
								Start: zcl.Pos{Line: 1, Column: 11, Byte: 10},
								End:   zcl.Pos{Line: 1, Column: 15, Byte: 14},
							},
						},
						NameRange: zcl.Range{
							Start: zcl.Pos{Line: 1, Column: 2, Byte: 1},
							End:   zcl.Pos{Line: 1, Column: 9, Byte: 8},
						},
					},
				},
				SrcRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 1, Column: 16, Byte: 15},
				},
				OpenRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 1, Column: 2, Byte: 1},
				},
			},
			0,
		},
		{
			`{"hello": true, "bye": false}`,
			&objectVal{
				Attrs: map[string]*objectAttr{
					"hello": {
						Name: "hello",
						Value: &booleanVal{
							Value: true,
							SrcRange: zcl.Range{
								Start: zcl.Pos{Line: 1, Column: 11, Byte: 10},
								End:   zcl.Pos{Line: 1, Column: 15, Byte: 14},
							},
						},
						NameRange: zcl.Range{
							Start: zcl.Pos{Line: 1, Column: 2, Byte: 1},
							End:   zcl.Pos{Line: 1, Column: 9, Byte: 8},
						},
					},
					"bye": {
						Name: "bye",
						Value: &booleanVal{
							Value: false,
							SrcRange: zcl.Range{
								Start: zcl.Pos{Line: 1, Column: 24, Byte: 23},
								End:   zcl.Pos{Line: 1, Column: 29, Byte: 28},
							},
						},
						NameRange: zcl.Range{
							Start: zcl.Pos{Line: 1, Column: 17, Byte: 16},
							End:   zcl.Pos{Line: 1, Column: 22, Byte: 21},
						},
					},
				},
				SrcRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 1, Column: 30, Byte: 29},
				},
				OpenRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 1, Column: 2, Byte: 1},
				},
			},
			0,
		},
		{
			`{}`,
			&objectVal{
				Attrs: map[string]*objectAttr{},
				SrcRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 1, Column: 3, Byte: 2},
				},
				OpenRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 1, Column: 2, Byte: 1},
				},
			},
			0,
		},
		{
			`{"hello":true`,
			nil,
			1,
		},
		{
			`{"hello":true]`,
			nil,
			1,
		},
		{
			`{"hello":true,}`,
			nil,
			1,
		},
		{
			`{true:false}`,
			nil,
			1,
		},
		{
			`{"hello": true, "hello": true}`,
			nil,
			1,
		},
		{
			`{"hello": true, "hello": true, "hello", true}`,
			nil,
			2,
		},
		{
			`{"hello", "world"}`,
			nil,
			1,
		},
	}

	for _, test := range tests {
		t.Run(test.Input, func(t *testing.T) {
			got, diag := parseFileContent([]byte(test.Input), "")

			if len(diag) != test.DiagCount {
				t.Errorf("got %d diagnostics; want %d", len(diag), test.DiagCount)
				for _, d := range diag {
					t.Logf("  - %s", d.Error())
				}
			}

			if !reflect.DeepEqual(got, test.Want) {
				t.Errorf(
					"wrong result\ninput: %s\ngot:   %s\nwant:  %s",
					test.Input, spew.Sdump(got), spew.Sdump(test.Want),
				)
			}
		})
	}
}
