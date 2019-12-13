package integrationtest

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// TestTypeConvertFunc is an integration test of all of the layers involved
// in making the type conversion function from ext/typeexpr work.
//
// This requires co-operation between the hclsyntax package, the ext/typeexpr
// package, and the underlying cty functionality in order to work correctly.
//
// There are unit tests for the function implementation itself in the
// ext/typeexpr package, so this test is focused on making sure the function
// is given the opportunity to decode the second argument as a type expression
// when the function is called from HCL native syntax.
func TestTypeConvertFunc(t *testing.T) {
	// The convert function is special because it takes a type expression
	// rather than a value expression as its second argument. In this case,
	// we're asking it to convert a tuple into a list of strings:
	const exprSrc = `convert(["hello"], list(string))`
	// It achieves this by marking that second argument as being of a custom
	// type (a "capsule type", in cty terminology) that has a special
	// annotation which hclsyntax.FunctionCallExpr understands as allowing
	// the type to handle the analysis of the unevaluated expression, instead
	// of evaluating it as normal.
	//
	// To see more details of how this works, look at the definitions of
	// typexpr.TypeConstraintType and typeexpr.ConvertFunc, and at the
	// implementation of hclsyntax.FunctionCallExpr.Value.

	expr, diags := hclsyntax.ParseExpression([]byte(exprSrc), "", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatalf("unexpected problems: %s", diags.Error())
	}

	ctx := &hcl.EvalContext{
		Functions: map[string]function.Function{
			"convert": typeexpr.ConvertFunc,
		},
	}
	got, diags := expr.Value(ctx)
	if diags.HasErrors() {
		t.Fatalf("unexpected problems: %s", diags.Error())
	}
	want := cty.ListVal([]cty.Value{cty.StringVal("hello")})
	if !want.RawEquals(got) {
		t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, want)
	}
}
