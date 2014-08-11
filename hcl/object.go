package hcl

import (
	"strings"
)

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

// Get gets all the objects that match the given key.
//
// It returns the resulting objects as a single Object structure with
// the linked list populated.
func (o *Object) Get(k string, insensitive bool) *Object {
	if o.Type != ValueTypeObject {
		return nil
	}

	var current, result *Object
	m := o.Value.(map[string]*Object)
	for _, o := range m {
		if o.Key != k {
			if !insensitive || !strings.EqualFold(o.Key, k) {
				continue
			}
		}

		o2 := *o
		o2.Next = nil
		if result == nil {
			result = &o2
			current = result
		} else {
			current.Next = &o2
			current = current.Next
		}
	}

	return result
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
