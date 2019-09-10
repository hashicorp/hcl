package hcl

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

var simpleParsers = map[string]func(filename string, src []byte) (*File, Diagnostics){}

// SimpleParse is an easy entry-point into HCL parsing, intended for use in
// applications with straightforward configuratin parsing needs.
//
// SimpleParse uses the filename suffix of the given filename to select a
// syntax that has been registered using RegisterSimpleParser. Two special
// packages in this module automatically register the .hcl and .json suffixes
// when imported, so you can add support for those using two anonymous imports:
//
//     import (
//         // Treat .hcl suffix as HCL native syntax
//         _ "github.com/hashicorp/hcl/v2/hclparse/nativesyntax"
//
//         // Treat .json suffix as HCL JSON
//         _ "github.com/hashicorp/hcl/v2/hclparse/jsonsyntax"
//     )
//
// Once you've successfully parsed data into a *File value, you can use its
// Body field with the decoders in one of the following packages:
//
//     github.com/hashicorp/hcl/v2/gohcl  - To decode into Go struct values
//     github.com/hashicorp/hcl/v2/hcldec - To decode into a dynamic value representation
//
// SimpleParse returns an error value if parsing produces any error diagnostics.
// If error is non-nil then you can optionally type-assert it to
// hcl.Diagnostics to access the individual diagnostics.
//
// Because this function relies on globally-registered parsers, it's not
// appropriate for use from libraries that might be embedded in other
// applications. For library use-cases, or any other situation where a global
// registry is inappropriate, use the API in the package
// github.com/hashicorp/hcl/v2/hclparse instead.
func SimpleParse(filename string, src []byte) (*File, error) {
	suffix := filepath.Ext(filename)
	parser, ok := simpleParsers[suffix]
	if !ok {
		var diags Diagnostics
		diags = diags.Append(&Diagnostic{
			Severity: DiagError,
			Summary:  "Unsupported file format",
			Detail:   fmt.Sprintf("Cannot read from %s: unrecognized file format suffix %q.", filename, suffix),
		})
		return nil, diags
	}
	f, diags := parser(filename, src)
	var err error
	if diags.HasErrors() {
		err = diags
	}
	return f, err
}

// SimpleParseFile is a variant of SimpleParse that first reads the data from
// the given file, and then parses it as per SimpleParse.
//
// The same usage caveats as for SimpleParse apply: use this function only
// from a top-level application, not from a library.
//
// SimpleParseFile returns an error value if reading fails or if parsing
// produces any error diagnostics. If error is non-nil then you can optionally
// type-assert it to hcl.Diagnostics to access the individual diagnostics.
func SimpleParseFile(filename string) (*File, error) {
	src, err := ioutil.ReadFile(filename)
	if err != nil {
		var diags Diagnostics
		if os.IsNotExist(err) {
			diags = diags.Append(&Diagnostic{
				Severity: DiagError,
				Summary:  "File not found",
				Detail:   fmt.Sprintf("There is no file at %s.", filename),
			})
		} else {
			diags = diags.Append(&Diagnostic{
				Severity: DiagError,
				Summary:  "Failed to read file",
				Detail:   fmt.Sprintf("Could not read %s: %s.", filename, err),
			})
		}
		return nil, diags
	}
	return SimpleParse(filename, src)
}

// RegisterSimpleParser can be called from an init() function to associate a
// new parser with a particular filename suffix for use with functions
// SimpleParse and SimpleParseFile.
//
// The suffix must be a string of at least two characters where the first
// character is "." and the remainder are file suffix characters that will
// select the given parser function.
//
// If the given suffix string is invalid or if there's already a parser
// registered for it then this function will panic. It's the responsibility
// of the main application to either directly register or arrange for other
// packages to register the suffixes and parsers it needs, and to coordinate
// to ensure that all registered suffixes are unique.
//
// This function is not thread safe. It should be called _only_ during init(),
// and not from concurrent code. The main way to interact with this function
// is to anonymous-import a package representing a particular syntax
func RegisterSimpleParser(suffix string, fn func(filename string, src []byte) (*File, Diagnostics)) {
	if len(suffix) < 2 || suffix[0] != '.' {
		panic(fmt.Sprintf("invalid filename suffix %q", suffix))
	}
	if _, exists := simpleParsers[suffix]; exists {
		panic(fmt.Sprintf("filename suffix %q is already registered", suffix))
	}
	simpleParsers[suffix] = fn
}
