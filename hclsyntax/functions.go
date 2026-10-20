// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hclsyntax

import (
	"github.com/hashicorp/hcl/v2"
)

func Functions(expr Expression) []hcl.Traversal {
	walker := make(fnWalker, 0)
	Walk(expr, &walker)
	return walker
}

type fnWalker []hcl.Traversal

func (w *fnWalker) Enter(node Node) hcl.Diagnostics {
	if fn, ok := node.(*FunctionCallExpr); ok {
		*w = append(*w, hcl.Traversal{hcl.TraverseRoot{
			Name:     fn.Name,
			SrcRange: fn.NameRange,
		}})
	}
	return nil
}
func (w *fnWalker) Exit(node Node) hcl.Diagnostics {
	return nil
}
