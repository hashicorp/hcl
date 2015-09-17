package json

import (
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	cases := []struct {
		Name string
		Err  bool
	}{
		{
			"basic.json",
			false,
		},
		{
			"object.json",
			false,
		},
		{
			"array.json",
			false,
		},
		{
			"types.json",
			false,
		},
	}

	for _, tc := range cases {
		d, err := ioutil.ReadFile(filepath.Join(fixtureDir, tc.Name))
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		_, err = Parse(string(d))
		if (err != nil) != tc.Err {
			t.Fatalf("Input: %s\n\nError: %s", tc.Name, err)
		}
	}
}

func TestParseComments(t *testing.T) {
	input, err := ioutil.ReadFile(filepath.Join(fixtureDir, "comments.json"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	actual, err := Parse(string(input))
	if err != nil {
		t.Fatalf("Input: %s\n\nError: %s", "comments.json", err)
	}

	after, err := ioutil.ReadFile(filepath.Join(fixtureDir, "comments-after.json"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expected, err := Parse(string(after))
	if err != nil {
		t.Fatalf("Input: %s\n\nError: %s", "comments-after.json", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf(
			"Input: %s\n\nBad: %#v\n\nExpected: %#v",
			input, actual, expected)
	}
}
