package hcl

import (
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	cases := []struct {
		Input  string
		Output map[string]interface{}
	}{
		{
			"comment.hcl",
			map[string]interface{}{
				"foo": "bar",
			},
		},
	}

	for _, tc := range cases {
		d, err := ioutil.ReadFile(filepath.Join(fixtureDir, tc.Input))
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		actual, err := Parse(string(d))
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		if !reflect.DeepEqual(actual, tc.Output) {
			t.Fatalf("Input: %s\n\nBad: %#v", tc.Input, actual)
		}
	}
}
