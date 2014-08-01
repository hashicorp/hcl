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
		Output *ObjectNode
	}{
		{
			"comment.hcl",
			&ObjectNode{
				Elem: map[string][]Node{
					"foo": []Node{
						ValueNode{
							Type:  ValueTypeString,
							Value: "bar",
						},
					},
				},
			},
		},
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
