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
			`Invalid index`,
		},
		{
			cty.StringVal("hello"),
			(cty.Path)(nil).Index(cty.NumberIntVal(0)),
			cty.NilVal,
			`Invalid index`,
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
			cty.TupleVal([]cty.Value{
				cty.StringVal("hello"),
			}),
			(cty.Path)(nil).Index(cty.NumberIntVal(0)),
			cty.StringVal("hello"),
			``,
		},
		{
			cty.ListValEmpty(cty.String),
			(cty.Path)(nil).Index(cty.NumberIntVal(0)),
			cty.NilVal,
			`Invalid index`,
		},
		{
			cty.ListVal([]cty.Value{
				cty.StringVal("hello"),
			}),
			(cty.Path)(nil).Index(cty.NumberIntVal(1)),
			cty.NilVal,
			`Invalid index`,
		},
		{
			cty.ListVal([]cty.Value{
				cty.StringVal("hello"),
			}),
			(cty.Path)(nil).Index(cty.NumberFloatVal(0.5)),
			cty.NilVal,
			`Invalid index`,
		},
		{
			cty.ListVal([]cty.Value{
				cty.StringVal("hello"),
			}),
			(cty.Path)(nil).Index(cty.NumberIntVal(0)).GetAttr("foo"),
			cty.NilVal,
			`Unsupported attribute`,
		},
		{
			cty.ListVal([]cty.Value{
				cty.EmptyObjectVal,
			}),
			(cty.Path)(nil).Index(cty.NumberIntVal(0)).GetAttr("foo"),
			cty.NilVal,
			`Unsupported attribute`,
		},
		{
			cty.NullVal(cty.List(cty.String)),
			(cty.Path)(nil).Index(cty.NumberIntVal(0)),
			cty.NilVal,
			`Attempt to index null value`,
		},
		{
			cty.NullVal(cty.Map(cty.String)),
			(cty.Path)(nil).Index(cty.NumberIntVal(0)),
			cty.NilVal,
			`Attempt to index null value`,
		},
		{
			cty.NullVal(cty.EmptyObject),
			(cty.Path)(nil).GetAttr("foo"),
			cty.NilVal,
			`Attempt to get attribute from null value`,
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

				if gotErrStr := diags[0].Summary; gotErrStr != test.WantErr {
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
