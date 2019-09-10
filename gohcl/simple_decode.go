package gohcl

import (
	"github.com/hashicorp/hcl/v2"
)

// SimpleDecode is a helper wrapper around hcl.SimpleParse and gohcl.DecodeBody
// for simple applications that just want to read the contents of a single
// configuration file into a Go struct value in one step.
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
// hcl.SimpleParse, while the "ctx" and "val" arguments are interpreted the
// same way as for gohcl.DecodeBody.
//
// SimpleDecode returns a non-nil error if parsing or decoding produce any
// error diagnostics. You can type-assert a non-nil error into hcl.Diagnostics
// to access the individual diagnostic messages.
//
// As with hcl.SimpleParse, this function is not appropriate to use from a
// library because it relies on a single global table of file formats. For
// more complex use-cases, use the API in the "hclparse" package to obtain
// an hcl.Body value and then pass it to gohcl.DecodeBody.
func SimpleDecode(filename string, src []byte, ctx *hcl.EvalContext, val interface{}) error {
	f, err := hcl.SimpleParse(filename, src)
	if err != nil {
		return err
	}

	diags := DecodeBody(f.Body, ctx, val)
	if diags.HasErrors() {
		return diags
	}
	return nil
}

// SimpleDecodeFile is a variant of SimpleDecode that will first read all of
// the data from the given filename, and then process the contents in the
// same way as SimpleDecode.
//
// See the SimpleDecode function documentation for some prerequisites and
// caveats for this function, and for some alternatives for more complex
// use-cases.
func SimpleDecodeFile(filename string, ctx *hcl.EvalContext, val interface{}) error {
	f, err := hcl.SimpleParseFile(filename)
	if err != nil {
		return err
	}

	diags := DecodeBody(f.Body, ctx, val)
	if diags.HasErrors() {
		return diags
	}
	return nil
}
