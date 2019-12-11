package main

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/userfunc"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

type specFileContent struct {
	Variables map[string]cty.Value
	Functions map[string]function.Function
	RootSpec  hcldec.Spec
}

var specCtx = &hcl.EvalContext{
	Functions: specFuncs,
}

func loadSpecFile(filename string) (specFileContent, hcl.Diagnostics) {
	file, diags := parser.ParseHCLFile(filename)
	if diags.HasErrors() {
		return specFileContent{RootSpec: errSpec}, diags
	}

	vars, funcs, specBody, declDiags := decodeSpecDecls(file.Body)
	diags = append(diags, declDiags...)

	spec, specDiags := decodeSpecRoot(specBody)
	diags = append(diags, specDiags...)

	return specFileContent{
		Variables: vars,
		Functions: funcs,
		RootSpec:  spec,
	}, diags
}

func decodeSpecDecls(body hcl.Body) (map[string]cty.Value, map[string]function.Function, hcl.Body, hcl.Diagnostics) {
	funcs, body, diags := userfunc.DecodeUserFunctions(body, "function", func() *hcl.EvalContext {
		return specCtx
	})

	content, body, moreDiags := body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type: "variables",
			},
		},
	})
	diags = append(diags, moreDiags...)

	vars := make(map[string]cty.Value)
	for _, block := range content.Blocks {
		// We only have one block type in our schema, so we can assume all
		// blocks are of that type.
		attrs, moreDiags := block.Body.JustAttributes()
		diags = append(diags, moreDiags...)

		for name, attr := range attrs {
			val, moreDiags := attr.Expr.Value(specCtx)
			diags = append(diags, moreDiags...)
			vars[name] = val
		}
	}

	return vars, funcs, body, diags
}

func decodeSpecRoot(body hcl.Body) (hcldec.Spec, hcl.Diagnostics) {
	content, diags := body.Content(specSchemaUnlabelled)

	if len(content.Blocks) == 0 {
		if diags.HasErrors() {
			// If we already have errors then they probably explain
			// why we have no blocks, so we'll skip our additional
			// error message added below.
			return errSpec, diags
		}

		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Missing spec block",
			Detail:   "A spec file must have exactly one root block specifying how to map to a JSON value.",
			Subject:  body.MissingItemRange().Ptr(),
		})
		return errSpec, diags
	}

	if len(content.Blocks) > 1 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Extraneous spec block",
			Detail:   "A spec file must have exactly one root block specifying how to map to a JSON value.",
			Subject:  &content.Blocks[1].DefRange,
		})
		return errSpec, diags
	}

	spec, specDiags := decodeSpecBlock(content.Blocks[0])
	diags = append(diags, specDiags...)
	return spec, diags
}

func decodeSpecBlock(block *hcl.Block) (hcldec.Spec, hcl.Diagnostics) {
	var impliedName string
	if len(block.Labels) > 0 {
		impliedName = block.Labels[0]
	}

	switch block.Type {

	case "object":
		return decodeObjectSpec(block.Body)

	case "array":
		return decodeArraySpec(block.Body)

	case "attr":
		return decodeAttrSpec(block.Body, impliedName)

	case "block":
		return decodeBlockSpec(block.Body, impliedName)

	case "block_list":
		return decodeBlockListSpec(block.Body, impliedName)

	case "block_set":
		return decodeBlockSetSpec(block.Body, impliedName)

	case "block_map":
		return decodeBlockMapSpec(block.Body, impliedName)

	case "block_attrs":
		return decodeBlockAttrsSpec(block.Body, impliedName)

	case "default":
		return decodeDefaultSpec(block.Body)

	case "transform":
		return decodeTransformSpec(block.Body)

	case "literal":
		return decodeLiteralSpec(block.Body)

	default:
		// Should never happen, because the above cases should be exhaustive
		// for our schema.
		var diags hcl.Diagnostics
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid spec block",
			Detail:   fmt.Sprintf("Blocks of type %q are not expected here.", block.Type),
			Subject:  &block.TypeRange,
		})
		return errSpec, diags
	}
}

func decodeObjectSpec(body hcl.Body) (hcldec.Spec, hcl.Diagnostics) {
	content, diags := body.Content(specSchemaLabelled)

	spec := make(hcldec.ObjectSpec)
	for _, block := range content.Blocks {
		propSpec, propDiags := decodeSpecBlock(block)
		diags = append(diags, propDiags...)
		spec[block.Labels[0]] = propSpec
	}

	return spec, diags
}

