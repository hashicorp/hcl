package hcl

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	cases := []struct {
		Name string
		Err  bool
	}{
		{
			"comment.hcl",
			false,
		},
		{
			"multiple.hcl",
			false,
		},
		{
			"structure.hcl",
			false,
		},
		{
			"structure_basic.hcl",
			false,
		},
		{
			"structure_empty.hcl",
			false,
		},
		{
			"assign_deep.hcl",
			false,
		},
		{
			"complex.hcl",
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

/*
	cases := []struct {
		Input  string
		Output *ObjectNode
	}{
		{
			"comment.hcl",
			&ObjectNode{
				Key: "",
				Elem: []Node{
					AssignmentNode{
						Key: "foo",
						Value: LiteralNode{
							Type:  ValueTypeString,
							Value: "bar",
						},
					},
				},
			},
		},
		/*
		{
			"multiple.hcl",
			&ObjectNode{
				Elem: map[string][]Node{
					"foo": []Node{
						ValueNode{
							Type:  ValueTypeString,
							Value: "bar",
						},
					},
					"key": []Node{
						ValueNode{
							Type:  ValueTypeInt,
							Value: 7,
						},
					},
				},
			},
		},
		/*
			{
				"structure_basic.hcl",
				&ObjectNode{
					Elem: map[string][]Node{
						"foo": []Node{
							ObjectNode{
								Elem: map[string][]Node{
									"value": []Node{
										ValueNode{
											Type:  ValueTypeInt,
											Value: 7,
										},
									},
								},
							},
						},
					},
				},
			},
			{
				"structure.hcl",
				&ObjectNode{
					Elem: map[string][]Node{
						"foo": []Node{
							ObjectNode{
								Elem: map[string][]Node{
									"bar": []Node{
										ObjectNode{
											Elem: map[string][]Node{
												"baz": []Node{
													ObjectNode{
														Elem: map[string][]Node{
															"key": []Node{
																ValueNode{
																	Type:  ValueTypeInt,
																	Value: 7,
																},
															},

															"foo": []Node{
																ValueNode{
																	Type:  ValueTypeString,
																	Value: "bar",
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		{
			"complex.hcl",
			nil,
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

		if tc.Output != nil {
			if !reflect.DeepEqual(actual, tc.Output) {
				t.Fatalf("Input: %s\n\nBad: %#v", tc.Input, actual)
			}
		}
	}
}
*/
