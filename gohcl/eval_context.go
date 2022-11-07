package gohcl

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"reflect"
)

func EvalContext(v interface{}) *hcl.EvalContext {
	return &hcl.EvalContext{
		Variables: structMapVal(v),
	}
}

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
			//k := key.Name
			variables[k] = reflectVal(value)
		}

	}
	return variables

}

func reflectVal(v reflect.Value) cty.Value {
	switch v.Kind() {
	case reflect.Int:
		return cty.NumberIntVal(v.Int())
	case reflect.String:
		return cty.StringVal(v.String())
	case reflect.Struct:
		return structVal(v)
	case reflect.Slice:
		return sliceVal(v)
	default:
		panic(fmt.Sprintf("target value must be pointer to int, string, slice, struct or map, not %s", v.String()))
	}
}

func sliceVal(v reflect.Value) cty.Value {
	elems := []cty.Value{}
	for i := 0; i < v.Len(); i++ {
		elems = append(elems, reflectVal(v.Index(i)))
	}
	return cty.TupleVal(elems)
}

func structVal(v reflect.Value) cty.Value {
	var ctyVals = make(map[string]cty.Value)
	for index := 0; index < v.Type().NumField(); index++ {
		key := v.Type().Field(index)
		value := v.Field(index)
		ctyVals[marshalKey(key.Name)] = reflectVal(value)
	}
	return cty.MapVal(ctyVals)
}

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
