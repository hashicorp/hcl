// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package dynblock

import (
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/hcl/v2/hcltest"
	"github.com/zclconf/go-cty-debug/ctydebug"
	"github.com/zclconf/go-cty/cty"
)

func TestExpand(t *testing.T) {
	srcBody := hcltest.MockBody(&hcl.BodyContent{
		Blocks: hcl.Blocks{
			{
				Type:        "a",
				Labels:      []string{"static0"},
				LabelRanges: []hcl.Range{hcl.Range{}},
				Body: hcltest.MockBody(&hcl.BodyContent{
					Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
						"val": hcltest.MockExprLiteral(cty.StringVal("static a 0")),
					}),
				}),
			},
			{
				Type: "b",
				Body: hcltest.MockBody(&hcl.BodyContent{
					Blocks: hcl.Blocks{
						{
							Type: "c",
							Body: hcltest.MockBody(&hcl.BodyContent{
								Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
									"val0": hcltest.MockExprLiteral(cty.StringVal("static c 0")),
								}),
							}),
						},
						{
							Type:        "dynamic",
							Labels:      []string{"c"},
							LabelRanges: []hcl.Range{hcl.Range{}},
							Body: hcltest.MockBody(&hcl.BodyContent{
								Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
									"for_each": hcltest.MockExprLiteral(cty.ListVal([]cty.Value{
										cty.StringVal("dynamic c 0"),
										cty.StringVal("dynamic c 1"),
									})),
									"iterator": hcltest.MockExprVariable("dyn_c"),
								}),
								Blocks: hcl.Blocks{
									{
										Type: "content",
										Body: hcltest.MockBody(&hcl.BodyContent{
											Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
												"val0": hcltest.MockExprTraversalSrc("dyn_c.value"),
											}),
										}),
									},
								},
							}),
						},
					},
				}),
			},
			{
				Type:        "dynamic",
				Labels:      []string{"a"},
				LabelRanges: []hcl.Range{hcl.Range{}},
				Body: hcltest.MockBody(&hcl.BodyContent{
					Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
						"for_each": hcltest.MockExprLiteral(cty.ListVal([]cty.Value{
							cty.StringVal("dynamic a 0"),
							cty.StringVal("dynamic a 1"),
							cty.StringVal("dynamic a 2"),
						})),
						"labels": hcltest.MockExprList([]hcl.Expression{
							hcltest.MockExprTraversalSrc("a.key"),
						}),
					}),
					Blocks: hcl.Blocks{
						{
							Type: "content",
							Body: hcltest.MockBody(&hcl.BodyContent{
								Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
									"val": hcltest.MockExprTraversalSrc("a.value"),
								}),
							}),
						},
					},
				}),
			},
			{
				Type:        "dynamic",
				Labels:      []string{"b"},
				LabelRanges: []hcl.Range{hcl.Range{}},
				Body: hcltest.MockBody(&hcl.BodyContent{
					Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
						"for_each": hcltest.MockExprLiteral(cty.ListVal([]cty.Value{
							cty.StringVal("dynamic b 0"),
							cty.StringVal("dynamic b 1"),
						})),
						"iterator": hcltest.MockExprVariable("dyn_b"),
					}),
					Blocks: hcl.Blocks{
						{
							Type: "content",
							Body: hcltest.MockBody(&hcl.BodyContent{
								Blocks: hcl.Blocks{
									{
										Type: "c",
										Body: hcltest.MockBody(&hcl.BodyContent{
											Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
												"val0": hcltest.MockExprLiteral(cty.StringVal("static c 1")),
												"val1": hcltest.MockExprTraversalSrc("dyn_b.value"),
											}),
										}),
									},
									{
										Type:        "dynamic",
										Labels:      []string{"c"},
										LabelRanges: []hcl.Range{hcl.Range{}},
										Body: hcltest.MockBody(&hcl.BodyContent{
											Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
												"for_each": hcltest.MockExprLiteral(cty.ListVal([]cty.Value{
													cty.StringVal("dynamic c 2"),
													cty.StringVal("dynamic c 3"),
												})),
											}),
											Blocks: hcl.Blocks{
												{
													Type: "content",
													Body: hcltest.MockBody(&hcl.BodyContent{
														Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
															"val0": hcltest.MockExprTraversalSrc("c.value"),
															"val1": hcltest.MockExprTraversalSrc("dyn_b.value"),
														}),
													}),
												},
											},
										}),
									},
								},
							}),
						},
					},
				}),
			},
			{
				Type:        "dynamic",
				Labels:      []string{"b"},
				LabelRanges: []hcl.Range{hcl.Range{}},
				Body: hcltest.MockBody(&hcl.BodyContent{
					Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
						"for_each": hcltest.MockExprLiteral(cty.MapVal(map[string]cty.Value{
							"foo": cty.ListVal([]cty.Value{
								cty.StringVal("dynamic c nested 0"),
								cty.StringVal("dynamic c nested 1"),
							}),
						})),
						"iterator": hcltest.MockExprVariable("dyn_b"),
					}),
					Blocks: hcl.Blocks{
						{
							Type: "content",
							Body: hcltest.MockBody(&hcl.BodyContent{
								Blocks: hcl.Blocks{
									{
										Type:        "dynamic",
										Labels:      []string{"c"},
										LabelRanges: []hcl.Range{hcl.Range{}},
										Body: hcltest.MockBody(&hcl.BodyContent{
											Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
												"for_each": hcltest.MockExprTraversalSrc("dyn_b.value"),
											}),
											Blocks: hcl.Blocks{
												{
													Type: "content",
													Body: hcltest.MockBody(&hcl.BodyContent{
														Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
															"val0": hcltest.MockExprTraversalSrc("c.value"),
															"val1": hcltest.MockExprTraversalSrc("dyn_b.key"),
														}),
													}),
												},
											},
										}),
									},
								},
							}),
						},
					},
				}),
			},
			{
				Type:        "a",
				Labels:      []string{"static1"},
				LabelRanges: []hcl.Range{hcl.Range{}},
				Body: hcltest.MockBody(&hcl.BodyContent{
					Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
						"val": hcltest.MockExprLiteral(cty.StringVal("static a 1")),
					}),
				}),
			},
		},
	})

	dynBody := Expand(srcBody, nil)
	var remain hcl.Body

	t.Run("PartialDecode", func(t *testing.T) {
		decSpec := &hcldec.BlockMapSpec{
			TypeName:   "a",
			LabelNames: []string{"key"},
			Nested: &hcldec.AttrSpec{
				Name:     "val",
				Type:     cty.String,
				Required: true,
			},
		}

		var got cty.Value
		var diags hcl.Diagnostics
		got, remain, diags = hcldec.PartialDecode(dynBody, decSpec, nil)
		if len(diags) != 0 {
			t.Errorf("unexpected diagnostics")
			for _, diag := range diags {
				t.Logf("- %s", diag)
			}
			return
		}

		want := cty.MapVal(map[string]cty.Value{
			"static0": cty.StringVal("static a 0"),
			"static1": cty.StringVal("static a 1"),
			"0":       cty.StringVal("dynamic a 0"),
			"1":       cty.StringVal("dynamic a 1"),
			"2":       cty.StringVal("dynamic a 2"),
		})

		if !got.RawEquals(want) {
			t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, want)
		}
	})

	t.Run("Decode", func(t *testing.T) {
		decSpec := &hcldec.BlockListSpec{
			TypeName: "b",
			Nested: &hcldec.BlockListSpec{
				TypeName: "c",
				Nested: &hcldec.ObjectSpec{
					"val0": &hcldec.AttrSpec{
						Name: "val0",
						Type: cty.String,
					},
					"val1": &hcldec.AttrSpec{
						Name: "val1",
						Type: cty.String,
					},
				},
			},
		}

		var got cty.Value
		var diags hcl.Diagnostics
		got, diags = hcldec.Decode(remain, decSpec, nil)
		if len(diags) != 0 {
			t.Errorf("unexpected diagnostics")
			for _, diag := range diags {
				t.Logf("- %s", diag)
			}
			return
		}

		want := cty.ListVal([]cty.Value{
			cty.ListVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"val0": cty.StringVal("static c 0"),
					"val1": cty.NullVal(cty.String),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"val0": cty.StringVal("dynamic c 0"),
					"val1": cty.NullVal(cty.String),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"val0": cty.StringVal("dynamic c 1"),
					"val1": cty.NullVal(cty.String),
				}),
			}),
			cty.ListVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"val0": cty.StringVal("static c 1"),
					"val1": cty.StringVal("dynamic b 0"),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"val0": cty.StringVal("dynamic c 2"),
					"val1": cty.StringVal("dynamic b 0"),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"val0": cty.StringVal("dynamic c 3"),
					"val1": cty.StringVal("dynamic b 0"),
				}),
			}),
			cty.ListVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"val0": cty.StringVal("static c 1"),
					"val1": cty.StringVal("dynamic b 1"),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"val0": cty.StringVal("dynamic c 2"),
					"val1": cty.StringVal("dynamic b 1"),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"val0": cty.StringVal("dynamic c 3"),
					"val1": cty.StringVal("dynamic b 1"),
				}),
			}),
			cty.ListVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"val0": cty.StringVal("dynamic c nested 0"),
					"val1": cty.StringVal("foo"),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"val0": cty.StringVal("dynamic c nested 1"),
					"val1": cty.StringVal("foo"),
				}),
			}),
		})

		if !got.RawEquals(want) {
			t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, want)
		}
	})

}

