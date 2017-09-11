package json

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/hcl2/zcl"
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
			invalidVal{zcl.Range{
				Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
				End:   zcl.Pos{Line: 1, Column: 10, Byte: 9},
			}},
			1,
		},
		{
			`flase`,
			invalidVal{zcl.Range{
				Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
				End:   zcl.Pos{Line: 1, Column: 6, Byte: 5},
			}},
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
			invalidVal{zcl.Range{
				Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
				End:   zcl.Pos{Line: 1, Column: 7, Byte: 6},
			}},
			1,
		},
		{
			`"he\llo"`,
			invalidVal{zcl.Range{
				Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
				End:   zcl.Pos{Line: 1, Column: 9, Byte: 8},
			}},
			1,
		},
		{
			`1`,
			&numberVal{
				Value: mustBigFloat("1"),
				SrcRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 1, Column: 2, Byte: 1},
				},
			},
			0,
		},
		{
			`1.2`,
			&numberVal{
				Value: mustBigFloat("1.2"),
				SrcRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 1, Column: 4, Byte: 3},
				},
			},
			0,
		},
		{
			`-1`,
			&numberVal{
				Value: mustBigFloat("-1"),
				SrcRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 1, Column: 3, Byte: 2},
				},
			},
			0,
		},
		{
			`1.2e5`,
			&numberVal{
				Value: mustBigFloat("120000"),
				SrcRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 1, Column: 6, Byte: 5},
				},
			},
			0,
		},
		{
			`1.2e+5`,
			&numberVal{
				Value: mustBigFloat("120000"),
				SrcRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 1, Column: 7, Byte: 6},
				},
			},
			0,
		},
		{
			`1.2e-5`,
			&numberVal{
				Value: mustBigFloat("1.2e-5"),
				SrcRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 1, Column: 7, Byte: 6},
				},
			},
			0,
		},
		{
			`.1`,
			invalidVal{zcl.Range{
				Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
				End:   zcl.Pos{Line: 1, Column: 3, Byte: 2},
			}},
			1,
		},
		{
			`+2`,
			invalidVal{zcl.Range{
				Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
				End:   zcl.Pos{Line: 1, Column: 3, Byte: 2},
			}},
			1,
		},
		{
			`1 2`,
			&numberVal{
				Value: mustBigFloat("1"),
				SrcRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 1, Column: 2, Byte: 1},
				},
			},
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
				CloseRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 15, Byte: 14},
					End:   zcl.Pos{Line: 1, Column: 16, Byte: 15},
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
				CloseRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 29, Byte: 28},
					End:   zcl.Pos{Line: 1, Column: 30, Byte: 29},
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
				CloseRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 2, Byte: 1},
					End:   zcl.Pos{Line: 1, Column: 3, Byte: 2},
				},
			},
			0,
		},
		{
			`{"hello":true`,
			invalidVal{zcl.Range{
				Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
				End:   zcl.Pos{Line: 1, Column: 2, Byte: 1},
			}},
			1,
		},
		{
			`{"hello":true]`,
			invalidVal{zcl.Range{
				Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
				End:   zcl.Pos{Line: 1, Column: 2, Byte: 1},
			}},
			1,
		},
		{
			`{"hello":true,}`,
			invalidVal{zcl.Range{
				Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
				End:   zcl.Pos{Line: 1, Column: 2, Byte: 1},
			}},
			1,
		},
		{
			`{true:false}`,
			invalidVal{zcl.Range{
				Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
				End:   zcl.Pos{Line: 1, Column: 2, Byte: 1},
			}},
			1,
		},
		{
			`{"hello": true, "hello": true}`,
			&objectVal{
				Attrs: map[string]*objectAttr{
					"hello": {
						Name: "hello",
						Value: &booleanVal{
							Value: true,
							SrcRange: zcl.Range{
								Start: zcl.Pos{Line: 1, Column: 26, Byte: 25},
								End:   zcl.Pos{Line: 1, Column: 30, Byte: 29},
							},
						},
						NameRange: zcl.Range{
							Start: zcl.Pos{Line: 1, Column: 17, Byte: 16},
							End:   zcl.Pos{Line: 1, Column: 24, Byte: 23},
						},
					},
				},
				SrcRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 1, Column: 31, Byte: 30},
				},
				OpenRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 1, Column: 2, Byte: 1},
				},
				CloseRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 30, Byte: 29},
					End:   zcl.Pos{Line: 1, Column: 31, Byte: 30},
				},
			},
			1,
		},
		{
			`{"hello": true, "hello": true, "hello", true}`,
			invalidVal{zcl.Range{
				Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
				End:   zcl.Pos{Line: 1, Column: 2, Byte: 1},
			}},
			2,
		},
		{
			`{"hello", "world"}`,
			invalidVal{zcl.Range{
				Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
				End:   zcl.Pos{Line: 1, Column: 2, Byte: 1},
			}},
			1,
		},
		{
			`[]`,
			&arrayVal{
				Values: []node{},
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
			`[true]`,
			&arrayVal{
				Values: []node{
					&booleanVal{
						Value: true,
						SrcRange: zcl.Range{
							Start: zcl.Pos{Line: 1, Column: 2, Byte: 1},
							End:   zcl.Pos{Line: 1, Column: 6, Byte: 5},
						},
					},
				},
				SrcRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 1, Column: 7, Byte: 6},
				},
				OpenRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 1, Column: 2, Byte: 1},
				},
			},
			0,
		},
		{
			`[true, false]`,
			&arrayVal{
				Values: []node{
					&booleanVal{
						Value: true,
						SrcRange: zcl.Range{
							Start: zcl.Pos{Line: 1, Column: 2, Byte: 1},
							End:   zcl.Pos{Line: 1, Column: 6, Byte: 5},
						},
					},
					&booleanVal{
						Value: false,
						SrcRange: zcl.Range{
							Start: zcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:   zcl.Pos{Line: 1, Column: 13, Byte: 12},
						},
					},
				},
				SrcRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 1, Column: 14, Byte: 13},
				},
				OpenRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 1, Column: 2, Byte: 1},
				},
			},
			0,
		},
		{
			`[[]]`,
			&arrayVal{
				Values: []node{
					&arrayVal{
						Values: []node{},
						SrcRange: zcl.Range{
							Start: zcl.Pos{Line: 1, Column: 2, Byte: 1},
							End:   zcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
						OpenRange: zcl.Range{
							Start: zcl.Pos{Line: 1, Column: 2, Byte: 1},
							End:   zcl.Pos{Line: 1, Column: 3, Byte: 2},
						},
					},
				},
				SrcRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 1, Column: 5, Byte: 4},
				},
				OpenRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 1, Column: 2, Byte: 1},
				},
			},
			0,
		},
		{
			`[`,
			invalidVal{zcl.Range{
				Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
				End:   zcl.Pos{Line: 1, Column: 2, Byte: 1},
			}},
			2,
		},
		{
			`[true`,
			invalidVal{zcl.Range{
				Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
				End:   zcl.Pos{Line: 1, Column: 2, Byte: 1},
			}},
			1,
		},
		{
			`]`,
			invalidVal{zcl.Range{
				Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
				End:   zcl.Pos{Line: 1, Column: 2, Byte: 1},
			}},
			1,
		},
		{
			`[true,]`,
			invalidVal{zcl.Range{
				Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
				End:   zcl.Pos{Line: 1, Column: 2, Byte: 1},
			}},
			1,
		},
		{
			`[[],]`,
			invalidVal{zcl.Range{
				Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
				End:   zcl.Pos{Line: 1, Column: 2, Byte: 1},
			}},
			1,
		},
		{
			`["hello":true]`,
			invalidVal{zcl.Range{
				Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
				End:   zcl.Pos{Line: 1, Column: 2, Byte: 1},
			}},
			1,
		},
		{
			`[true}`,
			invalidVal{zcl.Range{
				Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
				End:   zcl.Pos{Line: 1, Column: 2, Byte: 1},
			}},
			1,
		},
		{
			`{"wrong"=true}`,
			invalidVal{zcl.Range{
				Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
				End:   zcl.Pos{Line: 1, Column: 2, Byte: 1},
			}},
			1,
		},
		{
			`{"wrong" = true}`,
			invalidVal{zcl.Range{
				Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
				End:   zcl.Pos{Line: 1, Column: 2, Byte: 1},
			}},
			1,
		},
		{
			`{"wrong" true}`,
			invalidVal{zcl.Range{
				Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
				End:   zcl.Pos{Line: 1, Column: 2, Byte: 1},
			}},
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

func mustBigFloat(s string) *big.Float {
	f, _, err := (&big.Float{}).Parse(s, 10)
	if err != nil {
		panic(err)
	}
	return f
}
