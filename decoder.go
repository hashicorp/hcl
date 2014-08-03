package hcl

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/hcl/ast"
)

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
func DecodeAST(out interface{}, obj *ast.ObjectNode) error {
	var d decoder
	return d.decode("root", *obj, reflect.ValueOf(out).Elem())
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

	switch k.Kind() {
	case reflect.Int:
		return d.decodeInt(name, n, result)
	case reflect.Interface:
		// When we see an interface, we make our own thing
		return d.decodeInterface(name, n, result)
	case reflect.Map:
		return d.decodeMap(name, n, result)
	case reflect.Slice:
		return d.decodeSlice(name, n, result)
	case reflect.String:
		return d.decodeString(name, n, result)
	default:
		return fmt.Errorf("%s: unknown kind: %s", name, result.Kind())
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
		result.Set(reflect.ValueOf(int64(n.Value.(int))))
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
	obj, ok := raw.(ast.ObjectNode)
	if !ok {
		return fmt.Errorf("%s: not an object type", name)
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
	case ast.ValueTypeString:
		result.Set(reflect.ValueOf(n.Value.(string)))
	default:
		return fmt.Errorf("%s: unknown type %s", name, n.Type)
	}

	return nil
}
