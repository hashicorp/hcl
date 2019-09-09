package hclsyntax

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/kr/pretty"
	"github.com/zclconf/go-cty/cty"
)

func TestVariables(t *testing.T) {
	tests := []struct {
		Expr Expression
		Want []hcl.Traversal
	}{
		{
			&LiteralValueExpr{
				Val: cty.True,
			},
			nil,
		},
		{
			&ScopeTraversalExpr{
				Traversal: hcl.Traversal{
					hcl.TraverseRoot{
						Name: "foo",
					},
				},
			},
			[]hcl.Traversal{
				{
					hcl.TraverseRoot{
						Name: "foo",
					},
				},
			},
		},
		{
			&BinaryOpExpr{
				LHS: &ScopeTraversalExpr{
					Traversal: hcl.Traversal{
						hcl.TraverseRoot{
							Name: "foo",
						},
					},
				},
				Op: OpAdd,
				RHS: &ScopeTraversalExpr{
					Traversal: hcl.Traversal{
						hcl.TraverseRoot{
							Name: "bar",
						},
					},
				},
			},
			[]hcl.Traversal{
				{
					hcl.TraverseRoot{
						Name: "foo",
					},
				},
				{
					hcl.TraverseRoot{
						Name: "bar",
					},
				},
			},
		},
		{
			&UnaryOpExpr{
				Val: &ScopeTraversalExpr{
					Traversal: hcl.Traversal{
						hcl.TraverseRoot{
							Name: "foo",
						},
					},
				},
				Op: OpNegate,
			},
			[]hcl.Traversal{
				{
					hcl.TraverseRoot{
						Name: "foo",
					},
				},
			},
		},
		{
			&ConditionalExpr{
				Condition: &ScopeTraversalExpr{
					Traversal: hcl.Traversal{
						hcl.TraverseRoot{
							Name: "foo",
						},
					},
				},
				TrueResult: &ScopeTraversalExpr{
					Traversal: hcl.Traversal{
						hcl.TraverseRoot{
							Name: "bar",
						},
					},
				},
				FalseResult: &ScopeTraversalExpr{
					Traversal: hcl.Traversal{
						hcl.TraverseRoot{
							Name: "baz",
						},
					},
				},
			},
			[]hcl.Traversal{
				{
					hcl.TraverseRoot{
						Name: "foo",
					},
				},
				{
					hcl.TraverseRoot{
						Name: "bar",
					},
				},
				{
					hcl.TraverseRoot{
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
					Traversal: hcl.Traversal{
						hcl.TraverseRoot{
							Name: "foo",
						},
					},
				},
				KeyExpr: &BinaryOpExpr{
					LHS: &ScopeTraversalExpr{
						Traversal: hcl.Traversal{
							hcl.TraverseRoot{
								Name: "k",
							},
						},
					},
					Op: OpAdd,
					RHS: &ScopeTraversalExpr{
						Traversal: hcl.Traversal{
							hcl.TraverseRoot{
								Name: "bar",
							},
						},
					},
				},
				ValExpr: &BinaryOpExpr{
					LHS: &ScopeTraversalExpr{
						Traversal: hcl.Traversal{
							hcl.TraverseRoot{
								Name: "v",
							},
						},
					},
					Op: OpAdd,
					RHS: &ScopeTraversalExpr{
						Traversal: hcl.Traversal{
							hcl.TraverseRoot{
								Name: "baz",
							},
						},
					},
				},
				CondExpr: &BinaryOpExpr{
					LHS: &ScopeTraversalExpr{
						Traversal: hcl.Traversal{
							hcl.TraverseRoot{
								Name: "k",
							},
						},
					},
					Op: OpLessThan,
					RHS: &ScopeTraversalExpr{
						Traversal: hcl.Traversal{
							hcl.TraverseRoot{
								Name: "limit",
							},
						},
					},
				},
			},
			[]hcl.Traversal{
				{
					hcl.TraverseRoot{
						Name: "foo",
					},
				},
				{
					hcl.TraverseRoot{
						Name: "bar",
					},
				},
				{
					hcl.TraverseRoot{
						Name: "baz",
					},
				},
				{
					hcl.TraverseRoot{
						Name: "limit",
					},
				},
			},
		},
		{
			&ScopeTraversalExpr{
				Traversal: hcl.Traversal{
					hcl.TraverseRoot{
						Name: "data",
					},
					hcl.TraverseAttr{
						Name: "null_data_source",
					},
					hcl.TraverseAttr{
						Name: "multi",
					},
					hcl.TraverseIndex{
						Key: cty.NumberFloatVal(0),
					},
				},
			},
			[]hcl.Traversal{
				{
					hcl.TraverseRoot{
						Name: "data",
					},
					hcl.TraverseAttr{
						Name: "null_data_source",
					},
					hcl.TraverseAttr{
						Name: "multi",
					},
					hcl.TraverseIndex{
						Key: cty.NumberFloatVal(0),
					},
				},
			},
		},
		{
			&RelativeTraversalExpr{
				Source: &FunctionCallExpr{
					Name: "sort",
					Args: []Expression{
						&ScopeTraversalExpr{
							Traversal: hcl.Traversal{
								hcl.TraverseRoot{
									Name: "data",
								},
								hcl.TraverseAttr{
									Name: "null_data_source",
								},
								hcl.TraverseAttr{
									Name: "multi",
								},
							},
						},
					},
				},
				Traversal: hcl.Traversal{
					hcl.TraverseIndex{
						Key: cty.NumberFloatVal(0),
					},
				},
			},
			[]hcl.Traversal{
				{
					hcl.TraverseRoot{
						Name: "data",
					},
					hcl.TraverseAttr{
						Name: "null_data_source",
					},
					hcl.TraverseAttr{
						Name: "multi",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v", test.Expr), func(t *testing.T) {
			got := Variables(test.Expr)

			if !reflect.DeepEqual(got, test.Want) {
				t.Errorf("wrong result\ngot:  %s\nwant: %s", pretty.Sprint(got), pretty.Sprint(test.Want))
			}
		})
	}
}
