package hcldec_test

import (
	"fmt"
	"log"

	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2/hcldec"

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

var simpleDecodeExampleSpec = hcldec.ObjectSpec{
	"foo": &hcldec.AttrSpec{Name: "foo", Type: cty.String, Required: true},
	"baz": &hcldec.AttrSpec{Name: "baz", Type: cty.String},
}

func ExampleSimpleDecode() {
	fromNative, err := hcldec.SimpleDecode(
		"example.hcl", []byte(simpleParseExampleConfigNativeSyntax),
		simpleDecodeExampleSpec, nil,
	)
	if err != nil {
		log.Fatalf("Failed: %s", err)
	}
	fmt.Printf("From native syntax we got: %#v\n", fromNative)

	fromJSON, err := hcldec.SimpleDecode(
		"example.json", []byte(simpleParseExampleConfigJSONSyntax),
		simpleDecodeExampleSpec, nil,
	)
	if err != nil {
		log.Fatalf("Failed: %s", err)
	}
	fmt.Printf("From JSON syntax we got:   %#v\n", fromJSON)

	// Output:
	// From native syntax we got: cty.ObjectVal(map[string]cty.Value{"baz":cty.StringVal("boop"), "foo":cty.StringVal("bar")})
	// From JSON syntax we got:   cty.ObjectVal(map[string]cty.Value{"baz":cty.StringVal("boop"), "foo":cty.StringVal("bar")})
}
