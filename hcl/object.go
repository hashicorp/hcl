package hcl

// ValueType is an enum represnting the type of a value in
// a LiteralNode.
type ValueType byte

const (
	ValueTypeUnknown ValueType = iota
	ValueTypeFloat
	ValueTypeInt
	ValueTypeString
	ValueTypeBool
	ValueTypeNil
	ValueTypeList
	ValueTypeObject
)

// Object represents any element of HCL: an object itself, a list,
// a literal, etc.
type Object struct {
	Key   string
	Type  ValueType
	Value interface{}
	Next  *Object
}

// ObjectList is a list of objects.
type ObjectList []*Object

// Map returns a flattened map structure of the list of objects.
func (l ObjectList) Map() map[string]*Object {
	m := make(map[string]*Object)
	for _, obj := range l {
		prev, ok := m[obj.Key]
		if !ok {
			m[obj.Key] = obj
			continue
		}

		for prev.Next != nil {
			prev = prev.Next
		}
		prev.Next = obj
	}

	return m
}
