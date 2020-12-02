package hclsyntax

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-test/deep"

	"github.com/hashicorp/hcl/v2"
)

func TestWalk(t *testing.T) {

	tests := []struct {
		src  string
		want []testWalkCall
	}{
		{
			`1`,
			[]testWalkCall{
				{testWalkEnter, "*hclsyntax.LiteralValueExpr"},
				{testWalkExit, "*hclsyntax.LiteralValueExpr"},
			},
		},
		{
			`foo`,
			[]testWalkCall{
				{testWalkEnter, "*hclsyntax.ScopeTraversalExpr"},
				{testWalkExit, "*hclsyntax.ScopeTraversalExpr"},
			},
		},
		{
			`1 + 1`,
			[]testWalkCall{
				{testWalkEnter, "*hclsyntax.BinaryOpExpr"},
				{testWalkEnter, "*hclsyntax.LiteralValueExpr"},
				{testWalkExit, "*hclsyntax.LiteralValueExpr"},
				{testWalkEnter, "*hclsyntax.LiteralValueExpr"},
				{testWalkExit, "*hclsyntax.LiteralValueExpr"},
				{testWalkExit, "*hclsyntax.BinaryOpExpr"},
			},
		},
		{
			`(1 + 1)`,
			[]testWalkCall{
				{testWalkEnter, "*hclsyntax.ParenthesesExpr"},
				{testWalkEnter, "*hclsyntax.BinaryOpExpr"},
				{testWalkEnter, "*hclsyntax.LiteralValueExpr"},
				{testWalkExit, "*hclsyntax.LiteralValueExpr"},
				{testWalkEnter, "*hclsyntax.LiteralValueExpr"},
				{testWalkExit, "*hclsyntax.LiteralValueExpr"},
				{testWalkExit, "*hclsyntax.BinaryOpExpr"},
				{testWalkExit, "*hclsyntax.ParenthesesExpr"},
			},
		},
		{
			`a[0]`,
			[]testWalkCall{
				// because the index is constant here, the index is absorbed into the traversal
				{testWalkEnter, "*hclsyntax.ScopeTraversalExpr"},
				{testWalkExit, "*hclsyntax.ScopeTraversalExpr"},
			},
		},
		{
			`0[foo]`, // semantically incorrect, but should still parse and be walkable
			[]testWalkCall{
				{testWalkEnter, "*hclsyntax.IndexExpr"},
				{testWalkEnter, "*hclsyntax.LiteralValueExpr"},
				{testWalkExit, "*hclsyntax.LiteralValueExpr"},
				{testWalkEnter, "*hclsyntax.ScopeTraversalExpr"},
				{testWalkExit, "*hclsyntax.ScopeTraversalExpr"},
				{testWalkExit, "*hclsyntax.IndexExpr"},
			},
		},
		{
			`bar()`,
			[]testWalkCall{
				{testWalkEnter, "*hclsyntax.FunctionCallExpr"},
				{testWalkExit, "*hclsyntax.FunctionCallExpr"},
			},
		},
		{
			`bar(1, a)`,
			[]testWalkCall{
				{testWalkEnter, "*hclsyntax.FunctionCallExpr"},
				{testWalkEnter, "*hclsyntax.LiteralValueExpr"},
				{testWalkExit, "*hclsyntax.LiteralValueExpr"},
				{testWalkEnter, "*hclsyntax.ScopeTraversalExpr"},
				{testWalkExit, "*hclsyntax.ScopeTraversalExpr"},
				{testWalkExit, "*hclsyntax.FunctionCallExpr"},
			},
		},
		{
			`bar(1, a)[0]`,
			[]testWalkCall{
				{testWalkEnter, "*hclsyntax.RelativeTraversalExpr"},
				{testWalkEnter, "*hclsyntax.FunctionCallExpr"},
				{testWalkEnter, "*hclsyntax.LiteralValueExpr"},
				{testWalkExit, "*hclsyntax.LiteralValueExpr"},
				{testWalkEnter, "*hclsyntax.ScopeTraversalExpr"},
				{testWalkExit, "*hclsyntax.ScopeTraversalExpr"},
				{testWalkExit, "*hclsyntax.FunctionCallExpr"},
				{testWalkExit, "*hclsyntax.RelativeTraversalExpr"},
			},
		},
		{
			`[for x in foo: x + 1 if x < 10]`,
			[]testWalkCall{
				{testWalkEnter, "*hclsyntax.ForExpr"},
				{testWalkEnter, "*hclsyntax.ScopeTraversalExpr"},
				{testWalkExit, "*hclsyntax.ScopeTraversalExpr"},
				{testWalkEnter, "hclsyntax.ChildScope"},
				{testWalkEnter, "*hclsyntax.BinaryOpExpr"},
				{testWalkEnter, "*hclsyntax.ScopeTraversalExpr"},
				{testWalkExit, "*hclsyntax.ScopeTraversalExpr"},
				{testWalkEnter, "*hclsyntax.LiteralValueExpr"},
				{testWalkExit, "*hclsyntax.LiteralValueExpr"},
				{testWalkExit, "*hclsyntax.BinaryOpExpr"},
				{testWalkExit, "hclsyntax.ChildScope"},
				{testWalkEnter, "hclsyntax.ChildScope"},
				{testWalkEnter, "*hclsyntax.BinaryOpExpr"},
				{testWalkEnter, "*hclsyntax.ScopeTraversalExpr"},
				{testWalkExit, "*hclsyntax.ScopeTraversalExpr"},
				{testWalkEnter, "*hclsyntax.LiteralValueExpr"},
				{testWalkExit, "*hclsyntax.LiteralValueExpr"},
				{testWalkExit, "*hclsyntax.BinaryOpExpr"},
				{testWalkExit, "hclsyntax.ChildScope"},
				{testWalkExit, "*hclsyntax.ForExpr"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.src, func(t *testing.T) {
			expr, diags := ParseExpression([]byte(test.src), "", hcl.Pos{Line: 1, Column: 1})
			if diags.HasErrors() {
				t.Fatalf("failed to parse expression: %s", diags.Error())
			}

			w := testWalker{}
			diags = Walk(expr, &w)
			if diags.HasErrors() {
				t.Fatalf("failed to walk: %s", diags.Error())
			}

			got := w.Calls
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("wrong calls\ngot: %swant: %s", spew.Sdump(got), spew.Sdump(test.want))
				for _, problem := range deep.Equal(got, test.want) {
					t.Errorf(problem)
				}
			}
		})
	}
}

type testWalkMethod int

const testWalkEnter testWalkMethod = 1
const testWalkExit testWalkMethod = 2

type testWalkCall struct {
	Method   testWalkMethod
	NodeType string
}

type testWalker struct {
	Calls []testWalkCall
}

func (w *testWalker) Enter(node Node) hcl.Diagnostics {
	w.Calls = append(w.Calls, testWalkCall{testWalkEnter, fmt.Sprintf("%T", node)})
	return nil
}

func (w *testWalker) Exit(node Node) hcl.Diagnostics {
	w.Calls = append(w.Calls, testWalkCall{testWalkExit, fmt.Sprintf("%T", node)})
	return nil
}
