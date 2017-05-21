package hclhil

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/apparentlymart/go-zcl/zcl"
)

// Parse attempts to parse the given buffer as HCL with HIL expressions and,
// if successful, returns a zcl.File for the zcl configuration represented by
// it.
//
// The returned file is valid only if the returned diagnostics returns false
// from its HasErrors method. If HasErrors returns true, the file represents
// the subset of data that was able to be parsed, which may be none.
func Parse(src []byte, filename string) (*zcl.File, zcl.Diagnostics) {
	return parse(src, filename)
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
