package include

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcltest"
	"github.com/hashicorp/hcl2/zcl"
	"github.com/zclconf/go-cty/cty"
)

func TestTransformer(t *testing.T) {
	caller := hcltest.MockBody(&zcl.BodyContent{
		Blocks: zcl.Blocks{
			{
				Type: "include",
				Body: hcltest.MockBody(&zcl.BodyContent{
					Attributes: hcltest.MockAttrs(map[string]zcl.Expression{
						"path": hcltest.MockExprVariable("var_path"),
					}),
				}),
			},
			{
				Type: "include",
				Body: hcltest.MockBody(&zcl.BodyContent{
					Attributes: hcltest.MockAttrs(map[string]zcl.Expression{
						"path": hcltest.MockExprLiteral(cty.StringVal("include2")),
					}),
				}),
			},
			{
				Type: "foo",
				Body: hcltest.MockBody(&zcl.BodyContent{
					Attributes: hcltest.MockAttrs(map[string]zcl.Expression{
						"from": hcltest.MockExprLiteral(cty.StringVal("caller")),
					}),
				}),
			},
		},
	})

	resolver := MapResolver(map[string]zcl.Body{
		"include1": hcltest.MockBody(&zcl.BodyContent{
			Blocks: zcl.Blocks{
				{
					Type: "foo",
					Body: hcltest.MockBody(&zcl.BodyContent{
						Attributes: hcltest.MockAttrs(map[string]zcl.Expression{
							"from": hcltest.MockExprLiteral(cty.StringVal("include1")),
						}),
					}),
				},
			},
		}),
		"include2": hcltest.MockBody(&zcl.BodyContent{
			Blocks: zcl.Blocks{
				{
					Type: "foo",
					Body: hcltest.MockBody(&zcl.BodyContent{
						Attributes: hcltest.MockAttrs(map[string]zcl.Expression{
							"from": hcltest.MockExprLiteral(cty.StringVal("include2")),
						}),
					}),
				},
			},
		}),
	})

	ctx := &zcl.EvalContext{
		Variables: map[string]cty.Value{
			"var_path": cty.StringVal("include1"),
		},
	}

	transformer := Transformer("include", ctx, resolver)
	merged := transformer.TransformBody(caller)

	type foo struct {
		From string `zcl:"from,attr"`
	}
	type result struct {
		Foos []foo `zcl:"foo,block"`
	}
	var got result
	diags := gohcl.DecodeBody(merged, nil, &got)
	if len(diags) != 0 {
		t.Errorf("unexpected diags")
		for _, diag := range diags {
			t.Logf("- %s", diag)
		}
	}

	want := result{
		Foos: []foo{
			{
				From: "caller",
			},
			{
				From: "include1",
			},
			{
				From: "include2",
			},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("wrong result\ngot: %swant: %s", spew.Sdump(got), spew.Sdump(want))
	}
}