func TestExpandWithForEachCheck(t *testing.T) {
	forEachExpr := hcltest.MockExprLiteral(cty.MapValEmpty(cty.String).Mark("boop"))
	evalCtx := &hcl.EvalContext{}
	srcContent := &hcl.BodyContent{
		Blocks: hcl.Blocks{
			{
				Type:        "dynamic",
				Labels:      []string{"foo"},
				LabelRanges: []hcl.Range{{}},
				Body: hcltest.MockBody(&hcl.BodyContent{
					Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
						"for_each": forEachExpr,
					}),
					Blocks: hcl.Blocks{
						{
							Type: "content",
							Body: hcltest.MockBody(&hcl.BodyContent{}),
						},
					},
				}),
			},
		},
	}
	srcBody := hcltest.MockBody(srcContent)

	hookCalled := false
	var gotV cty.Value
	var gotEvalCtx *hcl.EvalContext

	expBody := Expand(
		srcBody, evalCtx,
		OptCheckForEach(func(v cty.Value, e hcl.Expression, ec *hcl.EvalContext) hcl.Diagnostics {
			hookCalled = true
			gotV = v
			gotEvalCtx = ec
			return hcl.Diagnostics{
				&hcl.Diagnostic{
					Severity:    hcl.DiagError,
					Summary:     "Bad for_each",
					Detail:      "I don't like it.",
					Expression:  e,
					EvalContext: ec,
					Extra:       "diagnostic extra",
				},
			}
		}),
	)

	_, diags := expBody.Content(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type: "foo",
			},
		},
	})
	if !diags.HasErrors() {
		t.Fatal("succeeded; want an error")
	}
	if len(diags) != 1 {
		t.Fatalf("wrong number of diagnostics; want only one\n%s", spew.Sdump(diags))
	}
	if got, want := diags[0].Summary, "Bad for_each"; got != want {
		t.Fatalf("wrong error\ngot:  %s\nwant: %s\n\n%s", got, want, spew.Sdump(diags[0]))
	}
	if got, want := diags[0].Extra, "diagnostic extra"; got != want {
		// This is important to allow the application which provided the
		// hook to pass application-specific extra values through this
		// API in case the hook's diagnostics need some sort of special
		// treatment.
		t.Fatalf("diagnostic didn't preserve 'extra' field\ngot:  %s\nwant: %s\n\n%s", got, want, spew.Sdump(diags[0]))
	}

	if !hookCalled {
		t.Fatal("check hook wasn't called")
	}
	if !gotV.HasMark("boop") {
		t.Errorf("wrong value passed to check hook; want the value marked \"boop\"\n%s", ctydebug.ValueString(gotV))
	}
	if gotEvalCtx != evalCtx {
		t.Error("wrong EvalContext passed to check hook; want the one passed to Expand")
	}
}

