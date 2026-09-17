// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package hclwrite

func newFunctionCallExpr(name string) *FunctionCallExpr {
	return &FunctionCallExpr{
		name:   name,
		args:   newNodeSet(),
		inTree: newInTree(),
	}
}

type FunctionCallExpr struct {
	inTree

	name        string
	args        nodeSet
	expandFinal bool
}

func (call *FunctionCallExpr) Name() string {
	return call.name
}
