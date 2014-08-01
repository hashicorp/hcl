package hcl

type ValueType byte

const (
	ValueTypeUnknown ValueType = iota
	ValueTypeInt
	ValueTypeString
)

type Node interface{}

type ObjectNode struct {
	Key  string
	Elem []Node
}

type AssignmentNode struct {
	Key   string
	Value Node
}

type ListNode struct {
	Elem []Node
}

type LiteralNode struct {
	Type  ValueType
	Value interface{}
}
