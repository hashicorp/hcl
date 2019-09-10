// Package nativesyntax exists only to register the ".hcl" suffix as using
// HCL native syntax when parsing HCL configuration using the hcl.SimpleParse
// or hcl.SimpleParseFile functions.
//
// To enable treating ".hcl" files as HCL native syntax, import this package
// using an anonymous import:
//
//     import _ "github.com/hashicorp/hcl/v2/hclparse/nativesyntax"
package nativesyntax

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func init() {
	hcl.RegisterSimpleParser(".hcl", parseNativeSyntaxFile)
}

func parseNativeSyntaxFile(filename string, src []byte) (*hcl.File, hcl.Diagnostics) {
	return hclsyntax.ParseConfig(src, filename, hcl.Pos{Line: 1, Column: 1})
}
