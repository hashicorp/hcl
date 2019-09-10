// Package jsonsyntax exists only to register the ".json" suffix as using
// JSON syntax when parsing HCL configuration using the hcl.SimpleParse
// or hcl.SimpleParseFile functions.
//
// To enable treating ".json" files as HCL JSON, import this package using
// an anonymous import:
//
//     import _ "github.com/hashicorp/hcl/v2/hclparse/jsonsyntax"
package jsonsyntax

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/json"
)

func init() {
	hcl.RegisterSimpleParser(".json", parseJSONFile)
}

func parseJSONFile(filename string, src []byte) (*hcl.File, hcl.Diagnostics) {
	return json.Parse(src, filename)
}
