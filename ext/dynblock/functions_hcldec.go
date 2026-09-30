// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynblock

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hcldec"
)

// This is duplicated from ext/dynblock/variables_hcldec.go and modified to suit functions

// FunctionsHCLDec is a wrapper around WalkFunctions that uses the given hcldec
// specification to automatically drive the recursive walk through nested
// blocks in the given body.
//
// This is a drop-in replacement for hcldec.Functions which is able to treat
// blocks of type "dynamic" in the same special way that dynblock.Expand would,
// exposing both the functions referenced in the "for_each" and "labels"
// arguments and functions used in the nested "content" block.
func FunctionsHCLDec(body hcl.Body, spec hcldec.Spec) []hcl.Traversal {
	rootNode := WalkFunctions(body)
	return walkFunctionsWithHCLDec(rootNode, spec)
}

// ExpandFunctionsHCLDec is like FunctionsHCLDec but it includes only the
// minimal set of functions required to call Expand, ignoring functions that
// are referenced only inside normal block contents. See WalkExpandFunctions
// for more information.
func ExpandFunctionsHCLDec(body hcl.Body, spec hcldec.Spec) []hcl.Traversal {
	rootNode := WalkExpandFunctions(body)
	return walkFunctionsWithHCLDec(rootNode, spec)
}

func walkFunctionsWithHCLDec(node WalkFunctionsNode, spec hcldec.Spec) []hcl.Traversal {
	vars, children := node.Visit(hcldec.ImpliedSchema(spec))

	if len(children) > 0 {
		childSpecs := hcldec.ChildBlockTypes(spec)
		for _, child := range children {
			if childSpec, exists := childSpecs[child.BlockTypeName]; exists {
				vars = append(vars, walkFunctionsWithHCLDec(child.Node, childSpec)...)
			}
		}
	}

	return vars
}