func decodeArraySpec(body hcl.Body) (hcldec.Spec, hcl.Diagnostics) {
	content, diags := body.Content(specSchemaUnlabelled)

	spec := make(hcldec.TupleSpec, 0, len(content.Blocks))
	for _, block := range content.Blocks {
		elemSpec, elemDiags := decodeSpecBlock(block)
		diags = append(diags, elemDiags...)
		spec = append(spec, elemSpec)
	}

	return spec, diags
}

func decodeAttrSpec(body hcl.Body, impliedName string) (hcldec.Spec, hcl.Diagnostics) {
	type content struct {
		Name     *string        `hcl:"name"`
		Type     hcl.Expression `hcl:"type"`
		Required *bool          `hcl:"required"`
	}

	var args content
	diags := gohcl.DecodeBody(body, nil, &args)
	if diags.HasErrors() {
		return errSpec, diags
	}

	spec := &hcldec.AttrSpec{
		Name: impliedName,
	}

	if args.Required != nil {
		spec.Required = *args.Required
	}
	if args.Name != nil {
		spec.Name = *args.Name
	}

	var typeDiags hcl.Diagnostics
	spec.Type, typeDiags = evalTypeExpr(args.Type)
	diags = append(diags, typeDiags...)

	if spec.Name == "" {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Missing name in attribute spec",
			Detail:   "The name attribute is required, to specify the attribute name that is expected in an input HCL file.",
			Subject:  body.MissingItemRange().Ptr(),
		})
		return errSpec, diags
	}

	return spec, diags
}

func decodeBlockSpec(body hcl.Body, impliedName string) (hcldec.Spec, hcl.Diagnostics) {
	type content struct {
		TypeName *string  `hcl:"block_type"`
		Required *bool    `hcl:"required"`
		Nested   hcl.Body `hcl:",remain"`
	}

	var args content
	diags := gohcl.DecodeBody(body, nil, &args)
	if diags.HasErrors() {
		return errSpec, diags
	}

	spec := &hcldec.BlockSpec{
		TypeName: impliedName,
	}

	if args.Required != nil {
		spec.Required = *args.Required
	}
	if args.TypeName != nil {
		spec.TypeName = *args.TypeName
	}

	nested, nestedDiags := decodeBlockNestedSpec(args.Nested)
	diags = append(diags, nestedDiags...)
	spec.Nested = nested

	return spec, diags
}

func decodeBlockListSpec(body hcl.Body, impliedName string) (hcldec.Spec, hcl.Diagnostics) {
	type content struct {
		TypeName *string  `hcl:"block_type"`
		MinItems *int     `hcl:"min_items"`
		MaxItems *int     `hcl:"max_items"`
		Nested   hcl.Body `hcl:",remain"`
	}

	var args content
	diags := gohcl.DecodeBody(body, nil, &args)
	if diags.HasErrors() {
		return errSpec, diags
	}

	spec := &hcldec.BlockListSpec{
		TypeName: impliedName,
	}

	if args.MinItems != nil {
		spec.MinItems = *args.MinItems
	}
	if args.MaxItems != nil {
		spec.MaxItems = *args.MaxItems
	}
	if args.TypeName != nil {
		spec.TypeName = *args.TypeName
	}

	nested, nestedDiags := decodeBlockNestedSpec(args.Nested)
	diags = append(diags, nestedDiags...)
	spec.Nested = nested

	if spec.TypeName == "" {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Missing block_type in block_list spec",
			Detail:   "The block_type attribute is required, to specify the block type name that is expected in an input HCL file.",
			Subject:  body.MissingItemRange().Ptr(),
		})
		return errSpec, diags
	}

	return spec, diags
}

