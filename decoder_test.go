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
		{
			"structure.hcl",
			false,
			map[string]interface{}{
				"foo": []interface{}{
					map[string]interface{}{
						"baz": []interface{}{
							map[string]interface{}{
								"key": 7,
								"foo": "bar",
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		d, err := ioutil.ReadFile(filepath.Join(fixtureDir, tc.File))
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		var out map[string]interface{}
		err = Decode(&out, string(d))
		if (err != nil) != tc.Err {
			t.Fatalf("Input: %s\n\nError: %s", tc.File, err)
		}

		if !reflect.DeepEqual(out, tc.Out) {
			t.Fatalf("Input: %s\n\n%#v", tc.File, out)
		}
	}
}
