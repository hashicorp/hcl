package zclsyntax

import (
	"fmt"
	"testing"

	"reflect"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/hcl2/zcl"
	"github.com/zclconf/go-cty/cty"
)

func TestVariables(t *testing.T) {
	tests := []struct {
		Expr Expression
		Want []zcl.Traversal
	}{
		{
			&LiteralValueExpr{
				Val: cty.True,
			},
			nil,
		},
		{
			&ScopeTraversalExpr{
				Traversal: zcl.Traversal{
					zcl.TraverseRoot{
						Name: "foo",
					},
				},
			},
			[]zcl.Traversal{
				{
					zcl.TraverseRoot{
						Name: "foo",
					},
				},
			},
		},
		{
			&BinaryOpExpr{
				LHS: &ScopeTraversalExpr{
					Traversal: zcl.Traversal{
						zcl.TraverseRoot{
							Name: "foo",
						},
					},
				},
				Op: OpAdd,
				RHS: &ScopeTraversalExpr{
					Traversal: zcl.Traversal{
						zcl.TraverseRoot{
							Name: "bar",
						},
					},
				},
			},
			[]zcl.Traversal{
				{
					zcl.TraverseRoot{
						Name: "foo",
					},
				},
				{
					zcl.TraverseRoot{
						Name: "bar",
					},
				},
			},
		},
		{
			&UnaryOpExpr{
				Val: &ScopeTraversalExpr{
					Traversal: zcl.Traversal{
						zcl.TraverseRoot{
							Name: "foo",
						},
					},
				},
				Op: OpNegate,
			},
			[]zcl.Traversal{
				{
					zcl.TraverseRoot{
						Name: "foo",
					},
				},
			},
		},
		{
			&ConditionalExpr{
				Condition: &ScopeTraversalExpr{
					Traversal: zcl.Traversal{
						zcl.TraverseRoot{
							Name: "foo",
						},
					},
				},
				TrueResult: &ScopeTraversalExpr{
					Traversal: zcl.Traversal{
						zcl.TraverseRoot{
							Name: "bar",
						},
					},
				},
				FalseResult: &ScopeTraversalExpr{
					Traversal: zcl.Traversal{
						zcl.TraverseRoot{
							Name: "baz",
						},
					},
				},
			},
			[]zcl.Traversal{
				{
					zcl.TraverseRoot{
						Name: "foo",
					},
				},
				{
					zcl.TraverseRoot{
						Name: "bar",
					},
				},
				{
					zcl.TraverseRoot{
						Name: "baz",
					},
				},
			},
		},
		{
			&ForExpr{
				KeyVar: "k",
				ValVar: "v",

				CollExpr: &ScopeTraversalExpr{
					Traversal: zcl.Traversal{
						zcl.TraverseRoot{
							Name: "foo",
						},
					},
				},
				KeyExpr: &BinaryOpExpr{
					LHS: &ScopeTraversalExpr{
						Traversal: zcl.Traversal{
							zcl.TraverseRoot{
								Name: "k",
							},
						},
					},
					Op: OpAdd,
					RHS: &ScopeTraversalExpr{
						Traversal: zcl.Traversal{
							zcl.TraverseRoot{
								Name: "bar",
							},
						},
					},
				},
				ValExpr: &BinaryOpExpr{
					LHS: &ScopeTraversalExpr{
						Traversal: zcl.Traversal{
							zcl.TraverseRoot{
								Name: "v",
							},
						},
					},
					Op: OpAdd,
					RHS: &ScopeTraversalExpr{
						Traversal: zcl.Traversal{
							zcl.TraverseRoot{
								Name: "baz",
							},
						},
					},
				},
				CondExpr: &BinaryOpExpr{
					LHS: &ScopeTraversalExpr{
						Traversal: zcl.Traversal{
							zcl.TraverseRoot{
								Name: "k",
							},
						},
					},
					Op: OpLessThan,
					RHS: &ScopeTraversalExpr{
						Traversal: zcl.Traversal{
							zcl.TraverseRoot{
								Name: "limit",
							},
						},
					},
				},
			},
			[]zcl.Traversal{
				{
					zcl.TraverseRoot{
						Name: "foo",
					},
				},
				{
					zcl.TraverseRoot{
						Name: "bar",
					},
				},
				{
					zcl.TraverseRoot{
						Name: "baz",
					},
				},
				{
					zcl.TraverseRoot{
						Name: "limit",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v", test.Expr), func(t *testing.T) {
			got := Variables(test.Expr)

			if !reflect.DeepEqual(got, test.Want) {
				t.Errorf("wrong result\ngot:  %s\nwant: %s", spew.Sdump(got), spew.Sdump(test.Want))
			}
		})
	}
}
