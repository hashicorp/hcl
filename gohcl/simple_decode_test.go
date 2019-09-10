package gohcl_test

import (
	"fmt"
	"log"

	"github.com/hashicorp/hcl/v2/gohcl"

	// These anonymous imports register the parsers for the native syntax and
	// the JSON syntax, selected by filename suffix.
	_ "github.com/hashicorp/hcl/v2/hclparse/jsonsyntax"   // parse .json as JSON syntax
	_ "github.com/hashicorp/hcl/v2/hclparse/nativesyntax" // parse .hcl as native syntax
)

const simpleParseExampleConfigNativeSyntax = `
foo = "bar"
baz = "boop"
`

const simpleParseExampleConfigJSONSyntax = `
{
  "foo": "bar",
  "baz": "boop"
}
`

type simpleDecodeExample struct {
	Foo string `hcl:"foo"`
	Baz string `hcl:"baz"`
}

func ExampleSimpleDecode() {
	var fromNative simpleDecodeExample
	err := gohcl.SimpleDecode(
		"example.hcl", []byte(simpleParseExampleConfigNativeSyntax),
		nil, &fromNative,
	)
	if err != nil {
		log.Fatalf("Failed: %s", err)
	}
	fmt.Printf("From native syntax we got: %v\n", fromNative)

	var fromJSON simpleDecodeExample
	err = gohcl.SimpleDecode(
		"example.json", []byte(simpleParseExampleConfigJSONSyntax),
		nil, &fromJSON,
	)
	if err != nil {
		log.Fatalf("Failed: %s", err)
	}
	fmt.Printf("From JSON syntax we got:   %v\n", fromJSON)

	// Output:
	// From native syntax we got: {bar boop}
	// From JSON syntax we got:   {bar boop}
}
