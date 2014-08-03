package hcl

import (
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDecode(t *testing.T) {
	cases := []struct {
		File string
		Err  bool
		Out  interface{}
	}{
		{
			"basic.hcl",
			false,
			map[string]interface{}{
				"foo": "bar",
			},
		},
		/*
			{
				"structure.hcl",
				false,
				map[string]interface{}{
					"foo": []interface{}{
						map[string]interface{}{
							"baz": []interface{}{
								map[string]interface{}{
									"foo": "bar",
									"key": 7,
								},
							},
						},
					},
				},
			},
		*/
	}

	for _, tc := range cases {
		d, err := ioutil.ReadFile(filepath.Join(fixtureDir, tc.File))
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		var out interface{}
		err = Decode(&out, string(d))
		if (err != nil) != tc.Err {
			t.Fatalf("Input: %s\n\nError: %s", tc.File, err)
		}

		if !reflect.DeepEqual(out, tc.Out) {
			t.Fatalf("Input: %s\n\n%#v\n\n%#v", tc.File, out, tc.Out)
		}
	}
}

func TestDecode_equal(t *testing.T) {
	cases := []struct {
		One, Two string
	}{
		{
			"basic.hcl",
			"basic.json",
		},
		{
			"structure.hcl",
			"structure.json",
		},
		{
			"structure.hcl",
			"structure_flat.json",
		},
		{
			"structure_multi.hcl",
			"structure_multi.json",
		},
		{
			"structure2.hcl",
			"structure2.json",
		},
	}

	for _, tc := range cases {
		p1 := filepath.Join(fixtureDir, tc.One)
		p2 := filepath.Join(fixtureDir, tc.Two)

		d1, err := ioutil.ReadFile(p1)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		d2, err := ioutil.ReadFile(p2)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		var i1, i2 interface{}
		err = Decode(&i1, string(d1))
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		err = Decode(&i2, string(d2))
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		if !reflect.DeepEqual(i1, i2) {
			t.Fatalf(
				"%s != %s\n\n%#v\n\n%#v",
				tc.One, tc.Two,
				i1, i2)
		}
	}
}

func TestDecode_flatMap(t *testing.T) {
	var val map[string]map[string]string

	err := Decode(&val, testReadFile(t, "structure_flatmap.hcl"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := map[string]map[string]string{
		"foo": map[string]string{
			"foo": "bar",
			"key": "7",
		},
	}

	if !reflect.DeepEqual(val, expected) {
		t.Fatalf("Actual: %#v\n\nExpected: %#v", val, expected)
	}
}

func TestDecode_structure(t *testing.T) {
	type V struct {
		Key int
		Foo string
	}

	var actual V

	err := Decode(&actual, testReadFile(t, "flat.hcl"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := V{
		Key: 7,
		Foo: "bar",
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("Actual: %#v\n\nExpected: %#v", actual, expected)
	}
}
