package include

import (
	"github.com/zclconf/go-zcl/ext/transform"
	"github.com/zclconf/go-zcl/gozcl"
	"github.com/zclconf/go-zcl/zcl"
)

// Transformer builds a transformer that finds any "include" blocks in a body
// and produces a merged body that contains the original content plus the
// content of the other bodies referenced by the include blocks.
//
// blockType specifies the type of block to interpret. The conventional type name
// is "include".
//
// ctx provides an evaluation context for the path expressions in include blocks.
// If nil, path expressions may not reference variables nor functions.
//
// The given resolver is used to translate path strings (after expression
// evaluation) into bodies. FileResolver returns a reasonable implementation for
// applications that read configuration files from local disk.
//
// The returned Transformer can either be used directly to process includes
// in a shallow fashion on a single body, or it can be used with
// transform.Deep (from the sibling transform package) to allow includes
// at all levels of a nested block structure:
//
//    transformer = include.Transformer("include", nil, include.FileResolver(".", parser))
//    body = transform.Deep(body, transformer)
//    // "body" will now have includes resolved in its own content and that
//    // of any descendent blocks.
//
func Transformer(blockType string, ctx *zcl.EvalContext, resolver Resolver) transform.Transformer {
	return &transformer{
		Schema: &zcl.BodySchema{
			Blocks: []zcl.BlockHeaderSchema{
				{
					Type: blockType,
				},
			},
		},
		Ctx:      ctx,
		Resolver: resolver,
	}
}

type transformer struct {
	Schema   *zcl.BodySchema
	Ctx      *zcl.EvalContext
	Resolver Resolver
}

func (t *transformer) TransformBody(in zcl.Body) zcl.Body {
	content, remain, diags := in.PartialContent(t.Schema)

	if content == nil || len(content.Blocks) == 0 {
		// Nothing to do!
		return transform.BodyWithDiagnostics(remain, diags)
	}

	bodies := make([]zcl.Body, 1, len(content.Blocks)+1)
	bodies[0] = remain // content in "remain" takes priority over includes
	for _, block := range content.Blocks {
		incContent, incDiags := block.Body.Content(includeBlockSchema)
		diags = append(diags, incDiags...)
		if incDiags.HasErrors() {
			continue
		}

		pathExpr := incContent.Attributes["path"].Expr
		var path string
		incDiags = gozcl.DecodeExpression(pathExpr, t.Ctx, &path)
		diags = append(diags, incDiags...)
		if incDiags.HasErrors() {
			continue
		}

		incBody, incDiags := t.Resolver.ResolveBodyPath(path, pathExpr.Range())
		bodies = append(bodies, transform.BodyWithDiagnostics(incBody, incDiags))
	}

	return zcl.MergeBodies(bodies)
}

var includeBlockSchema = &zcl.BodySchema{
	Attributes: []zcl.AttributeSchema{
		{
			Name:     "path",
			Required: true,
		},
	},
}
