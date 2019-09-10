package hcldec

import (
	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2"
)

// SimpleDecode is a helper wrapper around hcl.SimpleParse and hcldec.Decode
// for simple applications that just want to read the contents of a single
// configuration file into a dynamic value in one step.
//
// SimpleDecode uses hcl.SimpleParse's table of registered file suffixes, so
// to read HCL native syntax and HCL JSON files (named with .hcl and .json
// suffixes respectively) you must import the two registration packages into
// your application:
//
//     import (
//         // Treat .hcl suffix as HCL native syntax
//         _ "github.com/hashicorp/hcl/v2/hclparse/nativesyntax"
//
//         // Treat .json suffix as HCL JSON
//         _ "github.com/hashicorp/hcl/v2/hclparse/jsonsyntax"
//     )
//
// The "filename" and "src" arguments are interpreted the same way as for
// hcl.SimpleParse, while the "spec" and "ctx" arguments are interpreted the
// same way as for hcldec.Decode.
//
// SimpleDecode returns a non-nil error if parsing or decoding produce any
// error diagnostics. You can type-assert a non-nil error into hcl.Diagnostics
// to access the individual diagnostic messages.
//
// As with hcl.SimpleParse, this function is not appropriate to use from a
// library because it relies on a single global table of file formats. For
// more complex use-cases, use the API in the "hclparse" package to obtain
// an hcl.Body value and then pass it to hcldec.Decode.
func SimpleDecode(filename string, src []byte, spec Spec, ctx *hcl.EvalContext) (cty.Value, error) {
	f, err := hcl.SimpleParse(filename, src)
	if err != nil {
		return cty.DynamicVal, err
	}

	val, diags := Decode(f.Body, spec, ctx)
	if diags.HasErrors() {
		if val == cty.NilVal {
			val = cty.DynamicVal // Save callers from having to deal with nils
		}
		return val, diags
	}
	return val, nil
}
