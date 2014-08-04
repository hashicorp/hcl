package hcl

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl/ast"
)

// This is the tag to use with structures to have settings for HCL
const tagName = "hcl"

// Decode reads the given input and decodes it into the structure
// given by `out`.
func Decode(out interface{}, in string) error {
	obj, err := Parse(in)
	if err != nil {
		return err
	}

	return DecodeAST(out, obj)
}

// DecodeAST is a lower-level version of Decode. It decodes a
// raw AST into the given output.
func DecodeAST(out interface{}, n ast.Node) error {
	var d decoder
	return d.decode("root", n, reflect.ValueOf(out).Elem())
}

type decoder struct{}

func (d *decoder) decode(name string, n ast.Node, result reflect.Value) error {
	k := result

	// If we have an interface with a valid value, we use that
	// for the check.
	if result.Kind() == reflect.Interface {
		elem := result.Elem()
		if elem.IsValid() {
			k = elem
		}
	}

	// If we have a pointer, unpointer it
	switch rn := n.(type) {
	case *ast.ObjectNode:
		n = *rn
	}

	switch k.Kind() {
	case reflect.Int:
		return d.decodeInt(name, n, result)
	case reflect.Interface:
		// When we see an interface, we make our own thing
		return d.decodeInterface(name, n, result)
	case reflect.Map:
		return d.decodeMap(name, n, result)
	case reflect.Ptr:
		return d.decodePtr(name, n, result)
	case reflect.Slice:
		return d.decodeSlice(name, n, result)
	case reflect.String:
		return d.decodeString(name, n, result)
	case reflect.Struct:
		return d.decodeStruct(name, n, result)
	default:
		return fmt.Errorf(
			"%s: unknown kind to decode into: %s", name, result.Kind())
	}

	return nil
}

func (d *decoder) decodeInt(name string, raw ast.Node, result reflect.Value) error {
	n, ok := raw.(ast.LiteralNode)
	if !ok {
		return fmt.Errorf("%s: not a literal type", name)
	}

	switch n.Type {
	case ast.ValueTypeInt:
		result.Set(reflect.ValueOf(n.Value.(int)))
	default:
		return fmt.Errorf("%s: unknown type %s", name, n.Type)
	}

	return nil
}

func (d *decoder) decodeInterface(name string, raw ast.Node, result reflect.Value) error {
	var set reflect.Value
	redecode := true

	switch n := raw.(type) {
	case ast.ObjectNode:
		var temp map[string]interface{}
		tempVal := reflect.ValueOf(temp)
		result := reflect.MakeMap(
			reflect.MapOf(
				reflect.TypeOf(""),
				tempVal.Type().Elem()))

		set = result
	case ast.ListNode:
		/*
			var temp []interface{}
			tempVal := reflect.ValueOf(temp)
			result := reflect.MakeSlice(
				reflect.SliceOf(tempVal.Type().Elem()), 0, 0)
			set = result
		*/
		redecode = false
		result := make([]interface{}, 0, len(n.Elem))

		for _, elem := range n.Elem {
			raw := new(interface{})
			err := d.decode(
				name, elem, reflect.Indirect(reflect.ValueOf(raw)))
			if err != nil {
				return err
			}

			result = append(result, *raw)
		}

		set = reflect.ValueOf(result)
	case ast.LiteralNode:
		switch n.Type {
		case ast.ValueTypeInt:
			var result int
			set = reflect.Indirect(reflect.New(reflect.TypeOf(result)))
		case ast.ValueTypeString:
			set = reflect.Indirect(reflect.New(reflect.TypeOf("")))
		default:
			return fmt.Errorf(
				"%s: unknown literal type: %s",
				name, n.Type)
		}
	default:
		return fmt.Errorf(
			"%s: cannot decode into interface: %T",
			name, raw)
	}

	// Set the result to what its supposed to be, then reset
	// result so we don't reflect into this method anymore.
	result.Set(set)

	if redecode {
		// Revisit the node so that we can use the newly instantiated
		// thing and populate it.
		if err := d.decode(name, raw, result); err != nil {
			return err
		}
	}

	return nil
}

