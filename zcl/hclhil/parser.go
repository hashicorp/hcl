package hclhil

import (
	"fmt"

	"github.com/hashicorp/hcl"
	hclast "github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl2/zcl"
)

func parse(src []byte, filename string) (*zcl.File, zcl.Diagnostics) {
	hclFile, err := hcl.ParseBytes(src)
	if err != nil {
		return nil, zcl.Diagnostics{
			&zcl.Diagnostic{
				Severity: zcl.DiagError,
				Summary:  "Syntax error in configuration",
				Detail:   fmt.Sprintf("The file %q could not be parsed: %s", filename, err),
				Subject:  errorRange(err),
			},
		}
	}

	return &zcl.File{
		Body: &body{
			oli: hclFile.Node.(*hclast.ObjectList),
		},
	}, nil
}