func TestExpandUnknownBodies(t *testing.T) {
	srcContent := &hcl.BodyContent{
		Blocks: hcl.Blocks{
			{
				Type:        "dynamic",
				Labels:      []string{"list"},
				LabelRanges: []hcl.Range{hcl.Range{}},
				Body: hcltest.MockBody(&hcl.BodyContent{
					Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
						"for_each": hcltest.MockExprLiteral(cty.UnknownVal(cty.Map(cty.String))),
					}),
					Blocks: hcl.Blocks{
						{
							Type: "content",
							Body: hcltest.MockBody(&hcl.BodyContent{
								Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
									"val": hcltest.MockExprTraversalSrc("each.value"),
								}),
							}),
						},
					},
				}),
			},
			{
				Type:        "dynamic",
				Labels:      []string{"tuple"},
				LabelRanges: []hcl.Range{hcl.Range{}},
				Body: hcltest.MockBody(&hcl.BodyContent{
					Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
						"for_each": hcltest.MockExprLiteral(cty.UnknownVal(cty.Map(cty.String))),
					}),
					Blocks: hcl.Blocks{
						{
							Type: "content",
							Body: hcltest.MockBody(&hcl.BodyContent{
								Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
									"val": hcltest.MockExprTraversalSrc("each.value"),
								}),
							}),
						},
					},
				}),
			},
			{
				Type:        "dynamic",
				Labels:      []string{"set"},
				LabelRanges: []hcl.Range{hcl.Range{}},
				Body: hcltest.MockBody(&hcl.BodyContent{
					Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
						"for_each": hcltest.MockExprLiteral(cty.UnknownVal(cty.Map(cty.String))),
					}),
					Blocks: hcl.Blocks{
						{
							Type: "content",
							Body: hcltest.MockBody(&hcl.BodyContent{
								Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
									"val": hcltest.MockExprTraversalSrc("each.value"),
								}),
							}),
						},
					},
				}),
			},
			{
				Type:        "dynamic",
				Labels:      []string{"map"},
				LabelRanges: []hcl.Range{hcl.Range{}},
				Body: hcltest.MockBody(&hcl.BodyContent{
					Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
						"for_each": hcltest.MockExprLiteral(cty.UnknownVal(cty.Map(cty.String))),
						"labels": hcltest.MockExprList([]hcl.Expression{
							hcltest.MockExprLiteral(cty.StringVal("static")),
						}),
					}),
					Blocks: hcl.Blocks{
						{
							Type: "content",
							Body: hcltest.MockBody(&hcl.BodyContent{
								Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
									"val": hcltest.MockExprTraversalSrc("each.value"),
								}),
							}),
						},
					},
				}),
			},
			{
				Type:        "dynamic",
				Labels:      []string{"object"},
				LabelRanges: []hcl.Range{hcl.Range{}},
				Body: hcltest.MockBody(&hcl.BodyContent{
					Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
						"for_each": hcltest.MockExprLiteral(cty.UnknownVal(cty.Map(cty.String))),
						"labels": hcltest.MockExprList([]hcl.Expression{
							hcltest.MockExprLiteral(cty.StringVal("static")),
						}),
					}),
					Blocks: hcl.Blocks{
						{
							Type: "content",
							Body: hcltest.MockBody(&hcl.BodyContent{
								Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
									"val": hcltest.MockExprTraversalSrc("each.value"),
								}),
							}),
						},
					},
				}),
			},
			{
				Type:        "dynamic",
				Labels:      []string{"invalid_list"},
				LabelRanges: []hcl.Range{hcl.Range{}},
				Body: hcltest.MockBody(&hcl.BodyContent{
					Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
						"for_each": hcltest.MockExprLiteral(cty.UnknownVal(cty.Map(cty.String))),
					}),
					Blocks: hcl.Blocks{
						{
							Type: "content",
							Body: hcltest.MockBody(&hcl.BodyContent{
								Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
									"val": hcltest.MockExprTraversalSrc("each.value"),
									// unexpected attributes should still produce an error
									"invalid": hcltest.MockExprLiteral(cty.StringVal("static")),
								}),
							}),
						},
					},
				}),
			},
		},
	}

	srcBody := hcltest.MockBody(srcContent)
	dynBody := Expand(srcBody, nil)

	t.Run("DecodeList", func(t *testing.T) {
		decSpec := &hcldec.BlockListSpec{
			TypeName: "list",
			Nested: &hcldec.ObjectSpec{
				"val": &hcldec.AttrSpec{
					Name: "val",
					Type: cty.String,
				},
			},
		}

		var got cty.Value
		var diags hcl.Diagnostics

		got, _, diags = hcldec.PartialDecode(dynBody, decSpec, nil)
		if len(diags) != 0 {
			t.Errorf("unexpected diagnostics")
			for _, diag := range diags {
				t.Logf("- %s", diag)
			}
			return
		}

		want := cty.UnknownVal(cty.List(cty.Object(map[string]cty.Type{
			"val": cty.String,
		})))

		if !got.RawEquals(want) {
			t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, want)
		}
	})

	t.Run("DecodeTuple", func(t *testing.T) {
		decSpec := &hcldec.BlockTupleSpec{
			TypeName: "tuple",
			Nested: &hcldec.ObjectSpec{
				"val": &hcldec.AttrSpec{
					Name: "val",
					Type: cty.String,
				},
			},
		}

		var got cty.Value
		var diags hcl.Diagnostics

		got, _, diags = hcldec.PartialDecode(dynBody, decSpec, nil)
		if len(diags) != 0 {
			t.Errorf("unexpected diagnostics")
			for _, diag := range diags {
				t.Logf("- %s", diag)
			}
			return
		}

		want := cty.DynamicVal

		if !got.RawEquals(want) {
			t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, want)
		}
	})

	t.Run("DecodeSet", func(t *testing.T) {
		decSpec := &hcldec.BlockSetSpec{
			TypeName: "tuple",
			Nested: &hcldec.ObjectSpec{
				"val": &hcldec.AttrSpec{
					Name: "val",
					Type: cty.String,
				},
			},
		}

		var got cty.Value
		var diags hcl.Diagnostics

		got, _, diags = hcldec.PartialDecode(dynBody, decSpec, nil)
		if len(diags) != 0 {
			t.Errorf("unexpected diagnostics")
			for _, diag := range diags {
				t.Logf("- %s", diag)
			}
			return
		}

		want := cty.UnknownVal(cty.Set(cty.Object(map[string]cty.Type{
			"val": cty.String,
		})))

		if !got.RawEquals(want) {
			t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, want)
		}
	})

	t.Run("DecodeMap", func(t *testing.T) {
		decSpec := &hcldec.BlockMapSpec{
			TypeName:   "map",
			LabelNames: []string{"key"},
			Nested: &hcldec.ObjectSpec{
				"val": &hcldec.AttrSpec{
					Name: "val",
					Type: cty.String,
				},
			},
		}

		var got cty.Value
		var diags hcl.Diagnostics

		got, _, diags = hcldec.PartialDecode(dynBody, decSpec, nil)
		if len(diags) != 0 {
			t.Errorf("unexpected diagnostics")
			for _, diag := range diags {
				t.Logf("- %s", diag)
			}
			return
		}

		want := cty.UnknownVal(cty.Map(cty.Object(map[string]cty.Type{
			"val": cty.String,
		})))

		if !got.RawEquals(want) {
			t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, want)
		}
	})

	t.Run("DecodeInvalidList", func(t *testing.T) {
		decSpec := &hcldec.BlockListSpec{
			TypeName: "invalid_list",
			Nested: &hcldec.ObjectSpec{
				"val": &hcldec.AttrSpec{
					Name: "val",
					Type: cty.String,
				},
			},
		}

		_, _, diags := hcldec.PartialDecode(dynBody, decSpec, nil)
		if len(diags) != 1 {
			t.Error("expected 1 extraneous argument")
		}

		want := `Mock body has extraneous argument "invalid"`

		if !strings.Contains(diags.Error(), want) {
			t.Errorf("unexpected diagnostics: %v", diags)
		}
	})

}

