package hclsimple_test

import (
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

func Example_nativeSyntax() {
	type Config struct {
		Foo string `hcl:"foo"`
		Baz string `hcl:"baz"`
	}

	const exampleConfig = `
	foo = "bar"
	baz = "boop"
	`

	var config Config
	err := hclsimple.Decode(
		"example.hcl", []byte(exampleConfig),
		nil, &config,
	)
	if err != nil {
		log.Fatalf("Failed to load configuration: %s", err)
	}
	fmt.Printf("Configuration is %v\n", config)

	// Output:
	// Configuration is {bar boop}
}

func Example_jsonSyntax() {
	type Config struct {
		Foo string `hcl:"foo"`
		Baz string `hcl:"baz"`
	}

	const exampleConfig = `
	{
		"foo": "bar",
		"baz": "boop"
	}
	`

	var config Config
	err := hclsimple.Decode(
		"example.json", []byte(exampleConfig),
		nil, &config,
	)
	if err != nil {
		log.Fatalf("Failed to load configuration: %s", err)
	}
	fmt.Printf("Configuration is %v\n", config)

	// Output:
	// Configuration is {bar boop}
}

func TestDecodeFile(t *testing.T) {
	type Config struct {
		Foo string `hcl:"foo"`
		Baz string `hcl:"baz"`
	}

	var got Config
	err := hclsimple.DecodeFile("testdata/test.hcl", nil, &got)
	if err != nil {
		t.Fatalf("unexpected error(s): %s", err)
	}
	want := Config{
		Foo: "bar",
		Baz: "boop",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, want)
	}
}
