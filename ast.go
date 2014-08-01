package hcl

// ValueType is an enum represnting the type of a value in
// a LiteralNode.
type ValueType byte

const (
	ValueTypeUnknown ValueType = iota
	ValueTypeInt
	ValueTypeString
)

// Node is implemented by all AST nodes for HCL.
type Node interface {
	Accept(Visitor)
}

// Visitor is the interface that must be implemented by any
// structures who want to be visited as part of the visitor pattern
// on the AST.
type Visitor interface {
	Visit(Node)
}

// ObjectNode represents an object that has multiple elements.
// An object's elements may repeat (keys). This is expected to
// be validated/removed at a semantic check, rather than at a
// syntax level.
type ObjectNode struct {
	Key  string
	Elem []Node
}

// AssignmentNode represents a direct assignment with an equals
// sign.
type AssignmentNode struct {
	Key   string
	Value Node
}

// ListNode represents a list or array of items.
type ListNode struct {
	Elem []Node
}

// LiteralNode is a direct value.
type LiteralNode struct {
	Type  ValueType
	Value interface{}
}

func (n ObjectNode) Accept(v Visitor) {
	for _, e := range n.Elem {
		e.Accept(v)
	}

	v.Visit(n)
}

func (n AssignmentNode) Accept(v Visitor) {
	n.Value.Accept(v)
	v.Visit(n)
}

func (n ListNode) Accept(v Visitor) {
	for _, e := range n.Elem {
		e.Accept(v)
	}

	v.Visit(n)
}

func (n LiteralNode) Accept(v Visitor) {
	v.Visit(n)
}
