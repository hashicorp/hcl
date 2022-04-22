//go:build go1.18
// +build go1.18

package gohcl

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func TestDecodeWithRange(t *testing.T) {
	type Config struct {
		Name   WithRange[string] `hcl:"name"`
		Number int               `hcl:"number"`
	}

	configSrc := `
		name   = "Gerald"
		number = 12
	`

	f, diags := hclsyntax.ParseConfig([]byte(configSrc), "test.hcl", hcl.InitialPos)
	if diags.HasErrors() {
		t.Fatalf("unexpected errors: %s", diags)
	}

	var config Config
	diags = DecodeBody(f.Body, nil, &config)
	if diags.HasErrors() {
		t.Fatalf("unexpected errors: %s", diags)
	}

	want := Config{
		Name: WithRange[string]{
			Value: "Gerald",
			Range: hcl.Range{
				Filename: "test.hcl",
				Start:    hcl.Pos{ Line: 2, Column: 12, Byte: 12 },
				End:      hcl.Pos{ Line: 2, Column: 20, Byte: 20 },
			},
		},
		Number: 12,
	}
	if diff := cmp.Diff(want, config); diff != "" {
		t.Errorf("incorrect result\n%s", diff)
	}
}
