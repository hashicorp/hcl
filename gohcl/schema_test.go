package gohcl

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/hcl2/zcl"
)

func TestImpliedBodySchema(t *testing.T) {
	tests := []struct {
		val         interface{}
		wantSchema  *zcl.BodySchema
		wantPartial bool
	}{
		{
			struct{}{},
			&zcl.BodySchema{},
			false,
		},
		{
			struct {
				Ignored bool
			}{},
			&zcl.BodySchema{},
			false,
		},
		{
			struct {
				Attr1 bool `zcl:"attr1"`
				Attr2 bool `zcl:"attr2"`
			}{},
			&zcl.BodySchema{
				Attributes: []zcl.AttributeSchema{
					{
						Name:     "attr1",
						Required: true,
					},
					{
						Name:     "attr2",
						Required: true,
					},
				},
			},
			false,
		},
		{
			struct {
				Attr *bool `zcl:"attr,attr"`
			}{},
			&zcl.BodySchema{
				Attributes: []zcl.AttributeSchema{
					{
						Name:     "attr",
						Required: false,
					},
				},
			},
			false,
		},
		{
			struct {
				Thing struct{} `zcl:"thing,block"`
			}{},
			&zcl.BodySchema{
				Blocks: []zcl.BlockHeaderSchema{
					{
						Type: "thing",
					},
				},
			},
			false,
		},
		{
			struct {
				Thing struct {
					Type string `zcl:"type,label"`
					Name string `zcl:"name,label"`
				} `zcl:"thing,block"`
			}{},
			&zcl.BodySchema{
				Blocks: []zcl.BlockHeaderSchema{
					{
						Type:       "thing",
						LabelNames: []string{"type", "name"},
					},
				},
			},
			false,
		},
		{
			struct {
				Thing []struct {
					Type string `zcl:"type,label"`
					Name string `zcl:"name,label"`
				} `zcl:"thing,block"`
			}{},
			&zcl.BodySchema{
				Blocks: []zcl.BlockHeaderSchema{
					{
						Type:       "thing",
						LabelNames: []string{"type", "name"},
					},
				},
			},
			false,
		},
		{
			struct {
				Thing *struct {
					Type string `zcl:"type,label"`
					Name string `zcl:"name,label"`
				} `zcl:"thing,block"`
			}{},
			&zcl.BodySchema{
				Blocks: []zcl.BlockHeaderSchema{
					{
						Type:       "thing",
						LabelNames: []string{"type", "name"},
					},
				},
			},
			false,
		},
		{
			struct {
				Thing struct {
					Name      string `zcl:"name,label"`
					Something string `zcl:"something"`
				} `zcl:"thing,block"`
			}{},
			&zcl.BodySchema{
				Blocks: []zcl.BlockHeaderSchema{
					{
						Type:       "thing",
						LabelNames: []string{"name"},
					},
				},
			},
			false,
		},
		{
			struct {
				Doodad string `zcl:"doodad"`
				Thing  struct {
					Name string `zcl:"name,label"`
				} `zcl:"thing,block"`
			}{},
			&zcl.BodySchema{
				Attributes: []zcl.AttributeSchema{
					{
						Name:     "doodad",
						Required: true,
					},
				},
				Blocks: []zcl.BlockHeaderSchema{
					{
						Type:       "thing",
						LabelNames: []string{"name"},
					},
				},
			},
			false,
		},
		{
			struct {
				Doodad string `zcl:"doodad"`
				Config string `zcl:",remain"`
			}{},
			&zcl.BodySchema{
				Attributes: []zcl.AttributeSchema{
					{
						Name:     "doodad",
						Required: true,
					},
				},
			},
			true,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v", test.val), func(t *testing.T) {
			schema, partial := ImpliedBodySchema(test.val)
			if !reflect.DeepEqual(schema, test.wantSchema) {
				t.Errorf(
					"wrong schema\ngot:  %s\nwant: %s",
					spew.Sdump(schema), spew.Sdump(test.wantSchema),
				)
			}

			if partial != test.wantPartial {
				t.Errorf(
					"wrong partial flag\ngot:  %#v\nwant: %#v",
					partial, test.wantPartial,
				)
			}
		})
	}
}
