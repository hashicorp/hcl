package json

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/zclconf/go-zcl/zcl"
	"github.com/davecgh/go-spew/spew"
)

func TestBodyPartialContent(t *testing.T) {
	tests := []struct {
		src       string
		schema    *zcl.BodySchema
		want      *zcl.BodyContent
		diagCount int
	}{
		{
			`{}`,
			&zcl.BodySchema{},
			&zcl.BodyContent{
				Attributes: map[string]*zcl.Attribute{},
			},
			0,
		},
		{
			`{"name":"Ermintrude"}`,
			&zcl.BodySchema{
				Attributes: []zcl.AttributeSchema{
					{
						Name: "name",
					},
				},
			},
			&zcl.BodyContent{
				Attributes: map[string]*zcl.Attribute{
					"name": &zcl.Attribute{
						Name: "name",
						Expr: &expression{
							src: &stringVal{
								Value: "Ermintrude",
								SrcRange: zcl.Range{
									Filename: "test.json",
									Start: zcl.Pos{
										Byte:   8,
										Line:   1,
										Column: 9,
									},
									End: zcl.Pos{
										Byte:   20,
										Line:   1,
										Column: 21,
									},
								},
							},
						},
						Range: zcl.Range{
							Filename: "test.json",
							Start: zcl.Pos{
								Byte:   1,
								Line:   1,
								Column: 2,
							},
							End: zcl.Pos{
								Byte:   20,
								Line:   1,
								Column: 21,
							},
						},
						NameRange: zcl.Range{
							Filename: "test.json",
							Start: zcl.Pos{
								Byte:   1,
								Line:   1,
								Column: 2,
							},
							End: zcl.Pos{
								Byte:   7,
								Line:   1,
								Column: 8,
							},
						},
					},
				},
			},
			0,
		},
		{
			`{"name":"Ermintrude"}`,
			&zcl.BodySchema{
				Attributes: []zcl.AttributeSchema{
					{
						Name:     "name",
						Required: true,
					},
					{
						Name:     "age",
						Required: true,
					},
				},
			},
			&zcl.BodyContent{
				Attributes: map[string]*zcl.Attribute{
					"name": &zcl.Attribute{
						Name: "name",
						Expr: &expression{
							src: &stringVal{
								Value: "Ermintrude",
								SrcRange: zcl.Range{
									Filename: "test.json",
									Start: zcl.Pos{
										Byte:   8,
										Line:   1,
										Column: 9,
									},
									End: zcl.Pos{
										Byte:   20,
										Line:   1,
										Column: 21,
									},
								},
							},
						},
						Range: zcl.Range{
							Filename: "test.json",
							Start: zcl.Pos{
								Byte:   1,
								Line:   1,
								Column: 2,
							},
							End: zcl.Pos{
								Byte:   20,
								Line:   1,
								Column: 21,
							},
						},
						NameRange: zcl.Range{
							Filename: "test.json",
							Start: zcl.Pos{
								Byte:   1,
								Line:   1,
								Column: 2,
							},
							End: zcl.Pos{
								Byte:   7,
								Line:   1,
								Column: 8,
							},
						},
					},
				},
			},
			1,
		},
		{
			`{"resource":{}}`,
			&zcl.BodySchema{
				Blocks: []zcl.BlockHeaderSchema{
					{
						Type: "resource",
					},
				},
			},
			&zcl.BodyContent{
				Attributes: map[string]*zcl.Attribute{},
				Blocks: zcl.Blocks{
					{
						Type:   "resource",
						Labels: []string{},
						Body: &body{
							obj: &objectVal{
								Attrs: map[string]*objectAttr{},
								SrcRange: zcl.Range{
									Filename: "test.json",
									Start: zcl.Pos{
										Byte:   12,
										Line:   1,
										Column: 13,
									},
									End: zcl.Pos{
										Byte:   14,
										Line:   1,
										Column: 15,
									},
								},
								OpenRange: zcl.Range{
									Filename: "test.json",
									Start: zcl.Pos{
										Byte:   12,
										Line:   1,
										Column: 13,
									},
									End: zcl.Pos{
										Byte:   13,
										Line:   1,
										Column: 14,
									},
								},
								CloseRange: zcl.Range{
									Filename: "test.json",
									Start: zcl.Pos{
										Byte:   13,
										Line:   1,
										Column: 14,
									},
									End: zcl.Pos{
										Byte:   14,
										Line:   1,
										Column: 15,
									},
								},
							},
						},

						DefRange: zcl.Range{
							Filename: "test.json",
							Start: zcl.Pos{
								Byte:   12,
								Line:   1,
								Column: 13,
							},
							End: zcl.Pos{
								Byte:   13,
								Line:   1,
								Column: 14,
							},
						},
						TypeRange: zcl.Range{
							Filename: "test.json",
							Start: zcl.Pos{
								Byte:   1,
								Line:   1,
								Column: 2,
							},
							End: zcl.Pos{
								Byte:   11,
								Line:   1,
								Column: 12,
							},
						},
						LabelRanges: []zcl.Range{},
					},
				},
			},
			0,
		},
		{
			`{"resource":[{},{}]}`,
			&zcl.BodySchema{
				Blocks: []zcl.BlockHeaderSchema{
					{
						Type: "resource",
					},
				},
			},
			&zcl.BodyContent{
				Attributes: map[string]*zcl.Attribute{},
				Blocks: zcl.Blocks{
					{
						Type:   "resource",
						Labels: []string{},
						Body: &body{
							obj: &objectVal{
								Attrs: map[string]*objectAttr{},
								SrcRange: zcl.Range{
									Filename: "test.json",
									Start: zcl.Pos{
										Byte:   13,
										Line:   1,
										Column: 14,
									},
									End: zcl.Pos{
										Byte:   15,
										Line:   1,
										Column: 16,
									},
								},
								OpenRange: zcl.Range{
									Filename: "test.json",
									Start: zcl.Pos{
										Byte:   13,
										Line:   1,
										Column: 14,
									},
									End: zcl.Pos{
										Byte:   14,
										Line:   1,
										Column: 15,
									},
								},
								CloseRange: zcl.Range{
									Filename: "test.json",
									Start: zcl.Pos{
										Byte:   14,
										Line:   1,
										Column: 15,
									},
									End: zcl.Pos{
										Byte:   15,
										Line:   1,
										Column: 16,
									},
								},
							},
						},

						DefRange: zcl.Range{
							Filename: "test.json",
							Start: zcl.Pos{
								Byte:   12,
								Line:   1,
								Column: 13,
							},
							End: zcl.Pos{
								Byte:   13,
								Line:   1,
								Column: 14,
							},
						},
						TypeRange: zcl.Range{
							Filename: "test.json",
							Start: zcl.Pos{
								Byte:   1,
								Line:   1,
								Column: 2,
							},
							End: zcl.Pos{
								Byte:   11,
								Line:   1,
								Column: 12,
							},
						},
						LabelRanges: []zcl.Range{},
					},
					{
						Type:   "resource",
						Labels: []string{},
						Body: &body{
							obj: &objectVal{
								Attrs: map[string]*objectAttr{},
								SrcRange: zcl.Range{
									Filename: "test.json",
									Start: zcl.Pos{
										Byte:   16,
										Line:   1,
										Column: 17,
									},
									End: zcl.Pos{
										Byte:   18,
										Line:   1,
										Column: 19,
									},
								},
								OpenRange: zcl.Range{
									Filename: "test.json",
									Start: zcl.Pos{
										Byte:   16,
										Line:   1,
										Column: 17,
									},
									End: zcl.Pos{
										Byte:   17,
										Line:   1,
										Column: 18,
									},
								},
								CloseRange: zcl.Range{
									Filename: "test.json",
									Start: zcl.Pos{
										Byte:   17,
										Line:   1,
										Column: 18,
									},
									End: zcl.Pos{
										Byte:   18,
										Line:   1,
										Column: 19,
									},
								},
							},
						},

						DefRange: zcl.Range{
							Filename: "test.json",
							Start: zcl.Pos{
								Byte:   12,
								Line:   1,
								Column: 13,
							},
							End: zcl.Pos{
								Byte:   13,
								Line:   1,
								Column: 14,
							},
						},
						TypeRange: zcl.Range{
							Filename: "test.json",
							Start: zcl.Pos{
								Byte:   1,
								Line:   1,
								Column: 2,
							},
							End: zcl.Pos{
								Byte:   11,
								Line:   1,
								Column: 12,
							},
						},
						LabelRanges: []zcl.Range{},
					},
				},
			},
			0,
		},
		{
			`{"resource":{"foo_instance":{"bar":{}}}}`,
			&zcl.BodySchema{
				Blocks: []zcl.BlockHeaderSchema{
					{
						Type:       "resource",
						LabelNames: []string{"type", "name"},
					},
				},
			},
			&zcl.BodyContent{
				Attributes: map[string]*zcl.Attribute{},
				Blocks: zcl.Blocks{
					{
						Type:   "resource",
						Labels: []string{"foo_instance", "bar"},
						Body: &body{
							obj: &objectVal{
								Attrs: map[string]*objectAttr{},
								SrcRange: zcl.Range{
									Filename: "test.json",
									Start: zcl.Pos{
										Byte:   35,
										Line:   1,
										Column: 36,
									},
									End: zcl.Pos{
										Byte:   37,
										Line:   1,
										Column: 38,
									},
								},
								OpenRange: zcl.Range{
									Filename: "test.json",
									Start: zcl.Pos{
										Byte:   35,
										Line:   1,
										Column: 36,
									},
									End: zcl.Pos{
										Byte:   36,
										Line:   1,
										Column: 37,
									},
								},
								CloseRange: zcl.Range{
									Filename: "test.json",
									Start: zcl.Pos{
										Byte:   36,
										Line:   1,
										Column: 37,
									},
									End: zcl.Pos{
										Byte:   37,
										Line:   1,
										Column: 38,
									},
								},
							},
						},

						DefRange: zcl.Range{
							Filename: "test.json",
							Start: zcl.Pos{
								Byte:   35,
								Line:   1,
								Column: 36,
							},
							End: zcl.Pos{
								Byte:   36,
								Line:   1,
								Column: 37,
							},
						},
						TypeRange: zcl.Range{
							Filename: "test.json",
							Start: zcl.Pos{
								Byte:   1,
								Line:   1,
								Column: 2,
							},
							End: zcl.Pos{
								Byte:   11,
								Line:   1,
								Column: 12,
							},
						},
						LabelRanges: []zcl.Range{
							{
								Filename: "test.json",
								Start: zcl.Pos{
									Byte:   13,
									Line:   1,
									Column: 14,
								},
								End: zcl.Pos{
									Byte:   27,
									Line:   1,
									Column: 28,
								},
							},
							{
								Filename: "test.json",
								Start: zcl.Pos{
									Byte:   29,
									Line:   1,
									Column: 30,
								},
								End: zcl.Pos{
									Byte:   34,
									Line:   1,
									Column: 35,
								},
							},
						},
					},
				},
			},
			0,
		},
		{
			`{"name":"Ermintrude"}`,
			&zcl.BodySchema{
				Blocks: []zcl.BlockHeaderSchema{
					{
						Type: "name",
					},
				},
			},
			&zcl.BodyContent{
				Attributes: map[string]*zcl.Attribute{},
			},
			1,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%02d-%s", i, test.src), func(t *testing.T) {
			file, diags := Parse([]byte(test.src), "test.json")
			if len(diags) != 0 {
				t.Fatalf("Parse produced diagnostics: %s", diags)
			}
			got, _, diags := file.Body.PartialContent(test.schema)
			if len(diags) != test.diagCount {
				t.Errorf("Wrong number of diagnostics %d; want %d", len(diags), test.diagCount)
				for _, diag := range diags {
					t.Logf(" - %s", diag)
				}
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("wrong result\ngot:  %s\nwant: %s", spew.Sdump(got), spew.Sdump(test.want))
			}
		})
	}
}