func (d *decoder) decodeMap(name string, raw ast.Node, result reflect.Value) error {
	// If we have a list, then we decode each element into a map
	if list, ok := raw.(ast.ListNode); ok {
		for i, elem := range list.Elem {
			fieldName := fmt.Sprintf("%s.%d", name, i)
			err := d.decode(fieldName, elem, result)
			if err != nil {
				return err
			}
		}

		return nil
	}

	obj, ok := raw.(ast.ObjectNode)
	if !ok {
		return fmt.Errorf("%s: not an object type (%T)", name, obj)
	}

	// If we have an interface, then we can address the interface,
	// but not the slice itself, so get the element but set the interface
	set := result
	if result.Kind() == reflect.Interface {
		result = result.Elem()
	}

	resultType := result.Type()
	resultElemType := resultType.Elem()
	resultKeyType := resultType.Key()
	if resultKeyType.Kind() != reflect.String {
		return fmt.Errorf(
			"%s: map must have string keys", name)
	}

	// Make a map if it is nil
	resultMap := result
	if result.IsNil() {
		resultMap = reflect.MakeMap(
			reflect.MapOf(resultKeyType, resultElemType))
	}

	// Go through each element and decode it.
	for _, elem := range obj.Elem {
		n := elem.(ast.AssignmentNode)
		objValue := n.Value

		// If we have an object node, expand to a list of objects
		if _, ok := objValue.(ast.ObjectNode); ok {
			objValue = ast.ListNode{
				Elem: []ast.Node{objValue},
			}
		}

		// Make the field name
		fieldName := fmt.Sprintf("%s.%s", name, n.Key())

		// Get the key/value as reflection values
		key := reflect.ValueOf(n.Key())
		val := reflect.Indirect(reflect.New(resultElemType))

		// If we have a pre-existing value in the map, use that
		oldVal := resultMap.MapIndex(key)
		if oldVal.IsValid() {
			val.Set(oldVal)
		}

		// Decode!
		if err := d.decode(fieldName, objValue, val); err != nil {
			return err
		}

		// Set the value on the map
		resultMap.SetMapIndex(key, val)
	}

	// Set the final map if we can
	set.Set(resultMap)
	return nil
}

func (d *decoder) decodePtr(name string, raw ast.Node, result reflect.Value) error {
	// Create an element of the concrete (non pointer) type and decode
	// into that. Then set the value of the pointer to this type.
	resultType := result.Type()
	resultElemType := resultType.Elem()
	val := reflect.New(resultElemType)
	if err := d.decode(name, raw, reflect.Indirect(val)); err != nil {
		return err
	}

	result.Set(val)
	return nil
}

func (d *decoder) decodeSlice(name string, raw ast.Node, result reflect.Value) error {
	n, ok := raw.(ast.ListNode)
	if !ok {
		return fmt.Errorf("%s: not a list type", name)
	}

	// If we have an interface, then we can address the interface,
	// but not the slice itself, so get the element but set the interface
	set := result
	if result.Kind() == reflect.Interface {
		result = result.Elem()
	}

	// Create the slice if it isn't nil
	resultType := result.Type()
	resultElemType := resultType.Elem()
	if result.IsNil() {
		resultSliceType := reflect.SliceOf(resultElemType)
		result = reflect.MakeSlice(
			resultSliceType, 0, 0)
	}

	for i, elem := range n.Elem {
		fieldName := fmt.Sprintf("%s[%d]", name, i)

		// Decode
		val := reflect.Indirect(reflect.New(resultElemType))
		if err := d.decode(fieldName, elem, val); err != nil {
			return err
		}

		// Append it onto the slice
		result = reflect.Append(result, val)
	}

	set.Set(result)
	return nil
}

func (d *decoder) decodeString(name string, raw ast.Node, result reflect.Value) error {
	n, ok := raw.(ast.LiteralNode)
	if !ok {
		return fmt.Errorf("%s: not a literal type", name)
	}

	switch n.Type {
	case ast.ValueTypeInt:
		result.Set(reflect.ValueOf(
			strconv.FormatInt(int64(n.Value.(int)), 10)))
	case ast.ValueTypeString:
		result.Set(reflect.ValueOf(n.Value.(string)))
	default:
		return fmt.Errorf("%s: unknown type to string: %s", name, n.Type)
	}

	return nil
}

