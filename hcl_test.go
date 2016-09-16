package hcl

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

// This is the directory where our test fixtures are.
const fixtureDir = "./test-fixtures"

func testReadFile(t *testing.T, n string) string {
	d, err := ioutil.ReadFile(filepath.Join(fixtureDir, n))
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	return string(d)
}

type DataTest struct {
	Matches []*MatchTest `json:"match" hcl:"match"`
}
type MatchTest struct {
	Type    string       `json:"type" hcl:"type"`
	Value   string       `json:"value" hcl:"value"`
	Matches []*MatchTest `json:"match,omitempty" hcl:"match"`
}

// Verify that input HCL is decoded equicalently from HCL and JSON. The Encoded
// representation doesn't necessarily matter, as long as the data structures
// are the same
func TestHCL_JSON_roundTrip(t *testing.T) {
	testCases := []struct {
		Name     string
		HCL      string
		FromHCL  interface{}
		FromJSON interface{}
	}{
		{
			Name: "basic",
			HCL: `
entryA "typeA" {}

entryB "typeB" "identB" {
  field1 = "string 1"
  field2 = "string 2"
}`,
		},

		/// from the existing test fixtures
		{
			Name: "basic.hcl",
		},
		{
			Name: "basic_squish.hcl",
		},
		{
			Name: "empty.hcl",
		},
		{
			Name: "tfvars.hcl",
		},
		{
			Name: "escape.hcl",
		},
		{
			Name: "float.hcl",
		},
		{
			Name: "multiline_literal.hcl",
		},
		{
			Name: "multiline.hcl",
		},
		{
			Name: "multiline_indented.hcl",
		},
		{
			Name: "multiline_no_hanging_indent.hcl",
		},
		{
			Name: "multiline_no_eof.hcl",
		},
		{
			Name: "scientific.hcl",
		},
		{
			Name: "terraform_heroku.hcl",
		},
		{
			Name: "structure_multi.hcl",
		},
		{
			Name: "list_of_maps.hcl",
		},
		{
			Name: "assign_deep.hcl",
		},
		{
			Name: "structure_list.hcl",
		},
		{
			Name: "nested_block_comment.hcl",
		},

		// things get more complicated when you add tags to define the data structure
		{
			Name: "GH-141",
			HCL: `
match "first" {
    type = "test"
    value = "first"
}
match "second" {
    type = "test"
    value = "second"
    match "inner" {
        type = "test"
        value = "third"
    }
}
`,
			FromHCL:  &DataTest{},
			FromJSON: &DataTest{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.FromHCL == nil {
				tc.FromHCL = new(interface{})
			}

			if tc.FromJSON == nil {
				tc.FromJSON = new(interface{})
			}

			if tc.HCL == "" {
				tc.HCL = testReadFile(t, tc.Name)
			}

			err := Decode(tc.FromHCL, tc.HCL)
			if err != nil {
				t.Fatal(err)
			}

			fmt.Println("FROM HCL:", spew.Sdump(tc.FromHCL))

			js, err := json.MarshalIndent(tc.FromHCL, "", "  ")
			if err != nil {
				t.Fatal(err)
			}

			fmt.Println("JSON\n", string(js))

			err = Decode(tc.FromJSON, string(js))
			if err != nil {
				t.Fatal(err)
			}

			fmt.Println("FROM JSON:", spew.Sdump(tc.FromJSON))

			if !reflect.DeepEqual(tc.FromHCL, tc.FromJSON) {
				t.Fatalf("DIFF\nFrom HCL: %s\nFrom JSON: %s",
					spew.Sdump(tc.FromHCL), spew.Sdump(tc.FromJSON))
			}
		})
	}
}
