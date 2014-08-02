package ast

import (
	"reflect"
	"testing"
)

func TestAssignmentNode_accept(t *testing.T) {
	n := AssignmentNode{
		Key:   "foo",
		Value: LiteralNode{Value: "foo"},
	}

	expected := []Node{
		n.Value,
		n,
	}

	v := new(MockVisitor)
	n.Accept(v)

	if !reflect.DeepEqual(v.Nodes, expected) {
		t.Fatalf("bad: %#v", v.Nodes)
	}
}

func TestListNode_accept(t *testing.T) {
	n := ListNode{
		Elem: []Node{
			LiteralNode{Value: "foo"},
			LiteralNode{Value: "bar"},
		},
	}

	expected := []Node{
		n.Elem[0],
		n.Elem[1],
		n,
	}

	v := new(MockVisitor)
	n.Accept(v)

	if !reflect.DeepEqual(v.Nodes, expected) {
		t.Fatalf("bad: %#v", v.Nodes)
	}
}

func TestObjectNode_accept(t *testing.T) {
	n := ObjectNode{
		Key: "foo",
		Elem: []Node{
			LiteralNode{Value: "foo"},
			LiteralNode{Value: "bar"},
		},
	}

	expected := []Node{
		n.Elem[0],
		n.Elem[1],
		n,
	}

	v := new(MockVisitor)
	n.Accept(v)

	if !reflect.DeepEqual(v.Nodes, expected) {
		t.Fatalf("bad: %#v", v.Nodes)
	}
}