func (d *decoder) decodeStruct(name string, raw ast.Node, result reflect.Value) error {
	// If we have a list, then we decode each element into a map
	if list, ok := raw.(ast.ListNode); ok {
		for i, elem := range list.Elem {
			fieldName := fmt.Sprintf("%s.%d", name, i)
			err := d.decode(fieldName, elem, result)
			if err != nil {
				return err
			}
		}

		return nil
	}

	obj, ok := raw.(ast.ObjectNode)
	if !ok {
		return fmt.Errorf(
			"%s: not an object type for struct (%T)", name, raw)
	}

	// This slice will keep track of all the structs we'll be decoding.
	// There can be more than one struct if there are embedded structs
	// that are squashed.
	structs := make([]reflect.Value, 1, 5)
	structs[0] = result

	// Compile the list of all the fields that we're going to be decoding
	// from all the structs.
	fields := make(map[*reflect.StructField]reflect.Value)
	for len(structs) > 0 {
		structVal := structs[0]
		structs = structs[1:]

		structType := structVal.Type()
		for i := 0; i < structType.NumField(); i++ {
			fieldType := structType.Field(i)

			if fieldType.Anonymous {
				fieldKind := fieldType.Type.Kind()
				if fieldKind != reflect.Struct {
					return fmt.Errorf(
						"%s: unsupported type to struct: %s",
						fieldType.Name, fieldKind)
				}

				// We have an embedded field. We "squash" the fields down
				// if specified in the tag.
				squash := false
				tagParts := strings.Split(fieldType.Tag.Get(tagName), ",")
				for _, tag := range tagParts[1:] {
					if tag == "squash" {
						squash = true
						break
					}
				}

				if squash {
					structs = append(
						structs, result.FieldByName(fieldType.Name))
					continue
				}
			}

			// Normal struct field, store it away
			fields[&fieldType] = structVal.Field(i)
		}
	}

	usedKeys := make(map[string]struct{})
	decodedFields := make([]string, 0, len(fields))
	decodedFieldsVal := make([]reflect.Value, 0)
	unusedKeysVal := make([]reflect.Value, 0)
	for fieldType, field := range fields {
		if !field.IsValid() {
			// This should never happen
			panic("field is not valid")
		}

		// If we can't set the field, then it is unexported or something,
		// and we just continue onwards.
		if !field.CanSet() {
			continue
		}

		fieldName := fieldType.Name

		tagValue := fieldType.Tag.Get(tagName)
		tagParts := strings.SplitN(tagValue, ",", 2)
		if len(tagParts) >= 2 {
			switch tagParts[1] {
			case "decodedFields":
				decodedFieldsVal = append(decodedFieldsVal, field)
				continue
			case "key":
				field.SetString(obj.Key())
				continue
			case "unusedKeys":
				unusedKeysVal = append(unusedKeysVal, field)
				continue
			}
		}

		if tagParts[0] != "" {
			fieldName = tagParts[0]
		}

		// Find the element matching this name
		elems := obj.Get(fieldName, true)
		if len(elems) == 0 {
			continue
		}

		// Track the used key
		usedKeys[fieldName] = struct{}{}

		// Create the field name and decode
		fieldName = fmt.Sprintf("%s.%s", name, fieldName)
		for _, elem := range elems {
			if err := d.decode(fieldName, elem, field); err != nil {
				return err
			}
		}

		decodedFields = append(decodedFields, fieldType.Name)
	}

	for _, v := range decodedFieldsVal {
		v.Set(reflect.ValueOf(decodedFields))
	}

	// If we want to know what keys are unused, compile that
	if len(unusedKeysVal) > 0 {
		unusedKeys := make([]string, 0, int(obj.Len())-len(usedKeys))

		for _, elem := range obj.Elem {
			k := elem.Key()
			if _, ok := usedKeys[k]; !ok {
				unusedKeys = append(unusedKeys, k)
			}
		}

		if len(unusedKeys) == 0 {
			unusedKeys = nil
		}

		for _, v := range unusedKeysVal {
			v.Set(reflect.ValueOf(unusedKeys))
		}
	}

	return nil
}
