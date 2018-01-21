package dynblock

import (
	"github.com/hashicorp/hcl2/hcl"
)

// ForEachVariables looks for "dynamic" blocks inside the given body
// (which should be a body that would be passed to Expand, not the return
// value of Expand) and returns any variables that are used within their
// "for_each" and "labels" expressions, for use in dynamically constructing a
// scope to pass as part of a hcl.EvalContext to Transformer.
func ForEachVariables(original hcl.Body) []hcl.Traversal {
	var traversals []hcl.Traversal
	container, _, _ := original.PartialContent(variableDetectionContainerSchema)
	if container == nil {
		return traversals
	}

	for _, block := range container.Blocks {
		inner, _, _ := block.Body.PartialContent(variableDetectionInnerSchema)
		if inner == nil {
			continue
		}
		iteratorName := block.Labels[0]
		if attr, exists := inner.Attributes["iterator"]; exists {
			iterTraversal, _ := hcl.AbsTraversalForExpr(attr.Expr)
			if len(iterTraversal) > 0 {
				iteratorName = iterTraversal.RootName()
			}
		}

		if attr, exists := inner.Attributes["for_each"]; exists {
			traversals = append(traversals, attr.Expr.Variables()...)
		}
		if attr, exists := inner.Attributes["labels"]; exists {
			// Filter out our own iterator name, since the caller
			// doesn't need to provide that.
			for _, traversal := range attr.Expr.Variables() {
				if traversal.RootName() != iteratorName {
					traversals = append(traversals, traversal)
				}
			}
		}
	}

	return traversals
}

// These are more-relaxed schemata than what's in schema.go, since we
// want to maximize the amount of variables we can find even if there
// are erroneous blocks.
var variableDetectionContainerSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		dynamicBlockHeaderSchema,
	},
}
var variableDetectionInnerSchema = &hcl.BodySchema{
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
}
