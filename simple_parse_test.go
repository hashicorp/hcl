package hcl_test

import (
	"fmt"
	"log"

	"github.com/hashicorp/hcl/v2"
	// Also import one or both of these to register the syntaxes you want
	// to support. You'll need the first of these to successfully run this
	// example.
	// "github.com/hashicorp/hcl/v2/hclparse/nativesyntax" // parse .hcl as native syntax
	// "github.com/hashicorp/hcl/v2/hclparse/jsonsyntax"   // parse .json as JSON syntax
)

const simpleParseExampleConfig = `
foo = "bar"
baz = "boop"
`

func ExampleSimpleParse() {
	f, err := hcl.SimpleParse("example.hcl", []byte(simpleParseExampleConfig))
	if err != nil {
		log.Fatalf("Failed: %s", err)
	}

	// SimpleParse returns a hcl.File object, which you can then process in
	// one of two ways. See the following packages for more information:
	// - github.com/hashicorp/hcl/v2/gohcl for decoding directly into Go structs.
	// - github.com/hashicorp/hcl/v2/hcldec for decoding into dynamic values
	//   for further analysis.
	fmt.Printf("Can now pass the file's %T to either gohcl or hcldec!\n", f.Body)

	// See gohcl.SimpleDecode for a way to decode source code into a Go struct
	// value in a single step.
}
