package typeexpr

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/zclconf/go-cty/cty"
)

var (
	valueComparer = cmp.Comparer(cty.Value.RawEquals)
)

func TestDefaults_Apply(t *testing.T) {
	simpleObject := cty.ObjectWithOptionalAttrs(map[string]cty.Type{
		"a": cty.String,
		"b": cty.Bool,
	}, []string{"b"})
	nestedObject := cty.ObjectWithOptionalAttrs(map[string]cty.Type{
		"c": simpleObject,
		"d": cty.Number,
	}, []string{"c"})

	testCases := map[string]struct {
		defaults *Defaults
		value    cty.Value
		want     cty.Value
	}{
		// Nothing happens when there are no default values and no children.
		"no defaults": {
			defaults: &Defaults{
				Type: cty.Map(cty.String),
			},
			value: cty.MapVal(map[string]cty.Value{
				"a": cty.StringVal("foo"),
				"b": cty.StringVal("bar"),
			}),
			want: cty.MapVal(map[string]cty.Value{
				"a": cty.StringVal("foo"),
				"b": cty.StringVal("bar"),
			}),
		},
		// Passing a map which does not include one of the attributes with a
		// default results in the default being applied to the output. Output
		// is always an object.
		"simple object with defaults applied": {
			defaults: &Defaults{
				Type: simpleObject,
				DefaultValues: map[string]cty.Value{
					"b": cty.True,
				},
			},
			value: cty.MapVal(map[string]cty.Value{
				"a": cty.StringVal("foo"),
			}),
			want: cty.ObjectVal(map[string]cty.Value{
				"a": cty.StringVal("foo"),
				"b": cty.True,
			}),
		},
		// Unknown values may be assigned to root modules during validation,
		// and we cannot apply defaults at that time.
		"simple object with defaults but unknown value": {
			defaults: &Defaults{
				Type: simpleObject,
				DefaultValues: map[string]cty.Value{
					"b": cty.True,
				},
			},
			value: cty.UnknownVal(cty.Map(cty.String)),
			want:  cty.UnknownVal(cty.Map(cty.String)),
		},
		// Defaults do not override attributes which are present in the given
		// value.
		"simple object with optional attributes specified": {
			defaults: &Defaults{
				Type: simpleObject,
				DefaultValues: map[string]cty.Value{
					"b": cty.True,
				},
			},
			value: cty.MapVal(map[string]cty.Value{
				"a": cty.StringVal("foo"),
				"b": cty.StringVal("false"),
			}),
			want: cty.ObjectVal(map[string]cty.Value{
				"a": cty.StringVal("foo"),
				"b": cty.StringVal("false"),
			}),
		},
		// Defaults will replace explicit nulls.
		"object with explicit null for attribute with default": {
			defaults: &Defaults{
				Type: simpleObject,
				DefaultValues: map[string]cty.Value{
					"b": cty.True,
				},
			},
			value: cty.MapVal(map[string]cty.Value{
				"a": cty.StringVal("foo"),
				"b": cty.NullVal(cty.String),
			}),
			want: cty.ObjectVal(map[string]cty.Value{
				"a": cty.StringVal("foo"),
				"b": cty.True,
			}),
		},
		// Defaults can be specified at any level of depth and will be applied
		// so long as there is a parent value to populate.
		"nested object with defaults applied": {
			defaults: &Defaults{
				Type: nestedObject,
				Children: map[string]*Defaults{
					"c": {
						Type: simpleObject,
						DefaultValues: map[string]cty.Value{
							"b": cty.False,
						},
					},
				},
			},
			value: cty.ObjectVal(map[string]cty.Value{
				"c": cty.ObjectVal(map[string]cty.Value{
					"a": cty.StringVal("foo"),
				}),
				"d": cty.NumberIntVal(5),
			}),
			want: cty.ObjectVal(map[string]cty.Value{
				"c": cty.ObjectVal(map[string]cty.Value{
					"a": cty.StringVal("foo"),
					"b": cty.False,
				}),
				"d": cty.NumberIntVal(5),
			}),
		},
		// Testing traversal of collections.
		"map of objects with defaults applied": {
			defaults: &Defaults{
				Type: cty.Map(simpleObject),
				Children: map[string]*Defaults{
					"": {
						Type: simpleObject,
						DefaultValues: map[string]cty.Value{
							"b": cty.True,
						},
					},
				},
			},
			value: cty.MapVal(map[string]cty.Value{
				"f": cty.ObjectVal(map[string]cty.Value{
					"a": cty.StringVal("foo"),
				}),
				"b": cty.ObjectVal(map[string]cty.Value{
					"a": cty.StringVal("bar"),
				}),
			}),
			want: cty.MapVal(map[string]cty.Value{
				"f": cty.ObjectVal(map[string]cty.Value{
					"a": cty.StringVal("foo"),
					"b": cty.True,
				}),
				"b": cty.ObjectVal(map[string]cty.Value{
					"a": cty.StringVal("bar"),
					"b": cty.True,
				}),
			}),
		},
		// A map variable value specified in a tfvars file will be an object,
		// in which case we must still traverse the defaults structure
		// correctly.
		"map of objects with defaults applied, given object instead of map": {
			defaults: &Defaults{
				Type: cty.Map(simpleObject),
				Children: map[string]*Defaults{
					"": {
						Type: simpleObject,
						DefaultValues: map[string]cty.Value{
							"b": cty.True,
						},
					},
				},
			},
			value: cty.ObjectVal(map[string]cty.Value{
				"f": cty.ObjectVal(map[string]cty.Value{
					"a": cty.StringVal("foo"),
				}),
				"b": cty.ObjectVal(map[string]cty.Value{
					"a": cty.StringVal("bar"),
				}),
			}),
			want: cty.ObjectVal(map[string]cty.Value{
				"f": cty.ObjectVal(map[string]cty.Value{
					"a": cty.StringVal("foo"),
					"b": cty.True,
				}),
				"b": cty.ObjectVal(map[string]cty.Value{
					"a": cty.StringVal("bar"),
					"b": cty.True,
				}),
			}),
		},
		// Another example of a collection type, this time exercising the code
		// processing a tuple input.
		"list of objects with defaults applied": {
			defaults: &Defaults{
				Type: cty.List(simpleObject),
				Children: map[string]*Defaults{
					"": {
						Type: simpleObject,
						DefaultValues: map[string]cty.Value{
							"b": cty.True,
						},
					},
				},
			},
			value: cty.TupleVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"a": cty.StringVal("foo"),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"a": cty.StringVal("bar"),
				}),
			}),
			want: cty.TupleVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"a": cty.StringVal("foo"),
					"b": cty.True,
				}),
				cty.ObjectVal(map[string]cty.Value{
					"a": cty.StringVal("bar"),
					"b": cty.True,
				}),
			}),
		},
		// Unlike collections, tuple variable types can have defaults for
		// multiple element types.
		"tuple of objects with defaults applied": {
			defaults: &Defaults{
				Type: cty.Tuple([]cty.Type{simpleObject, nestedObject}),
				Children: map[string]*Defaults{
					"0": {
						Type: simpleObject,
						DefaultValues: map[string]cty.Value{
							"b": cty.False,
						},
					},
					"1": {
						Type: nestedObject,
						DefaultValues: map[string]cty.Value{
							"c": cty.ObjectVal(map[string]cty.Value{
								"a": cty.StringVal("default"),
								"b": cty.True,
							}),
						},
					},
				},
			},
			value: cty.TupleVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"a": cty.StringVal("foo"),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"d": cty.NumberIntVal(5),
				}),
			}),
			want: cty.TupleVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"a": cty.StringVal("foo"),
					"b": cty.False,
				}),
				cty.ObjectVal(map[string]cty.Value{
					"c": cty.ObjectVal(map[string]cty.Value{
						"a": cty.StringVal("default"),
						"b": cty.True,
					}),
					"d": cty.NumberIntVal(5),
				}),
			}),
		},
		// More complex cases with deeply nested defaults, testing the "default
		// within a default" edges.
		"set of nested objects, no default sub-object": {
			defaults: &Defaults{
				Type: cty.Set(nestedObject),
				Children: map[string]*Defaults{
					"": {
						Type: nestedObject,
						Children: map[string]*Defaults{
							"c": {
								Type: simpleObject,
								DefaultValues: map[string]cty.Value{
									"b": cty.True,
								},
							},
						},
					},
				},
			},
			value: cty.TupleVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"c": cty.ObjectVal(map[string]cty.Value{
						"a": cty.StringVal("foo"),
					}),
					"d": cty.NumberIntVal(5),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"d": cty.NumberIntVal(7),
				}),
			}),
			want: cty.TupleVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"c": cty.ObjectVal(map[string]cty.Value{
						"a": cty.StringVal("foo"),
						"b": cty.True,
					}),
					"d": cty.NumberIntVal(5),
				}),
				cty.ObjectVal(map[string]cty.Value{
					// No default value for "c" specified, so none applied. The
					// convert stage will fill in a null.
					"d": cty.NumberIntVal(7),
				}),
			}),
		},
		"set of nested objects, empty default sub-object": {
			defaults: &Defaults{
				Type: cty.Set(nestedObject),
				Children: map[string]*Defaults{
					"": {
						Type: nestedObject,
						DefaultValues: map[string]cty.Value{
							// This is a convenient shorthand which causes a
							// missing sub-object to be filled with an object
							// with all of the default values specified in the
							// sub-object's type.
							"c": cty.EmptyObjectVal,
						},
						Children: map[string]*Defaults{
							"c": {
								Type: simpleObject,
								DefaultValues: map[string]cty.Value{
									"b": cty.True,
								},
							},
						},
					},
				},
			},
			value: cty.TupleVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"c": cty.ObjectVal(map[string]cty.Value{
						"a": cty.StringVal("foo"),
					}),
					"d": cty.NumberIntVal(5),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"d": cty.NumberIntVal(7),
				}),
			}),
			want: cty.TupleVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"c": cty.ObjectVal(map[string]cty.Value{
						"a": cty.StringVal("foo"),
						"b": cty.True,
					}),
					"d": cty.NumberIntVal(5),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"c": cty.ObjectVal(map[string]cty.Value{
						// Default value for "b" is applied to the empty object
						// specified as the default for "c"
						"b": cty.True,
					}),
					"d": cty.NumberIntVal(7),
				}),
			}),
		},
		"set of nested objects, overriding default sub-object": {
			defaults: &Defaults{
				Type: cty.Set(nestedObject),
				Children: map[string]*Defaults{
					"": {
						Type: nestedObject,
						DefaultValues: map[string]cty.Value{
							// If no value is given for "c", we use this object
							// of non-default values instead. These take
							// precedence over the default values specified in
							// the child type.
							"c": cty.ObjectVal(map[string]cty.Value{
								"a": cty.StringVal("fallback"),
								"b": cty.False,
							}),
						},
						Children: map[string]*Defaults{
							"c": {
								Type: simpleObject,
								DefaultValues: map[string]cty.Value{
									"b": cty.True,
								},
							},
						},
					},
				},
			},
			value: cty.TupleVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"c": cty.ObjectVal(map[string]cty.Value{
						"a": cty.StringVal("foo"),
					}),
					"d": cty.NumberIntVal(5),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"d": cty.NumberIntVal(7),
				}),
			}),
			want: cty.TupleVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"c": cty.ObjectVal(map[string]cty.Value{
						"a": cty.StringVal("foo"),
						"b": cty.True,
					}),
					"d": cty.NumberIntVal(5),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"c": cty.ObjectVal(map[string]cty.Value{
						// The default value for "b" is not applied, as the
						// default value for "c" includes a non-default value
						// already.
						"a": cty.StringVal("fallback"),
						"b": cty.False,
					}),
					"d": cty.NumberIntVal(7),
				}),
			}),
		},
		"set of nested objects, nulls in default sub-object overridden": {
			defaults: &Defaults{
				Type: cty.Set(nestedObject),
				Children: map[string]*Defaults{
					"": {
						Type: nestedObject,
						DefaultValues: map[string]cty.Value{
							// The default value for "c" is used to prepopulate
							// the nested object's value if not specified, but
							// the null default for its "b" attribute will be
							// overridden by the default specified in the child
							// type.
							"c": cty.ObjectVal(map[string]cty.Value{
								"a": cty.StringVal("fallback"),
								"b": cty.NullVal(cty.Bool),
							}),
						},
						Children: map[string]*Defaults{
							"c": {
								Type: simpleObject,
								DefaultValues: map[string]cty.Value{
									"b": cty.True,
								},
							},
						},
					},
				},
			},
			value: cty.TupleVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"c": cty.ObjectVal(map[string]cty.Value{
						"a": cty.StringVal("foo"),
					}),
					"d": cty.NumberIntVal(5),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"d": cty.NumberIntVal(7),
				}),
			}),
			want: cty.TupleVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"c": cty.ObjectVal(map[string]cty.Value{
						"a": cty.StringVal("foo"),
						"b": cty.True,
					}),
					"d": cty.NumberIntVal(5),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"c": cty.ObjectVal(map[string]cty.Value{
						// The default value for "b" overrides the explicit
						// null in the default value for "c".
						"a": cty.StringVal("fallback"),
						"b": cty.True,
					}),
					"d": cty.NumberIntVal(7),
				}),
			}),
		},
		"null objects do not get default values inserted": {
			defaults: &Defaults{
				Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
					"required": cty.String,
					"optional": cty.String,
				}, []string{"optional"}),
				DefaultValues: map[string]cty.Value{
					"optional": cty.StringVal("optional"),
				},
			},
			value: cty.NullVal(cty.Object(map[string]cty.Type{
				"required": cty.String,
				"optional": cty.String,
			})),
			want: cty.NullVal(cty.Object(map[string]cty.Type{
				"required": cty.String,
				"optional": cty.String,
			})),
		},
		"defaults with unset defaults are still applied (null)": {
			defaults: &Defaults{
				Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
					"required": cty.String,
					"optional_object": cty.ObjectWithOptionalAttrs(map[string]cty.Type{
						"nested_required": cty.String,
						"nested_optional": cty.String,
					}, []string{"nested_optional"}),
				}, []string{"optional_object"}),
				DefaultValues: map[string]cty.Value{
					"optional_object": cty.ObjectVal(map[string]cty.Value{
						"nested_required": cty.StringVal("required"),
						"nested_optional": cty.NullVal(cty.String),
					}),
				},
				Children: map[string]*Defaults{
					"optional_object": {
						Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
							"nested_required": cty.String,
							"nested_optional": cty.String,
						}, []string{"nested_optional"}),
						DefaultValues: map[string]cty.Value{
							"nested_optional": cty.StringVal("optional"),
						},
					},
				},
			},
			value: cty.ObjectVal(map[string]cty.Value{
				"required": cty.StringVal("required"),
				// optional_object is explicitly set to null for this test case.
				"optional_object": cty.NullVal(cty.Object(map[string]cty.Type{
					"nested_required": cty.String,
					"nested_optional": cty.String,
				})),
			}),
			want: cty.ObjectVal(map[string]cty.Value{
				"required": cty.StringVal("required"),
				"optional_object": cty.ObjectVal(map[string]cty.Value{
					"nested_required": cty.StringVal("required"),
					"nested_optional": cty.StringVal("optional"),
				}),
			}),
		},
		"defaults with unset defaults are still applied (missing)": {
			defaults: &Defaults{
				Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
					"required": cty.String,
					"optional_object": cty.ObjectWithOptionalAttrs(map[string]cty.Type{
						"nested_required": cty.String,
						"nested_optional": cty.String,
					}, []string{"nested_optional"}),
				}, []string{"optional_object"}),
				DefaultValues: map[string]cty.Value{
					"optional_object": cty.ObjectVal(map[string]cty.Value{
						"nested_required": cty.StringVal("required"),
						"nested_optional": cty.NullVal(cty.String),
					}),
				},
				Children: map[string]*Defaults{
					"optional_object": {
						Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
							"nested_required": cty.String,
							"nested_optional": cty.String,
						}, []string{"nested_optional"}),
						DefaultValues: map[string]cty.Value{
							"nested_optional": cty.StringVal("optional"),
						},
					},
				},
			},
			value: cty.ObjectVal(map[string]cty.Value{
				"required": cty.StringVal("required"),
				// optional_object is missing but not null for this test case.
			}),
			want: cty.ObjectVal(map[string]cty.Value{
				"required": cty.StringVal("required"),
				"optional_object": cty.ObjectVal(map[string]cty.Value{
					"nested_required": cty.StringVal("required"),
					"nested_optional": cty.StringVal("optional"),
				}),
			}),
		},
		// https://discuss.hashicorp.com/t/request-for-feedback-optional-object-type-attributes-with-defaults-in-v1-3-alpha/40550/6?u=alisdair
		"all child and nested values are optional with defaults": {
			defaults: &Defaults{
				Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
					"settings": cty.ObjectWithOptionalAttrs(map[string]cty.Type{
						"setting_one": cty.String,
						"setting_two": cty.Number,
					}, []string{"setting_one", "setting_two"}),
				}, []string{"settings"}),
				DefaultValues: map[string]cty.Value{
					"settings": cty.EmptyObjectVal,
				},
				Children: map[string]*Defaults{
					"settings": {
						Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
							"setting_one": cty.String,
							"setting_two": cty.String,
						}, []string{"setting_one", "setting_two"}),
						DefaultValues: map[string]cty.Value{
							"setting_one": cty.StringVal(""),
							"setting_two": cty.NumberIntVal(0),
						},
					},
				},
			},
			value: cty.EmptyObjectVal,
			want: cty.ObjectVal(map[string]cty.Value{
				"settings": cty.ObjectVal(map[string]cty.Value{
					"setting_one": cty.StringVal(""),
					"setting_two": cty.NumberIntVal(0),
				}),
			}),
		},
		"all nested values are optional with defaults, but direct child has no default": {
			defaults: &Defaults{
				Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
					"settings": cty.ObjectWithOptionalAttrs(map[string]cty.Type{
						"setting_one": cty.String,
						"setting_two": cty.Number,
					}, []string{"setting_one", "setting_two"}),
				}, []string{"settings"}),
				Children: map[string]*Defaults{
					"settings": {
						Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
							"setting_one": cty.String,
							"setting_two": cty.String,
						}, []string{"setting_one", "setting_two"}),
						DefaultValues: map[string]cty.Value{
							"setting_one": cty.StringVal(""),
							"setting_two": cty.NumberIntVal(0),
						},
					},
				},
			},
			value: cty.EmptyObjectVal,
			want:  cty.EmptyObjectVal,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := tc.defaults.Apply(tc.value)
			if !cmp.Equal(tc.want, got, valueComparer) {
				t.Errorf("wrong result\n%s", cmp.Diff(tc.want, got, valueComparer))
			}
		})
	}
}
