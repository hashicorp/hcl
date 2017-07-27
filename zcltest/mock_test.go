package zcltest

import (
	"testing"

	"reflect"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-zcl/zcl"
)

var mockBodyIsBody zcl.Body = mockBody{}
var mockExprLiteralIsExpr zcl.Expression = mockExprLiteral{}
var mockExprVariableIsExpr zcl.Expression = mockExprVariable("")

func TestMockBodyPartialContent(t *testing.T) {
	tests := map[string]struct {
		In        *zcl.BodyContent
		Schema    *zcl.BodySchema
		Want      *zcl.BodyContent
		Remain    *zcl.BodyContent
		DiagCount int
	}{
		"empty": {
			&zcl.BodyContent{},
			&zcl.BodySchema{},
			&zcl.BodyContent{
				Attributes: zcl.Attributes{},
				Blocks:     zcl.Blocks{},
			},
			&zcl.BodyContent{
				Attributes: zcl.Attributes{},
				Blocks:     zcl.Blocks{},
			},
			0,
		},
		"attribute requested": {
			&zcl.BodyContent{
				Attributes: MockAttrs(map[string]zcl.Expression{
					"name": MockExprLiteral(cty.StringVal("Ermintrude")),
				}),
			},
			&zcl.BodySchema{
				Attributes: []zcl.AttributeSchema{
					{
						Name: "name",
					},
				},
			},
			&zcl.BodyContent{
				Attributes: MockAttrs(map[string]zcl.Expression{
					"name": MockExprLiteral(cty.StringVal("Ermintrude")),
				}),
				Blocks: zcl.Blocks{},
			},
			&zcl.BodyContent{
				Attributes: zcl.Attributes{},
				Blocks:     zcl.Blocks{},
			},
			0,
		},
		"attribute remains": {
			&zcl.BodyContent{
				Attributes: MockAttrs(map[string]zcl.Expression{
					"name": MockExprLiteral(cty.StringVal("Ermintrude")),
				}),
			},
			&zcl.BodySchema{},
			&zcl.BodyContent{
				Attributes: zcl.Attributes{},
				Blocks:     zcl.Blocks{},
			},
			&zcl.BodyContent{
				Attributes: MockAttrs(map[string]zcl.Expression{
					"name": MockExprLiteral(cty.StringVal("Ermintrude")),
				}),
				Blocks: zcl.Blocks{},
			},
			0,
		},
		"attribute missing": {
			&zcl.BodyContent{
				Attributes: zcl.Attributes{},
			},
			&zcl.BodySchema{
				Attributes: []zcl.AttributeSchema{
					{
						Name:     "name",
						Required: true,
					},
				},
			},
			&zcl.BodyContent{
				Attributes: zcl.Attributes{},
				Blocks:     zcl.Blocks{},
			},
			&zcl.BodyContent{
				Attributes: zcl.Attributes{},
				Blocks:     zcl.Blocks{},
			},
			1, // missing attribute "name"
		},
		"block requested, no labels": {
			&zcl.BodyContent{
				Blocks: zcl.Blocks{
					{
						Type: "baz",
					},
				},
			},
			&zcl.BodySchema{
				Blocks: []zcl.BlockHeaderSchema{
					{
						Type: "baz",
					},
				},
			},
			&zcl.BodyContent{
				Attributes: zcl.Attributes{},
				Blocks: zcl.Blocks{
					{
						Type: "baz",
					},
				},
			},
			&zcl.BodyContent{
				Attributes: zcl.Attributes{},
				Blocks:     zcl.Blocks{},
			},
			0,
		},
		"block requested, wrong labels": {
			&zcl.BodyContent{
				Blocks: zcl.Blocks{
					{
						Type: "baz",
					},
				},
			},
			&zcl.BodySchema{
				Blocks: []zcl.BlockHeaderSchema{
					{
						Type:       "baz",
						LabelNames: []string{"foo"},
					},
				},
			},
			&zcl.BodyContent{
				Attributes: zcl.Attributes{},
				Blocks: zcl.Blocks{
					{
						Type: "baz",
					},
				},
			},
			&zcl.BodyContent{
				Attributes: zcl.Attributes{},
				Blocks:     zcl.Blocks{},
			},
			1, // "baz" requires 1 label
		},
		"block remains": {
			&zcl.BodyContent{
				Blocks: zcl.Blocks{
					{
						Type: "baz",
					},
				},
			},
			&zcl.BodySchema{},
			&zcl.BodyContent{
				Attributes: zcl.Attributes{},
				Blocks:     zcl.Blocks{},
			},
			&zcl.BodyContent{
				Attributes: zcl.Attributes{},
				Blocks: zcl.Blocks{
					{
						Type: "baz",
					},
				},
			},
			0,
		},
		"various": {
			&zcl.BodyContent{
				Attributes: MockAttrs(map[string]zcl.Expression{
					"name": MockExprLiteral(cty.StringVal("Ermintrude")),
					"age":  MockExprLiteral(cty.NumberIntVal(32)),
				}),
				Blocks: zcl.Blocks{
					{
						Type: "baz",
					},
					{
						Type:   "bar",
						Labels: []string{"foo1"},
					},
					{
						Type:   "bar",
						Labels: []string{"foo2"},
					},
				},
			},
			&zcl.BodySchema{
				Attributes: []zcl.AttributeSchema{
					{
						Name: "name",
					},
				},
				Blocks: []zcl.BlockHeaderSchema{
					{
						Type:       "bar",
						LabelNames: []string{"name"},
					},
				},
			},
			&zcl.BodyContent{
				Attributes: MockAttrs(map[string]zcl.Expression{
					"name": MockExprLiteral(cty.StringVal("Ermintrude")),
				}),
				Blocks: zcl.Blocks{
					{
						Type:   "bar",
						Labels: []string{"foo1"},
					},
					{
						Type:   "bar",
						Labels: []string{"foo2"},
					},
				},
			},
			&zcl.BodyContent{
				Attributes: MockAttrs(map[string]zcl.Expression{
					"age": MockExprLiteral(cty.NumberIntVal(32)),
				}),
				Blocks: zcl.Blocks{
					{
						Type: "baz",
					},
				},
			},
			0,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			inBody := MockBody(test.In)
			got, remainBody, diags := inBody.PartialContent(test.Schema)
			if len(diags) != test.DiagCount {
				t.Errorf("wrong number of diagnostics %d; want %d", len(diags), test.DiagCount)
				for _, diag := range diags {
					t.Logf("- %s", diag)
				}
			}

			if !reflect.DeepEqual(got, test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}

			gotRemain := remainBody.(mockBody).C
			if !reflect.DeepEqual(gotRemain, test.Remain) {
				t.Errorf("wrong remain\ngot:  %#v\nwant: %#v", gotRemain, test.Remain)
			}
		})
	}
}
