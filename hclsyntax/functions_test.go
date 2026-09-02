// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hclsyntax

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

// Covers similar cases to hclsyntax/variables_test.go

func TestFunctions(t *testing.T) {
	tests := []struct {
		Expr Expression
		Want []hcl.Traversal
	}{
		{
			&LiteralValueExpr{
				Val: cty.True,
			},
			[]hcl.Traversal{},
		},
		{
			&FunctionCallExpr{
				Name: "funky",
			},
			[]hcl.Traversal{
				{
					hcl.TraverseRoot{
						Name: "funky",
					},
				},
			},
		},
		{
			&BinaryOpExpr{
				LHS: &FunctionCallExpr{
					Name: "lhs",
				},
				Op: OpAdd,
				RHS: &FunctionCallExpr{
					Name: "rhs",
				},
			},
			[]hcl.Traversal{
				{
					hcl.TraverseRoot{
						Name: "lhs",
					},
				},
				{
					hcl.TraverseRoot{
						Name: "rhs",
					},
				},
			},
		},
		{
			&UnaryOpExpr{
				Val: &FunctionCallExpr{
					Name: "neg",
				},
				Op: OpNegate,
			},
			[]hcl.Traversal{
				{
					hcl.TraverseRoot{
						Name: "neg",
					},
				},
			},
		},
		{
			&ConditionalExpr{
				Condition: &FunctionCallExpr{
					Name: "cond",
				},
				TrueResult: &FunctionCallExpr{
					Name: "true",
				},
				FalseResult: &FunctionCallExpr{
					Name: "false",
				},
			},
			[]hcl.Traversal{
				{
					hcl.TraverseRoot{
						Name: "cond",
					},
				},
				{
					hcl.TraverseRoot{
						Name: "true",
					},
				},
				{
					hcl.TraverseRoot{
						Name: "false",
					},
				},
			},
		},
		{
			&ForExpr{
				KeyVar: "k",
				ValVar: "v",

				CollExpr: &FunctionCallExpr{
					Name: "coll",
				},
				KeyExpr: &BinaryOpExpr{
					LHS: &FunctionCallExpr{
						Name: "key_lhs",
					},
					Op: OpAdd,
					RHS: &FunctionCallExpr{
						Name: "key_rhs",
					},
				},
				ValExpr: &BinaryOpExpr{
					LHS: &FunctionCallExpr{
						Name: "val_lhs",
					},
					Op: OpAdd,
					RHS: &FunctionCallExpr{
						Name: "val_rhs",
					},
				},
				CondExpr: &BinaryOpExpr{
					LHS: &FunctionCallExpr{
						Name: "cond_lhs",
					},
					Op: OpLessThan,
					RHS: &FunctionCallExpr{
						Name: "cond_rhs",
					},
				},
			},
			[]hcl.Traversal{
				{
					hcl.TraverseRoot{
						Name: "coll",
					},
				},
				{
					hcl.TraverseRoot{
						Name: "key_lhs",
					},
				},
				{
					hcl.TraverseRoot{
						Name: "key_rhs",
					},
				},
				{
					hcl.TraverseRoot{
						Name: "val_lhs",
					},
				},
				{
					hcl.TraverseRoot{
						Name: "val_rhs",
					},
				},
				{
					hcl.TraverseRoot{
						Name: "cond_lhs",
					},
				},
				{
					hcl.TraverseRoot{
						Name: "cond_rhs",
					},
				},
			},
		},
		{
			&FunctionCallExpr{
				Name: "funky",
				Args: []Expression{
					&FunctionCallExpr{
						Name: "sub_a",
					},
					&FunctionCallExpr{
						Name: "sub_b",
					},
				},
			},
			[]hcl.Traversal{
				{
					hcl.TraverseRoot{
						Name: "funky",
					},
				},
				{
					hcl.TraverseRoot{
						Name: "sub_a",
					},
				},
				{
					hcl.TraverseRoot{
						Name: "sub_b",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v", test.Expr), func(t *testing.T) {
			got := Functions(test.Expr)

			if !reflect.DeepEqual(got, test.Want) {
				t.Errorf(
					"wrong result\ngot:  %s\nwant: %s",
					spew.Sdump(got), spew.Sdump(test.Want),
				)
			}
		})
	}
}
