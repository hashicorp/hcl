package gohcl

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"reflect"
)

// EvalContext constructs an expression evaluation context from a Go struct value,
// making the fields available as variables and the methods available as functions,
// after transforming the field and method names such that each word (starting with
// an uppercase letter) is all lowercase and separated by underscores.
//
// Cause of Functions variable are implemented by special stdlib functions,
// this function could not evaluation golang native function variable
func EvalContext(v interface{}) *hcl.EvalContext {
	return &hcl.EvalContext{
		Variables: structMapVal(v),
	}
}

// structMapVal use reflect to traverse the struct,
// input could be a pointer,it would check the source
// struct, and return a map of cty.Value.
func structMapVal(v interface{}) map[string]cty.Value {
	rt := reflect.TypeOf(v)
	rv := reflect.ValueOf(v)

	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	var variables = make(map[string]cty.Value)

	for index := 0; index < rt.NumField(); index++ {
		key := rt.Field(index)
		value := rv.Field(index)

		if !value.IsZero() {
			k := marshalKey(key.Name)
			refVal, err := reflectVal(value)
			if err == nil {
				variables[k] = refVal
			}
		}

	}
	return variables

}

// reflectVal receive a reflect.Value and according to the kind implemented,
// return a cty.Value. The value kind that have been implemented so far are
// Int/Uint, Float, String, and nest Struct and Slice
func reflectVal(v reflect.Value) (cty.Value, error) {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return cty.NumberIntVal(v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return cty.NumberUIntVal(v.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return cty.NumberFloatVal(v.Float()), nil
	case reflect.String:
		return cty.StringVal(v.String()), nil
	case reflect.Struct:
		return structVal(v)
	case reflect.Slice:
		return listVal(v)
	case reflect.Map:
		return mapVal(v)
	default:
		return cty.EmptyObjectVal, fmt.Errorf("target value must be pointer to number, string, slice, struct or map, not %s", v.String())
	}
}

func mapVal(v reflect.Value) (cty.Value, error) {
	var mapVar = make(map[string]cty.Value)
	kind := v.Type().Key().Kind()
	if kind == reflect.String {
		iter := v.MapRange()
		for iter.Next() {
			refVal, err := reflectVal(iter.Value())
			if err == nil {
				mapVar[marshalKey(iter.Key().String())] = refVal
			}
		}
		if len(mapVar) > 0 {
			return cty.MapVal(mapVar), nil
		}
		return cty.EmptyObjectVal, fmt.Errorf("target map error or is empty")
	}
	return cty.EmptyObjectVal, fmt.Errorf("target key should be string, %s is not support", kind.String())
}

// listVal receive a reflect.Value which should be asserted as Slice type.
// In the for loop, each var would be called by func reflectVal to return
// a cty.Value and add into a slice.Finally return cty.ListVal
func listVal(v reflect.Value) (cty.Value, error) {
	elems := []cty.Value{}
	for i := 0; i < v.Len(); i++ {
		refVal, err := reflectVal(v.Index(i))
		if err == nil {
			elems = append(elems, refVal)
		}
	}
	if len(elems) > 0 {
		return cty.ListVal(elems), nil
	}
	return cty.EmptyTupleVal, fmt.Errorf(" slice is empty, cty.ListVal must not call ListVal with empty slice")
}

// structVal received a reflect.Value which should be asserted as Struct type.
// It uses the NumFiled() of  reflect type to loop all struct fields,
// and return cty.MapVal

func structVal(v reflect.Value) (cty.Value, error) {
	var ctyVals = make(map[string]cty.Value)
	for index := 0; index < v.Type().NumField(); index++ {
		key := v.Type().Field(index)
		value := v.Field(index)
		refVal, err := reflectVal(value)
		if err == nil {
			ctyVals[marshalKey(key.Name)] = refVal
		}
	}
	if len(ctyVals) > 0 {
		return cty.MapVal(ctyVals), nil
	}
	return cty.EmptyObjectVal, fmt.Errorf(" slice is empty, cty.ListVal must not call ListVal with empty map")
}

// marshalKey trans camelcase to lowercase with separated by underscores
func marshalKey(input string) string {
	if input == "" {
		return ""
	}
	var output bytes.Buffer
	for index, letter := range input {
		if letter < 96 {
			letter = letter + 32
			if index > 0 {
				output.WriteString("_")
			}

		}
		output.WriteRune(letter)
	}
	return output.String()
}
