package hclsimple_test

import (
	"fmt"
	"log"
	"os"

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
		fmt.Fprintf(os.Stderr, "Failed: %s\n", err)
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
		fmt.Fprintf(os.Stderr, "Failed: %s\n", err)
		log.Fatalf("Failed to load configuration: %s", err)
	}
	fmt.Printf("Configuration is %v\n", config)

	// Output:
	// Configuration is {bar boop}
}
