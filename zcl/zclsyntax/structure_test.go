package zclsyntax

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-zcl/zcl"
)

func TestBodyJustAttributes(t *testing.T) {
	tests := []struct {
		body      *Body
		want      zcl.Attributes
		diagCount int
	}{
		{
			&Body{},
			zcl.Attributes{},
			0,
		},
		{
			&Body{
				Attributes: Attributes{},
			},
			zcl.Attributes{},
			0,
		},
		{
			&Body{
				Attributes: Attributes{
					"foo": &Attribute{
						Name: "foo",
						Expr: &LiteralValueExpr{
							Val: cty.StringVal("bar"),
						},
					},
				},
			},
			zcl.Attributes{
				"foo": &zcl.Attribute{
					Name: "foo",
					Expr: &LiteralValueExpr{
						Val: cty.StringVal("bar"),
					},
				},
			},
			0,
		},
		{
			&Body{
				Attributes: Attributes{
					"foo": &Attribute{
						Name: "foo",
						Expr: &LiteralValueExpr{
							Val: cty.StringVal("bar"),
						},
					},
				},
				Blocks: Blocks{
					{
						Type: "foo",
					},
				},
			},
			zcl.Attributes{
				"foo": &zcl.Attribute{
					Name: "foo",
					Expr: &LiteralValueExpr{
						Val: cty.StringVal("bar"),
					},
				},
			},
			1, // blocks are not allowed here
		},
		{
			&Body{
				Attributes: Attributes{
					"foo": &Attribute{
						Name: "foo",
						Expr: &LiteralValueExpr{
							Val: cty.StringVal("bar"),
						},
					},
				},
				hiddenAttrs: map[string]struct{}{
					"foo": struct{}{},
				},
			},
			zcl.Attributes{},
			0,
		},
	}

	prettyConfig := &pretty.Config{
		Diffable:          true,
		IncludeUnexported: true,
		PrintStringers:    true,
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			got, diags := test.body.JustAttributes()

			if len(diags) != test.diagCount {
				t.Errorf("wrong number of diagnostics %d; want %d", len(diags), test.diagCount)
				for _, diag := range diags {
					t.Logf(" - %s", diag.Error())
				}
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf(
					"wrong result\nbody: %s\ndiff: %s",
					prettyConfig.Sprint(test.body),
					prettyConfig.Compare(test.want, got),
				)
			}
		})
	}
}