func decodeBlockSetSpec(body hcl.Body, impliedName string) (hcldec.Spec, hcl.Diagnostics) {
	type content struct {
		TypeName *string  `hcl:"block_type"`
		MinItems *int     `hcl:"min_items"`
		MaxItems *int     `hcl:"max_items"`
		Nested   hcl.Body `hcl:",remain"`
	}

	var args content
	diags := gohcl.DecodeBody(body, nil, &args)
	if diags.HasErrors() {
		return errSpec, diags
	}

	spec := &hcldec.BlockSetSpec{
		TypeName: impliedName,
	}

	if args.MinItems != nil {
		spec.MinItems = *args.MinItems
	}
	if args.MaxItems != nil {
		spec.MaxItems = *args.MaxItems
	}
	if args.TypeName != nil {
		spec.TypeName = *args.TypeName
	}

	nested, nestedDiags := decodeBlockNestedSpec(args.Nested)
	diags = append(diags, nestedDiags...)
	spec.Nested = nested

	if spec.TypeName == "" {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Missing block_type in block_set spec",
			Detail:   "The block_type attribute is required, to specify the block type name that is expected in an input HCL file.",
			Subject:  body.MissingItemRange().Ptr(),
		})
		return errSpec, diags
	}

	return spec, diags
}

func decodeBlockMapSpec(body hcl.Body, impliedName string) (hcldec.Spec, hcl.Diagnostics) {
	type content struct {
		TypeName *string  `hcl:"block_type"`
		Labels   []string `hcl:"labels"`
		Nested   hcl.Body `hcl:",remain"`
	}

	var args content
	diags := gohcl.DecodeBody(body, nil, &args)
	if diags.HasErrors() {
		return errSpec, diags
	}

	spec := &hcldec.BlockMapSpec{
		TypeName: impliedName,
	}

	if args.TypeName != nil {
		spec.TypeName = *args.TypeName
	}
	spec.LabelNames = args.Labels

	nested, nestedDiags := decodeBlockNestedSpec(args.Nested)
	diags = append(diags, nestedDiags...)
	spec.Nested = nested

	if spec.TypeName == "" {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Missing block_type in block_map spec",
			Detail:   "The block_type attribute is required, to specify the block type name that is expected in an input HCL file.",
			Subject:  body.MissingItemRange().Ptr(),
		})
		return errSpec, diags
	}
	if len(spec.LabelNames) < 1 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid block label name list",
			Detail:   "A block_map must have at least one label specified.",
			Subject:  body.MissingItemRange().Ptr(),
		})
		return errSpec, diags
	}

	if hcldec.ImpliedType(spec).HasDynamicTypes() {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid block_map spec",
			Detail:   "A block_map spec may not contain attributes with type 'any'.",
			Subject:  body.MissingItemRange().Ptr(),
		})
	}

	return spec, diags
}

func decodeBlockNestedSpec(body hcl.Body) (hcldec.Spec, hcl.Diagnostics) {
	content, diags := body.Content(specSchemaUnlabelled)

	if len(content.Blocks) == 0 {
		if diags.HasErrors() {
			// If we already have errors then they probably explain
			// why we have no blocks, so we'll skip our additional
			// error message added below.
			return errSpec, diags
		}

		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Missing spec block",
			Detail:   "A block spec must have exactly one child spec specifying how to decode block contents.",
			Subject:  body.MissingItemRange().Ptr(),
		})
		return errSpec, diags
	}

	if len(content.Blocks) > 1 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Extraneous spec block",
			Detail:   "A block spec must have exactly one child spec specifying how to decode block contents.",
			Subject:  &content.Blocks[1].DefRange,
		})
		return errSpec, diags
	}

	spec, specDiags := decodeSpecBlock(content.Blocks[0])
	diags = append(diags, specDiags...)
	return spec, diags
}

func decodeBlockAttrsSpec(body hcl.Body, impliedName string) (hcldec.Spec, hcl.Diagnostics) {
	type content struct {
		TypeName    *string        `hcl:"block_type"`
		ElementType hcl.Expression `hcl:"element_type"`
		Required    *bool          `hcl:"required"`
	}

	var args content
	diags := gohcl.DecodeBody(body, nil, &args)
	if diags.HasErrors() {
		return errSpec, diags
	}

	spec := &hcldec.BlockAttrsSpec{
		TypeName: impliedName,
	}

	if args.Required != nil {
		spec.Required = *args.Required
	}
	if args.TypeName != nil {
		spec.TypeName = *args.TypeName
	}

	var typeDiags hcl.Diagnostics
	spec.ElementType, typeDiags = evalTypeExpr(args.ElementType)
	diags = append(diags, typeDiags...)

	if spec.TypeName == "" {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Missing block_type in block_attrs spec",
			Detail:   "The block_type attribute is required, to specify the block type name that is expected in an input HCL file.",
			Subject:  body.MissingItemRange().Ptr(),
		})
		return errSpec, diags
	}

	return spec, diags
}

