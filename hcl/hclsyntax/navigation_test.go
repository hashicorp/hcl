package hclsyntax

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/hcl2/hcl"
)

func TestNavigationContextString(t *testing.T) {
	cfg := `


resource {
}

resource "random_type" {
}

resource "null_resource" "baz" {
  name = "foo"
  boz = {
  	one = "111"
  	two = "22222"
  }
}

data "another" "baz" {
  name = "foo"
  boz = {
  	one = "111"
  	two = "22222"
  }
}
`
	file, diags := ParseConfig([]byte(cfg), "", hcl.Pos{Byte: 0, Line: 1, Column: 1})
	if len(diags) != 0 {
		fmt.Printf("offset %d\n", diags[0].Subject.Start.Byte)
		t.Errorf("Unexpected diagnostics: %s", diags)
	}
	if file == nil {
		t.Fatalf("Got nil file")
	}
	nav := file.Nav.(navigation)

	testCases := []struct {
		Offset int
		Want   string
	}{
		{0, ``},
		{2, ``},
		{4, `resource`},
		{17, `resource "random_type"`},
		{25, `resource "random_type"`},
		{45, `resource "null_resource" "baz"`},
		{142, `data "another" "baz"`},
		{180, `data "another" "baz"`},
		{99999, ``},
	}

	for _, tc := range testCases {
		t.Run(strconv.Itoa(tc.Offset), func(t *testing.T) {
			got := nav.ContextString(tc.Offset)

			if got != tc.Want {
				t.Errorf("wrong result\ngot:  %s\nwant: %s", got, tc.Want)
			}
		})
	}
}

func TestNavigationContextDefRange(t *testing.T) {
	cfg := `


resource {
}

resource "random_type" {
}

resource "null_resource" "baz" {
  name = "foo"
  boz = {
  	one = "111"
  	two = "22222"
  }
}

data "another" "baz" {
  name = "foo"
  boz = {
  	one = "111"
  	two = "22222"
  }
}
`
	file, diags := ParseConfig([]byte(cfg), "", hcl.Pos{Byte: 0, Line: 1, Column: 1})
	if len(diags) != 0 {
		fmt.Printf("offset %d\n", diags[0].Subject.Start.Byte)
		t.Errorf("Unexpected diagnostics: %s", diags)
	}
	if file == nil {
		t.Fatalf("Got nil file")
	}
	nav := file.Nav.(navigation)

	testCases := []struct {
		Offset    int
		WantRange hcl.Range
	}{
		{0, hcl.Range{}},
		{2, hcl.Range{}},
		{4, hcl.Range{Filename: "", Start: hcl.Pos{Line: 4, Column: 1, Byte: 3}, End: hcl.Pos{Line: 4, Column: 11, Byte: 13}}},
		{17, hcl.Range{Filename: "", Start: hcl.Pos{Line: 7, Column: 1, Byte: 17}, End: hcl.Pos{Line: 7, Column: 25, Byte: 41}}},
		{25, hcl.Range{Filename: "", Start: hcl.Pos{Line: 7, Column: 1, Byte: 17}, End: hcl.Pos{Line: 7, Column: 25, Byte: 41}}},
		{45, hcl.Range{Filename: "", Start: hcl.Pos{Line: 10, Column: 1, Byte: 45}, End: hcl.Pos{Line: 10, Column: 33, Byte: 77}}},
		{142, hcl.Range{Filename: "", Start: hcl.Pos{Line: 18, Column: 1, Byte: 142}, End: hcl.Pos{Line: 18, Column: 23, Byte: 164}}},
		{180, hcl.Range{Filename: "", Start: hcl.Pos{Line: 18, Column: 1, Byte: 142}, End: hcl.Pos{Line: 18, Column: 23, Byte: 164}}},
		{99999, hcl.Range{}},
	}

	for _, tc := range testCases {
		t.Run(strconv.Itoa(tc.Offset), func(t *testing.T) {
			got := nav.ContextDefRange(tc.Offset)

			if got != tc.WantRange {
				t.Errorf("wrong range\ngot:  %#v\nwant: %#v", got, tc.WantRange)
			}
		})
	}
}
