package hclsyntax

import (
	"testing"

	"github.com/go-test/deep"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/zclconf/go-cty/cty"
)

func TestTraversalStatic(t *testing.T) {
	expr, diags := ParseExpression([]byte(`a.b.c`), "", hcl.Pos{Line: 1, Column: 1})
	got, moreDiags := hcl.AbsTraversalForExpr(expr)
	diags = append(diags, moreDiags...)

	if len(diags) != 0 {
		t.Errorf("wrong number of diags %d; want 0", len(diags))
		for _, diag := range diags {
			t.Logf("- %s", diag)
		}
		return
	}

	want := hcl.Traversal{
		hcl.TraverseRoot{
			Name: "a",
			SrcRange: hcl.Range{
				Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
				End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
			},
		},
		hcl.TraverseAttr{
			Name: "b",
			SrcRange: hcl.Range{
				Start: hcl.Pos{Line: 1, Column: 2, Byte: 1},
				End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
			},
		},
		hcl.TraverseAttr{
			Name: "c",
			SrcRange: hcl.Range{
				Start: hcl.Pos{Line: 1, Column: 4, Byte: 3},
				End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
			},
		},
	}

	for _, problem := range deep.Equal(got, want) {
		t.Errorf(problem)
	}
}

func TestTupleStatic(t *testing.T) {
	expr, diags := ParseExpression([]byte(`[true, false]`), "", hcl.Pos{Line: 1, Column: 1})
	exprs, moreDiags := hcl.ExprList(expr)
	diags = append(diags, moreDiags...)
	if len(diags) != 0 {
		t.Errorf("wrong number of diags %d; want 0", len(diags))
		for _, diag := range diags {
			t.Logf("- %s", diag)
		}
		return
	}

	if got, want := len(exprs), 2; got != want {
		t.Fatalf("wrong length %d; want %d", got, want)
	}

	got := make([]cty.Value, len(exprs))
	want := []cty.Value{
		cty.True,
		cty.False,
	}
	for i, itemExpr := range exprs {
		val, valDiags := itemExpr.Value(nil)
		if len(valDiags) != 0 {
			t.Errorf("wrong number of diags %d; want 0", len(valDiags))
			for _, diag := range valDiags {
				t.Logf("- %s", diag)
			}
			return
		}
		got[i] = val
	}

	for _, problem := range deep.Equal(got, want) {
		t.Errorf(problem)
	}
}

func TestMapStatic(t *testing.T) {
	expr, diags := ParseExpression([]byte(`{"foo":true,"bar":false}`), "", hcl.Pos{Line: 1, Column: 1})
	items, moreDiags := hcl.ExprMap(expr)
	diags = append(diags, moreDiags...)
	if len(diags) != 0 {
		t.Errorf("wrong number of diags %d; want 0", len(diags))
		for _, diag := range diags {
			t.Logf("- %s", diag)
		}
		return
	}

	if got, want := len(items), 2; got != want {
		t.Fatalf("wrong length %d; want %d", got, want)
	}

	got := make(map[cty.Value]cty.Value)
	want := map[cty.Value]cty.Value{
		cty.StringVal("foo"): cty.True,
		cty.StringVal("bar"): cty.False,
	}
	for _, item := range items {
		var itemDiags hcl.Diagnostics
		key, keyDiags := item.Key.Value(nil)
		itemDiags = append(itemDiags, keyDiags...)
		val, valDiags := item.Value.Value(nil)
		itemDiags = append(itemDiags, valDiags...)
		if len(itemDiags) != 0 {
			t.Errorf("wrong number of diags %d; want 0", len(itemDiags))
			for _, diag := range itemDiags {
				t.Logf("- %s", diag)
			}
			return
		}
		got[key] = val
	}

	for _, problem := range deep.Equal(got, want) {
		t.Errorf(problem)
	}
}
