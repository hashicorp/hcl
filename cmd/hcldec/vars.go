package main

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

func parseVarsArg(src string, argIdx int) (map[string]cty.Value, hcl.Diagnostics) {
	fakeFn := fmt.Sprintf("<vars argument %d>", argIdx)
	f, diags := parser.ParseJSON([]byte(src), fakeFn)
	if f == nil {
		return nil, diags
	}
	vals, valsDiags := parseVarsBody(f.Body)
	diags = append(diags, valsDiags...)
	return vals, diags
}

func parseVarsFile(filename string) (map[string]cty.Value, hcl.Diagnostics) {
	var f *hcl.File
	var diags hcl.Diagnostics

	if strings.HasSuffix(filename, ".json") {
		f, diags = parser.ParseJSONFile(filename)
	} else {
		f, diags = parser.ParseHCLFile(filename)
	}

	if f == nil {
		return nil, diags
	}

	vals, valsDiags := parseVarsBody(f.Body)
	diags = append(diags, valsDiags...)
	return vals, diags

}

func parseVarsBody(body hcl.Body) (map[string]cty.Value, hcl.Diagnostics) {
	attrs, diags := body.JustAttributes()
	if attrs == nil {
		return nil, diags
	}

	vals := make(map[string]cty.Value, len(attrs))
	for name, attr := range attrs {
		val, valDiags := attr.Expr.Value(nil)
		diags = append(diags, valDiags...)
		vals[name] = val
	}
	return vals, diags
}

// varSpecs is an implementation of pflag.Value that accumulates a list of
// raw values, ignoring any quoting. This is similar to pflag.StringSlice
// but does not complain if there are literal quotes inside the value, which
// is important for us to accept JSON literals here.
type varSpecs []string

func (vs *varSpecs) String() string {
	return strings.Join([]string(*vs), ", ")
}

func (vs *varSpecs) Set(new string) error {
	*vs = append(*vs, new)
	return nil
}

func (vs *varSpecs) Type() string {
	return "json-or-file"
}
