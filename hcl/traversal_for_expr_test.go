package hcl

import (
	"testing"
)

type asTraversalSupported struct {
	staticExpr
	RootName string
}

type asTraversalSupportedAttr struct {
	staticExpr
	RootName string
	AttrName string
}

type asTraversalNotSupported struct {
	staticExpr
}

type asTraversalDeclined struct {
	staticExpr
}

type asTraversalWrappedDelegated struct {
	original Expression
	staticExpr
}

func (e asTraversalSupported) AsTraversal() Traversal {
	return Traversal{
		TraverseRoot{
			Name: e.RootName,
		},
	}
}

func (e asTraversalSupportedAttr) AsTraversal() Traversal {
	return Traversal{
		TraverseRoot{
			Name: e.RootName,
		},
		TraverseAttr{
			Name: e.AttrName,
		},
	}
}

func (e asTraversalDeclined) AsTraversal() Traversal {
	return nil
}

func (e asTraversalWrappedDelegated) UnwrapExpression() Expression {
	return e.original
}

func TestAbsTraversalForExpr(t *testing.T) {
	tests := []struct {
		Expr         Expression
		WantRootName string
	}{
		{
			asTraversalSupported{RootName: "foo"},
			"foo",
		},
		{
			asTraversalNotSupported{},
			"",
		},
		{
			asTraversalDeclined{},
			"",
		},
		{
			asTraversalWrappedDelegated{
				original: asTraversalSupported{RootName: "foo"},
			},
			"foo",
		},
		{
			asTraversalWrappedDelegated{
				original: asTraversalWrappedDelegated{
					original: asTraversalSupported{RootName: "foo"},
				},
			},
			"foo",
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			got, diags := AbsTraversalForExpr(test.Expr)
			switch {
			case got != nil:
				if test.WantRootName == "" {
					t.Fatalf("traversal was returned; want error")
				}
				if len(got) != 1 {
					t.Fatalf("wrong traversal length %d; want 1", len(got))
				}
				gotRoot, ok := got[0].(TraverseRoot)
				if !ok {
					t.Fatalf("first traversal step is %T; want hcl.TraverseRoot", got[0])
				}
				if gotRoot.Name != test.WantRootName {
					t.Errorf("wrong root name %q; want %q", gotRoot.Name, test.WantRootName)
				}
			default:
				if !diags.HasErrors() {
					t.Errorf("returned nil traversal without error diagnostics")
				}
				if test.WantRootName != "" {
					t.Errorf("traversal was not returned; want TraverseRoot(%q)", test.WantRootName)
				}
			}
		})
	}
}

func TestRelTraversalForExpr(t *testing.T) {
	tests := []struct {
		Expr          Expression
		WantFirstName string
	}{
		{
			asTraversalSupported{RootName: "foo"},
			"foo",
		},
		{
			asTraversalNotSupported{},
			"",
		},
		{
			asTraversalDeclined{},
			"",
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			got, diags := RelTraversalForExpr(test.Expr)
			switch {
			case got != nil:
				if test.WantFirstName == "" {
					t.Fatalf("traversal was returned; want error")
				}
				if len(got) != 1 {
					t.Fatalf("wrong traversal length %d; want 1", len(got))
				}
				gotRoot, ok := got[0].(TraverseAttr)
				if !ok {
					t.Fatalf("first traversal step is %T; want hcl.TraverseAttr", got[0])
				}
				if gotRoot.Name != test.WantFirstName {
					t.Errorf("wrong root name %q; want %q", gotRoot.Name, test.WantFirstName)
				}
			default:
				if !diags.HasErrors() {
					t.Errorf("returned nil traversal without error diagnostics")
				}
				if test.WantFirstName != "" {
					t.Errorf("traversal was not returned; want TraverseAttr(%q)", test.WantFirstName)
				}
			}
		})
	}
}

func TestExprAsKeyword(t *testing.T) {
	tests := []struct {
		Expr Expression
		Want string
	}{
		{
			asTraversalSupported{RootName: "foo"},
			"foo",
		},
		{
			asTraversalSupportedAttr{
				RootName: "foo",
				AttrName: "bar",
			},
			"",
		},
		{
			asTraversalNotSupported{},
			"",
		},
		{
			asTraversalDeclined{},
			"",
		},
		{
			asTraversalWrappedDelegated{
				original: asTraversalSupported{RootName: "foo"},
			},
			"foo",
		},
		{
			asTraversalWrappedDelegated{
				original: asTraversalWrappedDelegated{
					original: asTraversalSupported{RootName: "foo"},
				},
			},
			"foo",
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			got := ExprAsKeyword(test.Expr)
			if got != test.Want {
				t.Errorf("wrong result %q; want %q\ninput: %T", got, test.Want, test.Expr)
			}
		})
	}
}
