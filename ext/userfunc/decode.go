package userfunc

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty/function"
)

var funcBodySchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "params",
			Required: true,
		},
		{
			Name:     "variadic_param",
			Required: false,
		},
		{
			Name:     "result",
			Required: true,
		},
	},
}

func decodeUserFunctions(body hcl.Body, blockType string, contextFunc ContextFunc) (funcs map[string]function.Function, remain hcl.Body, diags hcl.Diagnostics) {
	schema := &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       blockType,
				LabelNames: []string{"name"},
			},
		},
	}

	content, remain, diags := body.PartialContent(schema)
	if diags.HasErrors() {
		return nil, remain, diags
	}

	// first call to getBaseCtx will populate context, and then the same
	// context will be used for all subsequent calls. It's assumed that
	// all functions in a given body should see an identical context.
	var baseCtx *hcl.EvalContext
	getBaseCtx := func() *hcl.EvalContext {
		if baseCtx == nil {
			if contextFunc != nil {
				baseCtx = contextFunc()
			}
		}
		// baseCtx might still be nil here, and that's okay
		return baseCtx
	}

	funcs = make(map[string]function.Function)

	for _, block := range content.Blocks {
		name := block.Labels[0]
		funcContent, funcDiags := block.Body.Content(funcBodySchema)
		diags = append(diags, funcDiags...)
		if funcDiags.HasErrors() {
			continue
		}

		paramsExpr := funcContent.Attributes["params"].Expr
		resultExpr := funcContent.Attributes["result"].Expr
		var varParamExpr hcl.Expression
		if funcContent.Attributes["variadic_param"] != nil {
			varParamExpr = funcContent.Attributes["variadic_param"].Expr
		}
		f, funcDiags := NewFunction(paramsExpr, varParamExpr, resultExpr, getBaseCtx)
		if funcDiags.HasErrors() {
			diags = append(diags, funcDiags...)
			continue
		}
		funcs[name] = f
	}

	return funcs, remain, diags
}
