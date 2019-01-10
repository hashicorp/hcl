package hclpack

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/hcl2/hcl"
)

func TestBodyContent(t *testing.T) {
	tests := map[string]struct {
		Body   *Body
		Schema *hcl.BodySchema
		Want   *hcl.BodyContent
	}{
		"empty": {
			&Body{},
			&hcl.BodySchema{},
			&hcl.BodyContent{},
		},
		"nil": {
			nil,
			&hcl.BodySchema{},
			&hcl.BodyContent{},
		},
		"attribute": {
			&Body{
				Attributes: map[string]Attribute{
					"foo": {
						Expr: Expression{
							Source:     []byte(`"hello"`),
							SourceType: ExprNative,
						},
					},
				},
			},
			&hcl.BodySchema{
				Attributes: []hcl.AttributeSchema{
					{Name: "foo", Required: true},
					{Name: "bar", Required: false},
				},
			},
			&hcl.BodyContent{
				Attributes: hcl.Attributes{
					"foo": {
						Name: "foo",
						Expr: &Expression{
							Source:     []byte(`"hello"`),
							SourceType: ExprNative,
						},
					},
				},
			},
		},
		"block": {
			&Body{
				ChildBlocks: []Block{
					{
						Type: "foo",
					},
				},
			},
			&hcl.BodySchema{
				Blocks: []hcl.BlockHeaderSchema{
					{Type: "foo"},
				},
			},
			&hcl.BodyContent{
				Blocks: hcl.Blocks{
					{
						Type: "foo",
						Body: &Body{},
					},
				},
			},
		},
		"block attributes": {
			&Body{
				ChildBlocks: []Block{
					{
						Type: "foo",
						Body: Body{
							Attributes: map[string]Attribute{
								"bar": {
									Expr: Expression{
										Source:     []byte(`"hello"`),
										SourceType: ExprNative,
									},
								},
							},
						},
					},
					{
						Type: "foo",
						Body: Body{
							Attributes: map[string]Attribute{
								"bar": {
									Expr: Expression{
										Source:     []byte(`"world"`),
										SourceType: ExprNative,
									},
								},
							},
						},
					},
				},
			},
			&hcl.BodySchema{
				Blocks: []hcl.BlockHeaderSchema{
					{Type: "foo"},
				},
			},
			&hcl.BodyContent{
				Blocks: hcl.Blocks{
					{
						Type: "foo",
						Body: &Body{
							Attributes: map[string]Attribute{
								"bar": {
									Expr: Expression{
										Source:     []byte(`"hello"`),
										SourceType: ExprNative,
									},
								},
							},
						},
					},
					{
						Type: "foo",
						Body: &Body{
							Attributes: map[string]Attribute{
								"bar": {
									Expr: Expression{
										Source:     []byte(`"world"`),
										SourceType: ExprNative,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, diags := test.Body.Content(test.Schema)
			for _, diag := range diags {
				t.Errorf("unexpected diagnostic: %s", diag.Error())
			}

			if !cmp.Equal(test.Want, got) {
				bytesAsString := func(s []byte) string {
					return string(s)
				}
				t.Errorf("wrong result\n%s", cmp.Diff(
					test.Want, got,
					cmp.Transformer("bytesAsString", bytesAsString),
				))
			}
		})
	}

}
