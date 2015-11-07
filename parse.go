package hcl

import (
	"fmt"

	"github.com/hashicorp/hcl/hcl/ast"
	hclParser "github.com/hashicorp/hcl/hcl/parser"
	"github.com/hashicorp/hcl/json"
)

// Parse parses the given input and returns the root object.
//
// The input format can be either HCL or JSON.
func Parse(input string) (*ast.File, error) {
	switch lexMode(input) {
	case lexModeHcl:
		return hclParser.Parse([]byte(input))
	case lexModeJson:
		return json.Parse(input)
	}

	return nil, fmt.Errorf("unknown config format")
}
