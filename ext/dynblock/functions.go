// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynblock

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// This is duplicated from ext/dynblock/variables.go and modified to suit functions

// WalkFunctions begins the recursive process of walking all expressions and
// nested blocks in the given body and its child bodies while taking into
// account any "dynamic" blocks.
//
// This function requires that the caller walk through the nested block
// structure in the given body level-by-level so that an appropriate schema
// can be provided at each level to inform further processing. This workflow
// is thus easiest to use for calling applications that have some higher-level
// schema representation available with which to drive this multi-step
// process. If your application uses the hcldec package, you may be able to
// use FunctionsHCLDec instead for a more automatic approach.
func WalkFunctions(body hcl.Body) WalkFunctionsNode {
	return WalkFunctionsNode{
		body:           body,
		includeContent: true,
	}
}

// WalkExpandFunctions is like Functions but it includes only the functions
// required for successful block expansion, ignoring any functions referenced
// inside block contents. The result is the minimal set of all functions
// required for a call to Expand, excluding functions that would only be
// needed to subsequently call Content or PartialContent on the expanded
// body.
func WalkExpandFunctions(body hcl.Body) WalkFunctionsNode {
	return WalkFunctionsNode{
		body: body,
	}
}

type WalkFunctionsNode struct {
	body hcl.Body
	it   *iteration

	includeContent bool
}

type WalkFunctionsChild struct {
	BlockTypeName string
	Node          WalkFunctionsNode
}

// Body returns the HCL Body associated with the child node, in case the caller
// wants to do some sort of inspection of it in order to decide what schema
// to pass to Visit.
//
// Most implementations should just fetch a fixed schema based on the
// BlockTypeName field and not access this. Deciding on a schema dynamically
// based on the body is a strange thing to do and generally necessary only if
// your caller is already doing other bizarre things with HCL bodies.
func (c WalkFunctionsChild) Body() hcl.Body {
	return c.Node.body
}

// exprFunctions handles the
func exprFunctions(expr hcl.Expression) []hcl.Traversal {
	if ef, ok := expr.(hcl.ExpressionWithFunctions); ok {
		return ef.Functions()
	}
	// hclsyntax Fallback
	if hsexpr, ok := expr.(hclsyntax.Expression); ok {
		return hclsyntax.Functions(hsexpr)
	}
	// Not exposed
	return nil
}

// Visit returns the function traversals required for any "dynamic" blocks
// directly in the body associated with this node, and also returns any child
// nodes that must be visited in order to continue the walk.
//
// Each child node has its associated block type name given in its BlockTypeName
// field, which the calling application should use to determine the appropriate
// schema for the content of each child node and pass it to the child node's
// own Visit method to continue the walk recursively.
func (n WalkFunctionsNode) Visit(schema *hcl.BodySchema) (vars []hcl.Traversal, children []WalkFunctionsChild) {
	extSchema := n.extendSchema(schema)
	container, _, _ := n.body.PartialContent(extSchema)
	if container == nil {
		return vars, children
	}

	children = make([]WalkFunctionsChild, 0, len(container.Blocks))

	if n.includeContent {
		for _, attr := range container.Attributes {
			for _, traversal := range exprFunctions(attr.Expr) {
				var ours, inherited bool
				if n.it != nil {
					ours = traversal.RootName() == n.it.IteratorName
					_, inherited = n.it.Inherited[traversal.RootName()]
				}

				if !(ours || inherited) {
					vars = append(vars, traversal)
				}
			}
		}
	}

	for _, block := range container.Blocks {
		switch block.Type {

		case "dynamic":
			blockTypeName := block.Labels[0]
			inner, _, _ := block.Body.PartialContent(functionDetectionInnerSchema)
			if inner == nil {
				continue
			}

			iteratorName := blockTypeName
			if attr, exists := inner.Attributes["iterator"]; exists {
				iterTraversal, _ := hcl.AbsTraversalForExpr(attr.Expr)
				if len(iterTraversal) == 0 {
					// Ignore this invalid dynamic block, since it'll produce
					// an error if someone tries to extract content from it
					// later anyway.
					continue
				}
				iteratorName = iterTraversal.RootName()
			}
			blockIt := n.it.MakeChild(iteratorName, cty.DynamicVal, cty.DynamicVal)

			if attr, exists := inner.Attributes["for_each"]; exists {
				// Filter out iterator names inherited from parent blocks
				for _, traversal := range exprFunctions(attr.Expr) {
					if _, inherited := blockIt.Inherited[traversal.RootName()]; !inherited {
						vars = append(vars, traversal)
					}
				}
			}
			if attr, exists := inner.Attributes["labels"]; exists {
				// Filter out both our own iterator name _and_ those inherited
				// from parent blocks, since we provide _both_ of these to the
				// label expressions.
				for _, traversal := range exprFunctions(attr.Expr) {
					ours := traversal.RootName() == iteratorName
					_, inherited := blockIt.Inherited[traversal.RootName()]

					if !(ours || inherited) {
						vars = append(vars, traversal)
					}
				}
			}

			for _, contentBlock := range inner.Blocks {
				// We only request "content" blocks in our schema, so we know
				// any blocks we find here will be content blocks. We require
				// exactly one content block for actual expansion, but we'll
				// be more liberal here so that callers can still collect
				// functions from erroneous "dynamic" blocks.
				children = append(children, WalkFunctionsChild{
					BlockTypeName: blockTypeName,
					Node: WalkFunctionsNode{
						body:           contentBlock.Body,
						it:             blockIt,
						includeContent: n.includeContent,
					},
				})
			}

		default:
			children = append(children, WalkFunctionsChild{
				BlockTypeName: block.Type,
				Node: WalkFunctionsNode{
					body:           block.Body,
					it:             n.it,
					includeContent: n.includeContent,
				},
			})

		}
	}

	return vars, children
}

func (n WalkFunctionsNode) extendSchema(schema *hcl.BodySchema) *hcl.BodySchema {
	// We augment the requested schema to also include our special "dynamic"
	// block type, since then we'll get instances of it interleaved with
	// all of the literal child blocks we must also include.
	extSchema := &hcl.BodySchema{
		Attributes: schema.Attributes,
		Blocks:     make([]hcl.BlockHeaderSchema, len(schema.Blocks), len(schema.Blocks)+1),
	}
	copy(extSchema.Blocks, schema.Blocks)
	extSchema.Blocks = append(extSchema.Blocks, dynamicBlockHeaderSchema)

	return extSchema
}

// This is a more relaxed schema than what's in schema.go, since we
// want to maximize the amount of functions we can find even if there
// are erroneous blocks.
var functionDetectionInnerSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "for_each",
			Required: false,
		},
		{
			Name:     "labels",
			Required: false,
		},
		{
			Name:     "iterator",
			Required: false,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: "content",
		},
	},
}
