package main

import (
	"fmt"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/hashicorp/hcl/v2/gohcl"
)

type TestFile struct {
	Result     cty.Value
	ResultType cty.Type

	ChecksTraversals   bool
	ExpectedTraversals []*TestFileExpectTraversal

	ExpectedDiags []*TestFileExpectDiag

	ResultRange     hcl.Range
	ResultTypeRange hcl.Range
}

type TestFileExpectTraversal struct {
	Traversal hcl.Traversal
	Range     hcl.Range
	DeclRange hcl.Range
}

type TestFileExpectDiag struct {
	Severity  hcl.DiagnosticSeverity
	Range     hcl.Range
	DeclRange hcl.Range
}

func (r *Runner) LoadTestFile(filename string) (*TestFile, hcl.Diagnostics) {
	f, diags := r.parser.ParseHCLFile(filename)
	if diags.HasErrors() {
		return nil, diags
	}

	content, moreDiags := f.Body.Content(testFileSchema)
	diags = append(diags, moreDiags...)
	if moreDiags.HasErrors() {
		return nil, diags
	}

	ret := &TestFile{
		ResultType: cty.DynamicPseudoType,
	}

	if typeAttr, exists := content.Attributes["result_type"]; exists {
		ty, moreDiags := typeexpr.TypeConstraint(typeAttr.Expr)
		diags = append(diags, moreDiags...)
		if !moreDiags.HasErrors() {
			ret.ResultType = ty
		}
		ret.ResultTypeRange = typeAttr.Expr.Range()
	}

	if resultAttr, exists := content.Attributes["result"]; exists {
		resultVal, moreDiags := resultAttr.Expr.Value(nil)
		diags = append(diags, moreDiags...)
		if !moreDiags.HasErrors() {
			resultVal, err := convert.Convert(resultVal, ret.ResultType)
			if err != nil {
				diags = diags.Append(&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid result value",
					Detail:   fmt.Sprintf("The result value does not conform to the given result type: %s.", err),
					Subject:  resultAttr.Expr.Range().Ptr(),
				})
			} else {
				ret.Result = resultVal
			}
		}
		ret.ResultRange = resultAttr.Expr.Range()
	}

	for _, block := range content.Blocks {
		switch block.Type {

		case "traversals":
			if ret.ChecksTraversals {
				// Indicates a duplicate traversals block
				diags = diags.Append(&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Duplicate \"traversals\" block",
					Detail:   fmt.Sprintf("Only one traversals block is expected."),
					Subject:  &block.TypeRange,
				})
				continue
			}
			expectTraversals, moreDiags := r.decodeTraversalsBlock(block)
			diags = append(diags, moreDiags...)
			if !moreDiags.HasErrors() {
				ret.ChecksTraversals = true
				ret.ExpectedTraversals = expectTraversals
			}

		case "diagnostics":
			if len(ret.ExpectedDiags) > 0 {
				// Indicates a duplicate diagnostics block
				diags = diags.Append(&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Duplicate \"diagnostics\" block",
					Detail:   fmt.Sprintf("Only one diagnostics block is expected."),
					Subject:  &block.TypeRange,
				})
				continue
			}
			expectDiags, moreDiags := r.decodeDiagnosticsBlock(block)
			diags = append(diags, moreDiags...)
			ret.ExpectedDiags = expectDiags

		default:
			// Shouldn't get here, because the above cases are exhaustive for
			// our test file schema.
			panic(fmt.Sprintf("unsupported block type %q", block.Type))
		}
	}

	if ret.Result != cty.NilVal && len(ret.ExpectedDiags) > 0 {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Conflicting spec expectations",
			Detail:   "This test spec includes expected diagnostics, so it may not also include an expected result.",
			Subject:  &content.Attributes["result"].Range,
		})
	}

	return ret, diags
}

func (r *Runner) decodeTraversalsBlock(block *hcl.Block) ([]*TestFileExpectTraversal, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	content, moreDiags := block.Body.Content(testFileTraversalsSchema)
	diags = append(diags, moreDiags...)
	if moreDiags.HasErrors() {
		return nil, diags
	}

	var ret []*TestFileExpectTraversal
	for _, block := range content.Blocks {
		// There's only one block type in our schema, so we can assume all
		// blocks are of that type.
		expectTraversal, moreDiags := r.decodeTraversalExpectBlock(block)
		diags = append(diags, moreDiags...)
		if expectTraversal != nil {
			ret = append(ret, expectTraversal)
		}
	}

	return ret, diags
}

