// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hcldec

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/apparentlymart/go-dump/dump"
	"github.com/google/go-cmp/cmp"
	"github.com/zclconf/go-cty-debug/ctydebug"
	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
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
var _ Spec = (*BlockAttrsSpec)(nil)
var _ Spec = (*BlockLabelSpec)(nil)
var _ Spec = (*DefaultSpec)(nil)
var _ Spec = (*TransformExprSpec)(nil)
var _ Spec = (*TransformFuncSpec)(nil)
var _ Spec = (*ValidateSpec)(nil)

var _ attrSpec = (*AttrSpec)(nil)
var _ attrSpec = (*DefaultSpec)(nil)

var _ blockSpec = (*BlockSpec)(nil)
var _ blockSpec = (*BlockListSpec)(nil)
var _ blockSpec = (*BlockSetSpec)(nil)
var _ blockSpec = (*BlockMapSpec)(nil)
var _ blockSpec = (*BlockAttrsSpec)(nil)
var _ blockSpec = (*DefaultSpec)(nil)

var _ specNeedingVariables = (*AttrSpec)(nil)
var _ specNeedingVariables = (*BlockSpec)(nil)
var _ specNeedingVariables = (*BlockListSpec)(nil)
var _ specNeedingVariables = (*BlockSetSpec)(nil)
var _ specNeedingVariables = (*BlockMapSpec)(nil)
var _ specNeedingVariables = (*BlockAttrsSpec)(nil)

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

func TestValidateFuncSpec(t *testing.T) {
	config := `
foo = "invalid"
`
	f, diags := hclsyntax.ParseConfig([]byte(config), "", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatal(diags.Error())
	}

	expectRange := map[string]*hcl.Range{
		"without_range": nil,
		"with_range": &hcl.Range{
			Filename: "foobar",
			Start:    hcl.Pos{Line: 99, Column: 99},
			End:      hcl.Pos{Line: 999, Column: 999},
		},
	}

	for name := range expectRange {
		t.Run(name, func(t *testing.T) {
			spec := &ValidateSpec{
				Wrapped: &AttrSpec{
					Name: "foo",
					Type: cty.String,
				},
				Func: func(value cty.Value) hcl.Diagnostics {
					if value.AsString() != "invalid" {
						return hcl.Diagnostics{
							&hcl.Diagnostic{
								Severity: hcl.DiagError,
								Summary:  "incorrect value",
								Detail:   fmt.Sprintf("invalid value passed in: %s", value.GoString()),
							},
						}
					}

					return hcl.Diagnostics{
						&hcl.Diagnostic{
							Severity: hcl.DiagWarning,
							Summary:  "OK",
							Detail:   "validation called correctly",
							Subject:  expectRange[name],
						},
					}
				},
			}

			_, diags = Decode(f.Body, spec, nil)
			if len(diags) != 1 ||
				diags[0].Severity != hcl.DiagWarning ||
				diags[0].Summary != "OK" ||
				diags[0].Detail != "validation called correctly" {
				t.Fatalf("unexpected diagnostics: %s", diags.Error())
			}

			if expectRange[name] == nil && diags[0].Subject == nil {
				t.Fatal("returned diagnostic subject missing")
			}

			if expectRange[name] != nil && !reflect.DeepEqual(expectRange[name], diags[0].Subject) {
				t.Fatalf("expected range %s, got range %s", expectRange[name], diags[0].Subject)
			}
		})
	}
}

func TestRefineValueSpec(t *testing.T) {
	config := `
foo = "hello"
bar = unk
`

	f, diags := hclsyntax.ParseConfig([]byte(config), "", hcl.InitialPos)
	if diags.HasErrors() {
		t.Fatal(diags.Error())
	}

	attrSpec := func(name string) Spec {
		return &RefineValueSpec{
			// RefineValueSpec should typically have a ValidateSpec wrapped
			// inside it to catch any values that are outside of the required
			// range and return a helpful error message about it. In this
			// case our refinement is .NotNull so the validation function
			// must reject null values.
			Wrapped: &ValidateSpec{
				Wrapped: &AttrSpec{
					Name:     name,
					Required: true,
					Type:     cty.String,
				},
				Func: func(value cty.Value) hcl.Diagnostics {
					var diags hcl.Diagnostics
					if value.IsNull() {
						diags = diags.Append(&hcl.Diagnostic{
							Severity: hcl.DiagError,
							Summary:  "Cannot be null",
							Detail:   "Argument is required.",
						})
					}
					return diags
				},
			},
			Refine: func(rb *cty.RefinementBuilder) *cty.RefinementBuilder {
				return rb.NotNull()
			},
		}
	}
	spec := &ObjectSpec{
		"foo": attrSpec("foo"),
		"bar": attrSpec("bar"),
	}

	got, diags := Decode(f.Body, spec, &hcl.EvalContext{
		Variables: map[string]cty.Value{
			"unk": cty.UnknownVal(cty.String),
		},
	})
	if diags.HasErrors() {
		t.Fatal(diags.Error())
	}

	want := cty.ObjectVal(map[string]cty.Value{
		// This argument had a known value, so it's unchanged but the
		// RefineValueSpec still checks that it isn't null to catch
		// bugs in the application's validation function.
		"foo": cty.StringVal("hello"),

		// The final value of bar is unknown but refined as non-null.
		"bar": cty.UnknownVal(cty.String).RefineNotNull(),
	})
	if diff := cmp.Diff(want, got, ctydebug.CmpOptions); diff != "" {
		t.Errorf("wrong result\n%s", diff)
	}
}
