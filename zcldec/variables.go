package zcldec

import (
	"github.com/zclconf/go-zcl/zcl"
)

// Variables processes the given body with the given spec and returns a
// list of the variable traversals that would be required to decode
// the same pairing of body and spec.
//
// This can be used to conditionally populate the variables in the EvalContext
// passed to Decode, for applications where a static scope is insufficient.
//
// If the given body is not compliant with the given schema, diagnostics are
// returned describing the problem, which could also serve as a pre-evaluation
// partial validation step.
func Variables(body zcl.Body, spec Spec) ([]zcl.Traversal, zcl.Diagnostics) {
	schema := ImpliedSchema(spec)

	content, _, diags := body.PartialContent(schema)

	var vars []zcl.Traversal
	if diags.HasErrors() {
		return vars, diags
	}

	if vs, ok := spec.(specNeedingVariables); ok {
		vars = append(vars, vs.variablesNeeded(content)...)
	}
	spec.visitSameBodyChildren(func(s Spec) {
		if vs, ok := s.(specNeedingVariables); ok {
			vars = append(vars, vs.variablesNeeded(content)...)
		}
	})

	return vars, diags
}
