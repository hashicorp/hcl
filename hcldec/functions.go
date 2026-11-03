// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hcldec

import (
	"github.com/hashicorp/hcl/v2"
)

// This is based off of hcldec/variables.go

// Functions processes the given body with the given spec and returns a
// list of the function traversals that would be required to decode
// the same pairing of body and spec.
//
// This can be used to conditionally populate the functions in the EvalContext
// passed to Decode, for applications where a static scope is insufficient.
//
// If the given body is not compliant with the given schema, the result may
// be incomplete, but that's assumed to be okay because the eventual call
// to Decode will produce error diagnostics anyway.
func Functions(body hcl.Body, spec Spec) []hcl.Traversal {
	var funcs []hcl.Traversal
	schema := ImpliedSchema(spec)
	content, _, _ := body.PartialContent(schema)

	if vs, ok := spec.(specNeedingFunctions); ok {
		funcs = append(funcs, vs.functionsNeeded(content)...)
	}

	var visitFn visitFunc
	visitFn = func(s Spec) {
		if vs, ok := s.(specNeedingFunctions); ok {
			funcs = append(funcs, vs.functionsNeeded(content)...)
		}
		s.visitSameBodyChildren(visitFn)
	}
	spec.visitSameBodyChildren(visitFn)

	return funcs
}
