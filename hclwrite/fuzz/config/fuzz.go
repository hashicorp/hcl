package fuzzconfig

import (
	"io/ioutil"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

func Fuzz(data []byte) int {
	file, diags := hclwrite.ParseConfig(data, "<fuzz-conf>", hcl.Pos{Line: 1, Column: 1})

	if diags.HasErrors() {
		return 0
	}

	_, err := file.WriteTo(ioutil.Discard)

	if err != nil {
		return 0
	}

	return 1
}
