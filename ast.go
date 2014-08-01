package hcl

type ValueType byte

const (
	ValueTypeUnknown ValueType = iota
	ValueTypeInt
	ValueTypeString
)

type Node interface{}

type ObjectNode struct {
	Elem map[string][]Node
}

type ValueNode struct {
	Type  ValueType
	Value interface{}
}
