package fuzztemplate

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func Fuzz(data []byte) int {
	_, diags := hclsyntax.ParseTemplate(data, "<fuzz-tmpl>", hcl.Pos{Line: 1, Column: 1})

	if diags.HasErrors() {
		return 0
	}

	return 1
}
