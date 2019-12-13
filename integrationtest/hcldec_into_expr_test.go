package integrationtest

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/customdecode"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// TestHCLDecDecodeToExpr tests both hcldec's support for types with custom
// expression decoding rules and the two expression capsule types implemented
// in ext/customdecode. This mechanism requires cooperation between those
// two components and cty in order to work, so it's helpful to exercise it in
// an integration test.
func TestHCLDecDecodeToExpr(t *testing.T) {
	// Here we're going to capture the structure of two simple expressions
	// without immediately evaluating them.
	const input = `
a = foo
b = foo
c = "hello"
`
	// We'll capture "a" directly as an expression, losing its evaluation
	// context but retaining its structure. We'll capture "b" as a
	// customdecode.ExpressionClosure, which gives us both the expression
	// itself and the evaluation context it was originally evaluated in.
	// We also have "c" here just to make sure we can still decode into a
	// "normal" type via standard expression evaluation.

	f, diags := hclsyntax.ParseConfig([]byte(input), "", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatalf("unexpected problems: %s", diags.Error())
	}

	spec := hcldec.ObjectSpec{
		"a": &hcldec.AttrSpec{
			Name:     "a",
			Type:     customdecode.ExpressionType,
			Required: true,
		},
		"b": &hcldec.AttrSpec{
			Name:     "b",
			Type:     customdecode.ExpressionClosureType,
			Required: true,
		},
		"c": &hcldec.AttrSpec{
			Name:     "c",
			Type:     cty.String,
			Required: true,
		},
	}
	ctx := &hcl.EvalContext{
		Variables: map[string]cty.Value{
			"foo": cty.StringVal("foo value"),
		},
	}
	objVal, diags := hcldec.Decode(f.Body, spec, ctx)
	if diags.HasErrors() {
		t.Fatalf("unexpected problems: %s", diags.Error())
	}

	aVal := objVal.GetAttr("a")
	bVal := objVal.GetAttr("b")
	cVal := objVal.GetAttr("c")

	if got, want := aVal.Type(), customdecode.ExpressionType; !got.Equals(want) {
		t.Fatalf("wrong type for 'a'\ngot:  %#v\nwant: %#v", got, want)
	}
	if got, want := bVal.Type(), customdecode.ExpressionClosureType; !got.Equals(want) {
		t.Fatalf("wrong type for 'b'\ngot:  %#v\nwant: %#v", got, want)
	}
	if got, want := cVal.Type(), cty.String; !got.Equals(want) {
		t.Fatalf("wrong type for 'c'\ngot:  %#v\nwant: %#v", got, want)
	}

	gotAExpr := customdecode.ExpressionFromVal(aVal)
	wantAExpr := &hclsyntax.ScopeTraversalExpr{
		Traversal: hcl.Traversal{
			hcl.TraverseRoot{
				Name: "foo",
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 5, Byte: 5},
					End:   hcl.Pos{Line: 2, Column: 8, Byte: 8},
				},
			},
		},
		SrcRange: hcl.Range{
			Start: hcl.Pos{Line: 2, Column: 5, Byte: 5},
			End:   hcl.Pos{Line: 2, Column: 8, Byte: 8},
		},
	}
	if diff := cmp.Diff(wantAExpr, gotAExpr, cmpopts.IgnoreUnexported(hcl.TraverseRoot{})); diff != "" {
		t.Errorf("wrong expression for a\n%s", diff)
	}

	bClosure := customdecode.ExpressionClosureFromVal(bVal)
	gotBVal, diags := bClosure.Value()
	wantBVal := cty.StringVal("foo value")
	if diags.HasErrors() {
		t.Fatalf("unexpected problems: %s", diags.Error())
	}
	if got, want := gotBVal, wantBVal; !want.RawEquals(got) {
		t.Errorf("wrong 'b' result\ngot:  %#v\nwant: %#v", got, want)
	}

	if got, want := cVal, cty.StringVal("hello"); !want.RawEquals(got) {
		t.Errorf("wrong 'c'\ngot:  %#v\nwant: %#v", got, want)
	}

	// One additional "trick" we can do with the expression closure is to
	// evaluate the expression in a _derived_ EvalContext, rather than the
	// captured one. This could be useful for introducing additional local
	// variables/functions in a particular context, for example.
	deriveCtx := bClosure.EvalContext.NewChild()
	deriveCtx.Variables = map[string]cty.Value{
		"foo": cty.StringVal("overridden foo value"),
	}
	gotBVal2, diags := bClosure.Expression.Value(deriveCtx)
	wantBVal2 := cty.StringVal("overridden foo value")
	if diags.HasErrors() {
		t.Fatalf("unexpected problems: %s", diags.Error())
	}
	if got, want := gotBVal2, wantBVal2; !want.RawEquals(got) {
		t.Errorf("wrong 'b' result with derived EvalContext\ngot:  %#v\nwant: %#v", got, want)
	}
}