func decodeLiteralSpec(body hcl.Body) (hcldec.Spec, hcl.Diagnostics) {
	type content struct {
		Value cty.Value `hcl:"value"`
	}

	var args content
	diags := gohcl.DecodeBody(body, specCtx, &args)
	if diags.HasErrors() {
		return errSpec, diags
	}

	return &hcldec.LiteralSpec{
		Value: args.Value,
	}, diags
}

func decodeDefaultSpec(body hcl.Body) (hcldec.Spec, hcl.Diagnostics) {
	content, diags := body.Content(specSchemaUnlabelled)

	if len(content.Blocks) == 0 {
		if diags.HasErrors() {
			// If we already have errors then they probably explain
			// why we have no blocks, so we'll skip our additional
			// error message added below.
			return errSpec, diags
		}

		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Missing spec block",
			Detail:   "A default block must have at least one nested spec, each specifying a possible outcome.",
			Subject:  body.MissingItemRange().Ptr(),
		})
		return errSpec, diags
	}

	if len(content.Blocks) == 1 && !diags.HasErrors() {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagWarning,
			Summary:  "Useless default block",
			Detail:   "A default block with only one spec is equivalent to using that spec alone.",
			Subject:  &content.Blocks[1].DefRange,
		})
	}

	var spec hcldec.Spec
	for _, block := range content.Blocks {
		candidateSpec, candidateDiags := decodeSpecBlock(block)
		diags = append(diags, candidateDiags...)
		if candidateDiags.HasErrors() {
			continue
		}

		if spec == nil {
			spec = candidateSpec
		} else {
			spec = &hcldec.DefaultSpec{
				Primary: spec,
				Default: candidateSpec,
			}
		}
	}

	return spec, diags
}

func decodeTransformSpec(body hcl.Body) (hcldec.Spec, hcl.Diagnostics) {
	type content struct {
		Result hcl.Expression `hcl:"result"`
		Nested hcl.Body       `hcl:",remain"`
	}

	var args content
	diags := gohcl.DecodeBody(body, nil, &args)
	if diags.HasErrors() {
		return errSpec, diags
	}

	spec := &hcldec.TransformExprSpec{
		Expr:         args.Result,
		VarName:      "nested",
		TransformCtx: specCtx,
	}

	nestedContent, nestedDiags := args.Nested.Content(specSchemaUnlabelled)
	diags = append(diags, nestedDiags...)

	if len(nestedContent.Blocks) != 1 {
		if nestedDiags.HasErrors() {
			// If we already have errors then they probably explain
			// why we have the wrong number of blocks, so we'll skip our
			// additional error message added below.
			return errSpec, diags
		}

		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid transform spec",
			Detail:   "A transform spec block must have exactly one nested spec block.",
			Subject:  body.MissingItemRange().Ptr(),
		})
		return errSpec, diags
	}

	nestedSpec, nestedDiags := decodeSpecBlock(nestedContent.Blocks[0])
	diags = append(diags, nestedDiags...)
	spec.Wrapped = nestedSpec

	return spec, diags
}

var errSpec = &hcldec.LiteralSpec{
	Value: cty.NullVal(cty.DynamicPseudoType),
}

var specBlockTypes = []string{
	"object",
	"array",

	"literal",

	"attr",

	"block",
	"block_list",
	"block_map",
	"block_set",

	"default",
	"transform",
}

var specSchemaUnlabelled *hcl.BodySchema
var specSchemaLabelled *hcl.BodySchema

var specSchemaLabelledLabels = []string{"key"}

func init() {
	specSchemaLabelled = &hcl.BodySchema{
		Blocks: make([]hcl.BlockHeaderSchema, 0, len(specBlockTypes)),
	}
	specSchemaUnlabelled = &hcl.BodySchema{
		Blocks: make([]hcl.BlockHeaderSchema, 0, len(specBlockTypes)),
	}

	for _, name := range specBlockTypes {
		specSchemaLabelled.Blocks = append(
			specSchemaLabelled.Blocks,
			hcl.BlockHeaderSchema{
				Type:       name,
				LabelNames: specSchemaLabelledLabels,
			},
		)
		specSchemaUnlabelled.Blocks = append(
			specSchemaUnlabelled.Blocks,
			hcl.BlockHeaderSchema{
				Type: name,
			},
		)
	}
}