func TestExpandMarkedForEach(t *testing.T) {
	srcBody := hcltest.MockBody(&hcl.BodyContent{
		Blocks: hcl.Blocks{
			{
				Type:        "dynamic",
				Labels:      []string{"b"},
				LabelRanges: []hcl.Range{{}},
				Body: hcltest.MockBody(&hcl.BodyContent{
					Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
						"for_each": hcltest.MockExprLiteral(cty.TupleVal([]cty.Value{
							cty.StringVal("hey"),
						}).Mark("boop")),
						"iterator": hcltest.MockExprTraversalSrc("dyn_b"),
					}),
					Blocks: hcl.Blocks{
						{
							Type: "content",
							Body: hcltest.MockBody(&hcl.BodyContent{
								Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
									"val0": hcltest.MockExprLiteral(cty.StringVal("static c 1")),
									"val1": hcltest.MockExprTraversalSrc("dyn_b.value"),
								}),
							}),
						},
					},
				}),
			},
		},
	})

	dynBody := Expand(srcBody, nil)

	t.Run("Decode", func(t *testing.T) {
		decSpec := &hcldec.BlockListSpec{
			TypeName: "b",
			Nested: &hcldec.ObjectSpec{
				"val0": &hcldec.AttrSpec{
					Name: "val0",
					Type: cty.String,
				},
				"val1": &hcldec.AttrSpec{
					Name: "val1",
					Type: cty.String,
				},
			},
		}

		want := cty.ListVal([]cty.Value{
			cty.ObjectVal(map[string]cty.Value{
				"val0": cty.StringVal("static c 1").Mark("boop"),
				"val1": cty.StringVal("hey").Mark("boop"),
			}).Mark("boop"),
		})
		got, diags := hcldec.Decode(dynBody, decSpec, nil)
		if diags.HasErrors() {
			t.Fatalf("unexpected errors\n%s", diags.Error())
		}
		if diff := cmp.Diff(want, got, ctydebug.CmpOptions); diff != "" {
			t.Errorf("wrong result\n%s", diff)
		}
	})
}

