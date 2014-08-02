package hcl

import (
	"fmt"

	"github.com/hashicorp/hcl/ast"
	"github.com/hashicorp/hcl/hcl"
	"github.com/hashicorp/hcl/json"
)

// Parse parses the given input and returns the root of the AST.
//
// The input format can be either HCL or JSON.
func Parse(input string) (*ast.ObjectNode, error) {
	switch lexMode(input) {
	case lexModeHcl:
		return hcl.Parse(input)
	case lexModeJson:
		return json.Parse(input)
	}

	return nil, fmt.Errorf("unknown config format")
}
