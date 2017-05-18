package parser

import (
	"github.com/apparentlymart/go-zcl/zcl"
	"github.com/apparentlymart/go-zcl/zcl/json"
)

// NOTE: This is the public interface for parsing. The actual parsers are
// in other packages alongside this one, with this package just wrapping them
// to provide a unified interface for the caller across all supported formats.

// Parser is the main interface for parsing configuration files. As well as
// parsing files, a parser also retains a registry of all of the files it
// has parsed so that multiple attempts to parse the same file will return
// the same object and so the collected files can be used when printing
// diagnostics.
//
// Any diagnostics for parsing a file are only returned once on the first
// call to parse that file. Callers are expected to collect up diagnostics
// and present them together, so returning diagnostics for the same file
// multiple times would create a confusing result.
type Parser struct {
	files map[string]*zcl.File
}

// ParseJSON parses the given JSON buffer (which is assumed to have been loaded
// from the given filename) and returns the zcl.File object representing it.
func (p *Parser) ParseJSON(src []byte, filename string) (*zcl.File, zcl.Diagnostics) {
	if existing := p.files[filename]; existing != nil {
		return existing, nil
	}

	file, diags := json.Parse(src, filename)
	p.files[filename] = file
	return file, diags
}

// ParseJSONFile reads the given filename and parses it as JSON, similarly to
// ParseJSON. An error diagnostic is returned if the given file cannot be read.
func (p *Parser) ParseJSONFile(filename string) (*zcl.File, zcl.Diagnostics) {
	if existing := p.files[filename]; existing != nil {
		return existing, nil
	}

	file, diags := json.ParseFile(filename)
	p.files[filename] = file
	return file, diags
}

// AddFile allows a caller to record in a parser a file that was parsed some
// other way, thus allowing it to be included in the registry of sources.
func (p *Parser) AddFile(filename string, file *zcl.File) {
	p.files[filename] = file
}

// Sources returns a map from filenames to the raw source code that was
// read from them. This is intended to be used, for example, to print
// diagnostics with contextual information.
//
// The arrays underlying the returned slices should not be modified.
func (p *Parser) Sources() map[string][]byte {
	ret := make(map[string][]byte)
	for fn, f := range p.files {
		ret[fn] = f.Bytes
	}
	return ret
}