func TestBodyContent(t *testing.T) {
	// We test most of the functionality already in TestBodyPartialContent, so
	// this test focuses on the handling of extraneous attributes.
	tests := []struct {
		src       string
		schema    *zcl.BodySchema
		diagCount int
	}{
		{
			`{"unknown": true}`,
			&zcl.BodySchema{},
			1,
		},
		{
			`{"unknow": true}`,
			&zcl.BodySchema{
				Attributes: []zcl.AttributeSchema{
					{
						Name: "unknown",
					},
				},
			},
			1,
		},
		{
			`{"unknow": true, "unnown": true}`,
			&zcl.BodySchema{
				Attributes: []zcl.AttributeSchema{
					{
						Name: "unknown",
					},
				},
			},
			2,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%02d-%s", i, test.src), func(t *testing.T) {
			file, diags := Parse([]byte(test.src), "test.json")
			if len(diags) != 0 {
				t.Fatalf("Parse produced diagnostics: %s", diags)
			}
			_, diags = file.Body.Content(test.schema)
			if len(diags) != test.diagCount {
				t.Errorf("Wrong number of diagnostics %d; want %d", len(diags), test.diagCount)
				for _, diag := range diags {
					t.Logf(" - %s", diag)
				}
			}
		})
	}
}

func TestJustAttributes(t *testing.T) {
	// We test most of the functionality already in TestBodyPartialContent, so
	// this test focuses on the handling of extraneous attributes.
	tests := []struct {
		src  string
		want zcl.Attributes
	}{
		{
			`{}`,
			map[string]*zcl.Attribute{},
		},
		{
			`{"foo": true}`,
			map[string]*zcl.Attribute{
				"foo": {
					Name: "foo",
					Expr: &expression{
						src: &booleanVal{
							Value: true,
							SrcRange: zcl.Range{
								Filename: "test.json",
								Start:    zcl.Pos{Byte: 8, Line: 1, Column: 9},
								End:      zcl.Pos{Byte: 12, Line: 1, Column: 13},
							},
						},
					},
					Range: zcl.Range{
						Filename: "test.json",
						Start:    zcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:      zcl.Pos{Byte: 12, Line: 1, Column: 13},
					},
					NameRange: zcl.Range{
						Filename: "test.json",
						Start:    zcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:      zcl.Pos{Byte: 6, Line: 1, Column: 7},
					},
				},
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%02d-%s", i, test.src), func(t *testing.T) {
			file, diags := Parse([]byte(test.src), "test.json")
			if len(diags) != 0 {
				t.Fatalf("Parse produced diagnostics: %s", diags)
			}
			got, diags := file.Body.JustAttributes()
			if len(diags) != 0 {
				t.Errorf("Wrong number of diagnostics %d; want %d", len(diags), 0)
				for _, diag := range diags {
					t.Logf(" - %s", diag)
				}
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("wrong result\ngot:  %s\nwant: %s", spew.Sdump(got), spew.Sdump(test.want))
			}
		})
	}
}