func (r *Runner) decodeTraversalExpectBlock(block *hcl.Block) (*TestFileExpectTraversal, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	rng, body, moreDiags := r.decodeRangeFromBody(block.Body)
	diags = append(diags, moreDiags...)

	content, moreDiags := body.Content(testFileTraversalExpectSchema)
	diags = append(diags, moreDiags...)
	if moreDiags.HasErrors() {
		return nil, diags
	}

	var traversal hcl.Traversal
	{
		refAttr := content.Attributes["ref"]
		traversal, moreDiags = hcl.AbsTraversalForExpr(refAttr.Expr)
		diags = append(diags, moreDiags...)
		if moreDiags.HasErrors() {
			return nil, diags
		}
	}

	return &TestFileExpectTraversal{
		Traversal: traversal,
		Range:     rng,
		DeclRange: block.DefRange,
	}, diags
}

func (r *Runner) decodeDiagnosticsBlock(block *hcl.Block) ([]*TestFileExpectDiag, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	content, moreDiags := block.Body.Content(testFileDiagnosticsSchema)
	diags = append(diags, moreDiags...)
	if moreDiags.HasErrors() {
		return nil, diags
	}

	if len(content.Blocks) == 0 {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Empty diagnostics block",
			Detail:   "If a diagnostics block is present, at least one expectation statement (\"error\" or \"warning\" block) must be included.",
			Subject:  &block.TypeRange,
		})
		return nil, diags
	}

	ret := make([]*TestFileExpectDiag, 0, len(content.Blocks))
	for _, block := range content.Blocks {
		rng, remain, moreDiags := r.decodeRangeFromBody(block.Body)
		diags = append(diags, moreDiags...)
		if diags.HasErrors() {
			continue
		}

		// Should have nothing else in the block aside from the range definition.
		_, moreDiags = remain.Content(&hcl.BodySchema{})
		diags = append(diags, moreDiags...)

		var severity hcl.DiagnosticSeverity
		switch block.Type {
		case "error":
			severity = hcl.DiagError
		case "warning":
			severity = hcl.DiagWarning
		default:
			panic(fmt.Sprintf("unsupported block type %q", block.Type))
		}

		ret = append(ret, &TestFileExpectDiag{
			Severity:  severity,
			Range:     rng,
			DeclRange: block.TypeRange,
		})
	}
	return ret, diags
}

func (r *Runner) decodeRangeFromBody(body hcl.Body) (hcl.Range, hcl.Body, hcl.Diagnostics) {
	type RawPos struct {
		Line   int `hcl:"line"`
		Column int `hcl:"column"`
		Byte   int `hcl:"byte"`
	}
	type RawRange struct {
		From   RawPos   `hcl:"from,block"`
		To     RawPos   `hcl:"to,block"`
		Remain hcl.Body `hcl:",remain"`
	}

	var raw RawRange
	diags := gohcl.DecodeBody(body, nil, &raw)

	return hcl.Range{
		// We intentionally omit Filename here, because the test spec doesn't
		// need to specify that explicitly: we can infer it to be the file
		// path we pass to hcldec.
		Start: hcl.Pos{
			Line:   raw.From.Line,
			Column: raw.From.Column,
			Byte:   raw.From.Byte,
		},
		End: hcl.Pos{
			Line:   raw.To.Line,
			Column: raw.To.Column,
			Byte:   raw.To.Byte,
		},
	}, raw.Remain, diags
}

var testFileSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: "result",
		},
		{
			Name: "result_type",
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: "traversals",
		},
		{
			Type: "diagnostics",
		},
	},
}

var testFileTraversalsSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: "expect",
		},
	},
}

var testFileTraversalExpectSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "ref",
			Required: true,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: "range",
		},
	},
}

var testFileDiagnosticsSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: "error",
		},
		{
			Type: "warning",
		},
	},
}

var testFileRangeSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: "from",
		},
		{
			Type: "to",
		},
	},
}

var testFilePosSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "line",
			Required: true,
		},
		{
			Name:     "column",
			Required: true,
		},
		{
			Name:     "byte",
			Required: true,
		},
	},
}
