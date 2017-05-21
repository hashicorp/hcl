package gozcl

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/apparentlymart/go-cty/cty"
	"github.com/apparentlymart/go-zcl/zcl"
)

func TestDecodeExpression(t *testing.T) {
	tests := []struct {
		Value     cty.Value
		Target    interface{}
		Want      interface{}
		DiagCount int
	}{
		{
			cty.StringVal("hello"),
			"",
			"hello",
			0,
		},
		{
			cty.StringVal("hello"),
			cty.NilVal,
			cty.StringVal("hello"),
			0,
		},
		{
			cty.NumberIntVal(2),
			"",
			"2",
			0,
		},
		{
			cty.StringVal("true"),
			false,
			true,
			0,
		},
		{
			cty.NullVal(cty.String),
			"",
			"",
			1, // null value is not allowed
		},
		{
			cty.UnknownVal(cty.String),
			"",
			"",
			1, // value must be known
		},
		{
			cty.ListVal([]cty.Value{cty.True}),
			false,
			false,
			1, // bool required
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			expr := &fixedExpression{test.Value}

			targetVal := reflect.New(reflect.TypeOf(test.Target))

			diags := DecodeExpression(expr, nil, targetVal.Interface())
			if len(diags) != test.DiagCount {
				t.Errorf("wrong number of diagnostics %d; want %d", len(diags), test.DiagCount)
				for _, diag := range diags {
					t.Logf(" - %s", diag.Error())
				}
			}
			got := targetVal.Elem().Interface()
			if !reflect.DeepEqual(got, test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

type fixedExpression struct {
	val cty.Value
}

func (e *fixedExpression) Value(ctx *zcl.EvalContext) (cty.Value, zcl.Diagnostics) {
	return e.val, nil
}

func (e *fixedExpression) Range() (r zcl.Range) {
	return
}
func (e *fixedExpression) StartRange() (r zcl.Range) {
	return
}
