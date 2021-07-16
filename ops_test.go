package hcl

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestApplyPath(t *testing.T) {
	tests := []struct {
		Start   cty.Value
		Path    cty.Path
		Want    cty.Value
		WantErr string
	}{
		{
			cty.StringVal("hello"),
			nil,
			cty.StringVal("hello"),
			``,
		},
		{
			cty.StringVal("hello"),
			(cty.Path)(nil).Index(cty.StringVal("boop")),
			cty.NilVal,
			`Invalid index: This value does not have any indices.`,
		},
		{
			cty.StringVal("hello"),
			(cty.Path)(nil).Index(cty.NumberIntVal(0)),
			cty.NilVal,
			`Invalid index: This value does not have any indices.`,
		},
		{
			cty.ListVal([]cty.Value{
				cty.StringVal("hello"),
			}),
			(cty.Path)(nil).Index(cty.NumberIntVal(0)),
			cty.StringVal("hello"),
			``,
		},
		{
			cty.ListVal([]cty.Value{
				cty.StringVal("hello"),
			}).Mark("x"),
			(cty.Path)(nil).Index(cty.NumberIntVal(0)),
			cty.StringVal("hello").Mark("x"),
			``,
		},
		{
			cty.TupleVal([]cty.Value{
				cty.StringVal("hello"),
			}),
			(cty.Path)(nil).Index(cty.NumberIntVal(0)),
			cty.StringVal("hello"),
			``,
		},
		{
			cty.MapVal(map[string]cty.Value{
				"a": cty.StringVal("foo").Mark("x"),
				"b": cty.StringVal("bar").Mark("x"),
			}).Mark("x"),
			cty.GetAttrPath("a"),
			cty.StringVal("foo").Mark("x"),
			``,
		},

		{
			cty.ListValEmpty(cty.String),
			(cty.Path)(nil).Index(cty.NumberIntVal(0)),
			cty.NilVal,
			`Invalid index: The given key does not identify an element in this collection value: the collection has no elements.`,
		},
		{
			cty.ListVal([]cty.Value{
				cty.StringVal("hello"),
			}),
			(cty.Path)(nil).Index(cty.NumberIntVal(1)),
			cty.NilVal,
			`Invalid index: The given key does not identify an element in this collection value: the given index is greater than or equal to the length of the collection.`,
		},
		{
			cty.ListVal([]cty.Value{
				cty.StringVal("hello"),
			}).Mark("boop"), // prevents us from making statements about the length of the list
			(cty.Path)(nil).Index(cty.NumberIntVal(1)),
			cty.NilVal,
			`Invalid index: The given key does not identify an element in this collection value.`,
		},
		{
			cty.ListVal([]cty.Value{
				cty.StringVal("hello"),
			}),
			(cty.Path)(nil).Index(cty.NumberIntVal(-1)),
			cty.NilVal,
			`Invalid index: The given key does not identify an element in this collection value: a negative number is not a valid index for a sequence.`,
		},
		{
			cty.ListVal([]cty.Value{
				cty.StringVal("hello"),
			}),
			(cty.Path)(nil).Index(cty.NumberFloatVal(0.5)),
			cty.NilVal,
			`Invalid index: The given key does not identify an element in this collection value: indexing a sequence requires a whole number, but the given index has a fractional part.`,
		},
		{
			cty.ListVal([]cty.Value{
				cty.StringVal("hello"),
			}),
			(cty.Path)(nil).Index(cty.NumberIntVal(0)).GetAttr("foo"),
			cty.NilVal,
			`Unsupported attribute: Can't access attributes on a primitive-typed value (string).`,
		},
		{
			cty.ListVal([]cty.Value{
				cty.EmptyObjectVal,
			}),
			(cty.Path)(nil).Index(cty.NumberIntVal(0)).GetAttr("foo"),
			cty.NilVal,
			`Unsupported attribute: This object does not have an attribute named "foo".`,
		},
		{
			cty.ListVal([]cty.Value{
				cty.EmptyObjectVal,
			}),
			(cty.Path)(nil).GetAttr("foo"),
			cty.NilVal,
			`Unsupported attribute: Can't access attributes on a list of objects. Did you mean to access an attribute for a specific element of the list, or across all elements of the list?`,
		},
		{
			cty.ListVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"foo": cty.True,
				}),
			}),
			(cty.Path)(nil).GetAttr("foo"),
			cty.NilVal,
			`Unsupported attribute: Can't access attributes on a list of objects. Did you mean to access attribute "foo" for a specific element of the list, or across all elements of the list?`,
		},

		{
			cty.EmptyTupleVal,
			(cty.Path)(nil).Index(cty.NumberIntVal(0)),
			cty.NilVal,
			`Invalid index: The given key does not identify an element in this collection value: the collection has no elements.`,
		},
		{
			cty.TupleVal([]cty.Value{
				cty.StringVal("hello"),
			}),
			(cty.Path)(nil).Index(cty.NumberIntVal(1)),
			cty.NilVal,
			`Invalid index: The given key does not identify an element in this collection value: the given index is greater than or equal to the length of the collection.`,
		},
		{
			cty.TupleVal([]cty.Value{
				cty.StringVal("hello"),
			}).Mark("boop"),
			(cty.Path)(nil).Index(cty.NumberIntVal(1)),
			cty.NilVal,
			`Invalid index: The given key does not identify an element in this collection value.`,
		},
		{
			cty.TupleVal([]cty.Value{
				cty.StringVal("hello"),
			}),
			(cty.Path)(nil).Index(cty.NumberIntVal(-1)),
			cty.NilVal,
			`Invalid index: The given key does not identify an element in this collection value: a negative number is not a valid index for a sequence.`,
		},
		{
			cty.TupleVal([]cty.Value{
				cty.StringVal("hello"),
			}),
			(cty.Path)(nil).Index(cty.NumberFloatVal(0.5)),
			cty.NilVal,
			`Invalid index: The given key does not identify an element in this collection value: indexing a sequence requires a whole number, but the given index has a fractional part.`,
		},
		{
			cty.TupleVal([]cty.Value{
				cty.StringVal("hello"),
			}),
			(cty.Path)(nil).Index(cty.NumberIntVal(0)).GetAttr("foo"),
			cty.NilVal,
			`Unsupported attribute: Can't access attributes on a primitive-typed value (string).`,
		},
		{
			cty.TupleVal([]cty.Value{
				cty.EmptyObjectVal,
			}),
			(cty.Path)(nil).Index(cty.NumberIntVal(0)).GetAttr("foo"),
			cty.NilVal,
			`Unsupported attribute: This object does not have an attribute named "foo".`,
		},
		{
			cty.TupleVal([]cty.Value{
				cty.EmptyObjectVal,
			}),
			(cty.Path)(nil).GetAttr("foo"),
			cty.NilVal,
			`Unsupported attribute: This value does not have any attributes.`,
		},
		{
			cty.TupleVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"foo": cty.True,
				}),
			}),
			(cty.Path)(nil).GetAttr("foo"),
			cty.NilVal,
			`Unsupported attribute: This value does not have any attributes.`,
		},

		{
			cty.SetVal([]cty.Value{
				cty.StringVal("hello"),
			}),
			(cty.Path)(nil).Index(cty.NumberIntVal(1)),
			cty.NilVal,
			`Invalid index: Elements of a set are identified only by their value and don't have any separate index or key to select with, so it's only possible to perform operations across all elements of the set.`,
		},
		{
			cty.SetVal([]cty.Value{
				cty.EmptyObjectVal,
			}),
			(cty.Path)(nil).GetAttr("foo"),
			cty.NilVal,
			`Unsupported attribute: Can't access attributes on a set of objects. Did you mean to access an attribute across all elements of the set?`,
		},
		{
			cty.NullVal(cty.List(cty.String)),
			(cty.Path)(nil).Index(cty.NumberIntVal(0)),
			cty.NilVal,
			`Attempt to index null value: This value is null, so it does not have any indices.`,
		},
		{
			cty.NullVal(cty.Map(cty.String)),
			(cty.Path)(nil).Index(cty.NumberIntVal(0)),
			cty.NilVal,
			`Attempt to index null value: This value is null, so it does not have any indices.`,
		},
		{
			cty.NullVal(cty.EmptyObject),
			(cty.Path)(nil).GetAttr("foo"),
			cty.NilVal,
			`Attempt to get attribute from null value: This value is null, so it does not have any attributes.`,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v %#v", test.Start, test.Path), func(t *testing.T) {
			got, diags := ApplyPath(test.Start, test.Path, nil)
			t.Logf("testing ApplyPath\nstart: %#v\npath:  %#v", test.Start, test.Path)

			for _, diag := range diags {
				t.Logf(diag.Error())
			}

			if test.WantErr != "" {
				if !diags.HasErrors() {
					t.Fatalf("succeeded, but want error\nwant error: %s", test.WantErr)
				}
				if len(diags) != 1 {
					t.Fatalf("wrong number of diagnostics %d; want 1", len(diags))
				}

				if gotErrStr := diags[0].Summary + ": " + diags[0].Detail; gotErrStr != test.WantErr {
					t.Fatalf("wrong error\ngot error:  %s\nwant error: %s", gotErrStr, test.WantErr)
				}
				return
			}

			if diags.HasErrors() {
				t.Fatalf("failed, but want success\ngot diagnostics:\n%s", diags.Error())
			}
			if !test.Want.RawEquals(got) {
				t.Fatalf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestIndex(t *testing.T) {
	tests := map[string]struct {
		coll cty.Value
		key  cty.Value
		want cty.Value
		err  string
	}{
		"marked key to maked value": {
			coll: cty.ListVal([]cty.Value{
				cty.StringVal("a"),
			}),
			key:  cty.NumberIntVal(0).Mark("marked"),
			want: cty.StringVal("a").Mark("marked"),
		},
		"missing list key": {
			coll: cty.ListVal([]cty.Value{
				cty.StringVal("a"),
			}),
			key:  cty.NumberIntVal(1).Mark("marked"),
			want: cty.DynamicVal,
			err:  "Invalid index",
		},
		"null marked key": {
			coll: cty.ListVal([]cty.Value{
				cty.StringVal("a"),
			}),
			key:  cty.NullVal(cty.Number).Mark("marked"),
			want: cty.DynamicVal,
			err:  "Invalid index",
		},
		"dynamic key": {
			coll: cty.ListVal([]cty.Value{
				cty.StringVal("a"),
			}),
			key:  cty.DynamicVal,
			want: cty.DynamicVal,
		},
		"invalid marked key type": {
			coll: cty.ListVal([]cty.Value{
				cty.StringVal("a"),
			}),
			key:  cty.StringVal("foo").Mark("marked"),
			want: cty.DynamicVal,
			err:  "Invalid index",
		},
		"marked map key": {
			coll: cty.MapVal(map[string]cty.Value{
				"foo": cty.StringVal("a"),
			}),
			key:  cty.StringVal("foo").Mark("marked"),
			want: cty.StringVal("a").Mark("marked"),
		},
		"missing marked map key": {
			coll: cty.MapVal(map[string]cty.Value{
				"foo": cty.StringVal("a"),
			}),
			key:  cty.StringVal("bar").Mark("mark"),
			want: cty.DynamicVal,
			err:  "Invalid index",
		},
		"marked object key": {
			coll: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.StringVal("a"),
			}),
			key: cty.StringVal("foo").Mark("marked"),
			// an object attribute is fetched by string index, and the marks
			// are not maintained
			want: cty.StringVal("a"),
		},
		"invalid marked object key type": {
			coll: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.StringVal("a"),
			}),
			key:  cty.ListVal([]cty.Value{cty.NullVal(cty.String)}).Mark("marked"),
			want: cty.DynamicVal,
			err:  "Invalid index",
		},
		"invalid marked object key": {
			coll: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.StringVal("a"),
			}),
			key:  cty.NumberIntVal(0).Mark("marked"),
			want: cty.DynamicVal,
			err:  "Invalid index",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Logf("testing Index\ncollection: %#v\nkey:  %#v", tc.coll, tc.key)

			got, diags := Index(tc.coll, tc.key, nil)

			for _, diag := range diags {
				t.Logf(diag.Error())
			}

			if tc.err != "" {
				if !diags.HasErrors() {
					t.Fatalf("succeeded, but want error\nwant error: %s", tc.err)
				}
				if len(diags) != 1 {
					t.Fatalf("wrong number of diagnostics %d; want 1", len(diags))
				}

				if gotErrStr := diags[0].Summary; gotErrStr != tc.err {
					t.Fatalf("wrong error\ngot error:  %s\nwant error: %s", gotErrStr, tc.err)
				}
				return
			}

			if diags.HasErrors() {
				t.Fatalf("failed, but want success\ngot diagnostics:\n%s", diags.Error())
			}
			if !tc.want.RawEquals(got) {
				t.Fatalf("wrong result\ngot:  %#v\nwant: %#v", got, tc.want)
			}
		})
	}
}
