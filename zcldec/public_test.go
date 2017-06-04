package zcldec

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-zcl/zcl"
	"github.com/zclconf/go-zcl/zcl/zclsyntax"
)

func TestDecode(t *testing.T) {
	tests := []struct {
		config    string
		spec      Spec
		ctx       *zcl.EvalContext
		want      cty.Value
		diagCount int
	}{
		{
			``,
			&ObjectSpec{},
			nil,
			cty.EmptyObjectVal,
			0,
		},
		{
			`a = 1`,
			&ObjectSpec{},
			nil,
			cty.EmptyObjectVal,
			1, // attribute named "a" is not expected here
		},
		{
			`a = 1`,
			&ObjectSpec{
				"a": &AttrSpec{
					Name: "a",
					Type: cty.Number,
				},
			},
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"a": cty.NumberIntVal(1),
			}),
			0,
		},
		{
			`a = 1`,
			&AttrSpec{
				Name: "a",
				Type: cty.Number,
			},
			nil,
			cty.NumberIntVal(1),
			0,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%02d-%s", i, test.config), func(t *testing.T) {
			file, parseDiags := zclsyntax.ParseConfig([]byte(test.config), "", zcl.Pos{Line: 1, Column: 1, Byte: 0})
			body := file.Body
			got, valDiags := Decode(body, test.spec, test.ctx)

			var diags zcl.Diagnostics
			diags = append(diags, parseDiags...)
			diags = append(diags, valDiags...)

			if len(diags) != test.diagCount {
				t.Errorf("wrong number of diagnostics %d; want %d", len(diags), test.diagCount)
				for _, diag := range diags {
					t.Logf(" - %s", diag.Error())
				}
			}

			if !got.RawEquals(test.want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.want)
			}
		})
	}
}
