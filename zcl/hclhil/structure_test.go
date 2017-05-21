package hclhil

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/apparentlymart/go-zcl/zcl"
	"github.com/davecgh/go-spew/spew"
	hclast "github.com/hashicorp/hcl/hcl/ast"
	hcltoken "github.com/hashicorp/hcl/hcl/token"
)

func TestBodyJustAttributes(t *testing.T) {
	tests := []struct {
		Source    string
		Want      zcl.Attributes
		DiagCount int
	}{
		{
			``,
			zcl.Attributes{},
			0,
		},
		{
			`foo = "a"`,
			zcl.Attributes{
				"foo": &zcl.Attribute{
					Name: "foo",
					Expr: &expression{
						src: &hclast.LiteralType{
							Token: hcltoken.Token{
								Type: hcltoken.STRING,
								Pos: hcltoken.Pos{
									Offset: 6,
									Line:   1,
									Column: 7,
								},
								Text: `"a"`,
							},
						},
					},
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   zcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
					NameRange: zcl.Range{
						Start: zcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   zcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
			},
			0,
		},
		{
			`foo = {}`,
			zcl.Attributes{
				"foo": &zcl.Attribute{
					Name: "foo",
					Expr: &expression{
						src: &hclast.ObjectType{
							List: &hclast.ObjectList{},
							Lbrace: hcltoken.Pos{
								Offset: 6,
								Line:   1,
								Column: 7,
							},
							Rbrace: hcltoken.Pos{
								Offset: 7,
								Line:   1,
								Column: 8,
							},
						},
					},
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   zcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
					NameRange: zcl.Range{
						Start: zcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   zcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
			},
			0,
		},
		{
			`foo {}`,
			zcl.Attributes{
				"foo": &zcl.Attribute{
					Name: "foo",
					Expr: &expression{
						src: &hclast.ObjectType{
							List: &hclast.ObjectList{},
							Lbrace: hcltoken.Pos{
								Offset: 4,
								Line:   1,
								Column: 5,
							},
							Rbrace: hcltoken.Pos{
								Offset: 5,
								Line:   1,
								Column: 6,
							},
						},
					},
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   zcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
					NameRange: zcl.Range{
						Start: zcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   zcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
			},
			1, // warning about using block syntax
		},
		{
			`foo "bar" {}`,
			zcl.Attributes{},
			1, // blocks are not allowed here
		},
		{
			`
			    foo = 1
			    foo = 2
			`,
			zcl.Attributes{
				"foo": &zcl.Attribute{
					Name: "foo",
					Expr: &expression{
						src: &hclast.LiteralType{
							Token: hcltoken.Token{
								Type: hcltoken.NUMBER,
								Pos: hcltoken.Pos{
									Offset: 14,
									Line:   2,
									Column: 14,
								},
								Text: `1`,
							},
						},
					},
					Range: zcl.Range{
						Start: zcl.Pos{Byte: 8, Line: 2, Column: 8},
						End:   zcl.Pos{Byte: 9, Line: 2, Column: 9},
					},
					NameRange: zcl.Range{
						Start: zcl.Pos{Byte: 8, Line: 2, Column: 8},
						End:   zcl.Pos{Byte: 9, Line: 2, Column: 9},
					},
				},
			},
			1, // duplicate definition of foo
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			file, diags := Parse([]byte(test.Source), "test.hcl")
			if len(diags) != 0 {
				t.Fatalf("diagnostics from parse: %s", diags.Error())
			}

			got, diags := file.Body.JustAttributes()
			if len(diags) != test.DiagCount {
				t.Errorf("wrong number of diagnostics %d; want %d", len(diags), test.DiagCount)
				for _, diag := range diags {
					t.Logf(" - %s", diag.Error())
				}
			}

			if !reflect.DeepEqual(got, test.Want) {
				t.Errorf("wrong result\ngot:  %s\nwant: %s", spew.Sdump(got), spew.Sdump(test.Want))
			}
		})
	}
}
