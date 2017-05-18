package json

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/apparentlymart/go-zcl/zcl"
)

// Parse attempts to parse the given buffer as JSON and, if successful, returns
// a zcl.File for the zcl configuration represented by it.
//
// This is not a generic JSON parser. Instead, it deals only with the profile
// of JSON used to express zcl configuration.
//
// The returned file is valid if it is non-nil, regardless of whether the
// diagnostics are also non-nil. If both are returned, the diagnostics should
// still be presented to the user because they may contain warnings.
func Parse(src []byte, filename string) (*zcl.File, zcl.Diagnostics) {
	var file *zcl.File
	rootNode, diags := parseFileContent(src, filename)
	if _, ok := rootNode.(*objectVal); !ok {
		return nil, diags.Append(&zcl.Diagnostic{
			Severity: zcl.DiagError,
			Summary:  "Root value must be object",
			Detail:   "The root value in a JSON-based configuration must be a JSON object.",
			Subject:  rootNode.StartRange().Ptr(),
		})
	}
	if rootNode != nil {
		file = &zcl.File{
			Body: &body{
				obj: rootNode.(*objectVal),
			},
			Bytes: src,
		}
	}
	return file, diags
}

// ParseFile is a convenience wrapper around Parse that first attempts to load
// data from the given filename, passing the result to Parse if successful.
//
// If the file cannot be read, an error diagnostic with nil context is returned.
func ParseFile(filename string) (*zcl.File, zcl.Diagnostics) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, zcl.Diagnostics{
			{
				Severity: zcl.DiagError,
				Summary:  "Failed to open file",
				Detail:   fmt.Sprintf("The file %q could not be opened.", filename),
			},
		}
	}
	defer f.Close()

	src, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, zcl.Diagnostics{
			{
				Severity: zcl.DiagError,
				Summary:  "Failed to read file",
				Detail:   fmt.Sprintf("The file %q was opened, but an error occured while reading it.", filename),
			},
		}
	}

	return Parse(src, filename)
}
