package hcldec

import (
	"reflect"
	"testing"

	"github.com/apparentlymart/go-dump/dump"
	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
)

// Verify that all of our spec types implement the necessary interfaces
var _ Spec = ObjectSpec(nil)
var _ Spec = TupleSpec(nil)
var _ Spec = (*AttrSpec)(nil)
var _ Spec = (*LiteralSpec)(nil)
var _ Spec = (*ExprSpec)(nil)
var _ Spec = (*BlockSpec)(nil)
var _ Spec = (*BlockListSpec)(nil)
var _ Spec = (*BlockSetSpec)(nil)
var _ Spec = (*BlockMapSpec)(nil)
var _ Spec = (*BlockLabelSpec)(nil)
var _ Spec = (*DefaultSpec)(nil)
var _ Spec = (*TransformExprSpec)(nil)
var _ Spec = (*TransformFuncSpec)(nil)

var _ attrSpec = (*AttrSpec)(nil)
var _ attrSpec = (*DefaultSpec)(nil)

var _ blockSpec = (*BlockSpec)(nil)
var _ blockSpec = (*BlockListSpec)(nil)
var _ blockSpec = (*BlockSetSpec)(nil)
var _ blockSpec = (*BlockMapSpec)(nil)
var _ blockSpec = (*DefaultSpec)(nil)

var _ specNeedingVariables = (*AttrSpec)(nil)
var _ specNeedingVariables = (*BlockSpec)(nil)
var _ specNeedingVariables = (*BlockListSpec)(nil)
var _ specNeedingVariables = (*BlockSetSpec)(nil)
var _ specNeedingVariables = (*BlockMapSpec)(nil)

func TestDefaultSpec(t *testing.T) {
	config := `
foo = fooval
bar = barval
`
	f, diags := hclsyntax.ParseConfig([]byte(config), "", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatal(diags.Error())
	}

	t.Run("primary set", func(t *testing.T) {
		spec := &DefaultSpec{
			Primary: &AttrSpec{
				Name: "foo",
				Type: cty.String,
			},
			Default: &AttrSpec{
				Name: "bar",
				Type: cty.String,
			},
		}

		gotVars := Variables(f.Body, spec)
		wantVars := []hcl.Traversal{
			{
				hcl.TraverseRoot{
					Name: "fooval",
					SrcRange: hcl.Range{
						Filename: "",
						Start:    hcl.Pos{Line: 2, Column: 7, Byte: 7},
						End:      hcl.Pos{Line: 2, Column: 13, Byte: 13},
					},
				},
			},
			{
				hcl.TraverseRoot{
					Name: "barval",
					SrcRange: hcl.Range{
						Filename: "",
						Start:    hcl.Pos{Line: 3, Column: 7, Byte: 20},
						End:      hcl.Pos{Line: 3, Column: 13, Byte: 26},
					},
				},
			},
		}
		if !reflect.DeepEqual(gotVars, wantVars) {
			t.Errorf("wrong Variables result\ngot: %s\nwant: %s", dump.Value(gotVars), dump.Value(wantVars))
		}

		ctx := &hcl.EvalContext{
			Variables: map[string]cty.Value{
				"fooval": cty.StringVal("foo value"),
				"barval": cty.StringVal("bar value"),
			},
		}

		got, err := Decode(f.Body, spec, ctx)
		if err != nil {
			t.Fatal(err)
		}
		want := cty.StringVal("foo value")
		if !got.RawEquals(want) {
			t.Errorf("wrong Decode result\ngot:  %#v\nwant: %#v", got, want)
		}
	})

	t.Run("primary not set", func(t *testing.T) {
		spec := &DefaultSpec{
			Primary: &AttrSpec{
				Name: "foo",
				Type: cty.String,
			},
			Default: &AttrSpec{
				Name: "bar",
				Type: cty.String,
			},
		}

		ctx := &hcl.EvalContext{
			Variables: map[string]cty.Value{
				"fooval": cty.NullVal(cty.String),
				"barval": cty.StringVal("bar value"),
			},
		}

		got, err := Decode(f.Body, spec, ctx)
		if err != nil {
			t.Fatal(err)
		}
		want := cty.StringVal("bar value")
		if !got.RawEquals(want) {
			t.Errorf("wrong Decode result\ngot:  %#v\nwant: %#v", got, want)
		}
	})
}