func TestExpandInvalidIteratorError(t *testing.T) {
	srcBody := hcltest.MockBody(&hcl.BodyContent{
		Blocks: hcl.Blocks{
			{
				Type:        "dynamic",
				Labels:      []string{"b"},
				LabelRanges: []hcl.Range{hcl.Range{}},
				Body: hcltest.MockBody(&hcl.BodyContent{
					Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
						"for_each": hcltest.MockExprLiteral(cty.ListVal([]cty.Value{
							cty.StringVal("dynamic b 0"),
							cty.StringVal("dynamic b 1"),
						})),
						"iterator": hcltest.MockExprLiteral(cty.StringVal("dyn_b")),
					}),
					Blocks: hcl.Blocks{
						{
							Type: "content",
							Body: hcltest.MockBody(&hcl.BodyContent{
								Blocks: hcl.Blocks{
									{
										Type: "c",
										Body: hcltest.MockBody(&hcl.BodyContent{
											Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
												"val0": hcltest.MockExprLiteral(cty.StringVal("static c 1")),
												"val1": hcltest.MockExprTraversalSrc("dyn_b.value"),
											}),
										}),
									},
								},
							}),
						},
					},
				}),
			},
		},
	})

	dynBody := Expand(srcBody, nil)

	t.Run("Decode", func(t *testing.T) {
		decSpec := &hcldec.BlockListSpec{
			TypeName: "b",
			Nested: &hcldec.BlockListSpec{
				TypeName: "c",
				Nested: &hcldec.ObjectSpec{
					"val0": &hcldec.AttrSpec{
						Name: "val0",
						Type: cty.String,
					},
					"val1": &hcldec.AttrSpec{
						Name: "val1",
						Type: cty.String,
					},
				},
			},
		}

		var diags hcl.Diagnostics
		_, diags = hcldec.Decode(dynBody, decSpec, nil)

		if len(diags) < 1 {
			t.Errorf("Expected diagnostics, got none")
		}
		if len(diags) > 1 {
			t.Errorf("Expected one diagnostic message, got %d", len(diags))
			for _, diag := range diags {
				t.Logf("- %s", diag)
			}
		}

		if diags[0].Summary != "Invalid expression" {
			t.Errorf("Expected error subject to be invalid expression, instead it was %q", diags[0].Summary)
		}
	})

}
