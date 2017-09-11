package transform

import (
	"testing"

	"reflect"

	"github.com/hashicorp/hcl2/hcltest"
	"github.com/hashicorp/hcl2/zcl"
	"github.com/zclconf/go-cty/cty"
)

// Assert that deepWrapper implements Body
var deepWrapperIsBody zcl.Body = deepWrapper{}

func TestDeep(t *testing.T) {

	testTransform := TransformerFunc(func(body zcl.Body) zcl.Body {
		_, remain, diags := body.PartialContent(&zcl.BodySchema{
			Blocks: []zcl.BlockHeaderSchema{
				{
					Type: "remove",
				},
			},
		})

		return BodyWithDiagnostics(remain, diags)
	})

	src := hcltest.MockBody(&zcl.BodyContent{
		Attributes: hcltest.MockAttrs(map[string]zcl.Expression{
			"true": hcltest.MockExprLiteral(cty.True),
		}),
		Blocks: []*zcl.Block{
			{
				Type: "remove",
				Body: zcl.EmptyBody(),
			},
			{
				Type: "child",
				Body: hcltest.MockBody(&zcl.BodyContent{
					Blocks: []*zcl.Block{
						{
							Type: "remove",
						},
					},
				}),
			},
		},
	})

	wrapped := Deep(src, testTransform)

	rootContent, diags := wrapped.Content(&zcl.BodySchema{
		Attributes: []zcl.AttributeSchema{
			{
				Name: "true",
			},
		},
		Blocks: []zcl.BlockHeaderSchema{
			{
				Type: "child",
			},
		},
	})
	if len(diags) != 0 {
		t.Errorf("unexpected diagnostics for root content")
		for _, diag := range diags {
			t.Logf("- %s", diag)
		}
	}

	wantAttrs := hcltest.MockAttrs(map[string]zcl.Expression{
		"true": hcltest.MockExprLiteral(cty.True),
	})
	if !reflect.DeepEqual(rootContent.Attributes, wantAttrs) {
		t.Errorf("wrong root attributes\ngot:  %#v\nwant: %#v", rootContent.Attributes, wantAttrs)
	}

	if got, want := len(rootContent.Blocks), 1; got != want {
		t.Fatalf("wrong number of root blocks %d; want %d", got, want)
	}
	if got, want := rootContent.Blocks[0].Type, "child"; got != want {
		t.Errorf("wrong block type %s; want %s", got, want)
	}

	childBlock := rootContent.Blocks[0]
	childContent, diags := childBlock.Body.Content(&zcl.BodySchema{})
	if len(diags) != 0 {
		t.Errorf("unexpected diagnostics for child content")
		for _, diag := range diags {
			t.Logf("- %s", diag)
		}
	}

	if len(childContent.Attributes) != 0 {
		t.Errorf("unexpected attributes in child content; want empty content")
	}
	if len(childContent.Blocks) != 0 {
		t.Errorf("unexpected blocks in child content; want empty content")
	}
}
