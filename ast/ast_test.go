package ast

import (
	"reflect"
	"testing"
)

func TestAssignmentNode_accept(t *testing.T) {
	n := AssignmentNode{
		K:     "foo",
		Value: LiteralNode{Value: "foo"},
	}

	expected := []Node{
		n,
		n.Value,
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
		n,
		n.Elem[0],
		n.Elem[1],
	}

	v := new(MockVisitor)
	n.Accept(v)

	if !reflect.DeepEqual(v.Nodes, expected) {
		t.Fatalf("bad: %#v", v.Nodes)
	}
}

func TestObjectNode_accept(t *testing.T) {
	n := ObjectNode{
		K: "foo",
		Elem: []AssignmentNode{
			AssignmentNode{K: "foo", Value: LiteralNode{Value: "foo"}},
			AssignmentNode{K: "bar", Value: LiteralNode{Value: "bar"}},
		},
	}

	expected := []Node{
		n,
		n.Elem[0],
		n.Elem[0].Value,
		n.Elem[1],
		n.Elem[1].Value,
	}

	v := new(MockVisitor)
	n.Accept(v)

	if !reflect.DeepEqual(v.Nodes, expected) {
		t.Fatalf("bad: %#v", v.Nodes)
	}
}

func TestObjectNodeGet(t *testing.T) {
	n := ObjectNode{
		K: "foo",
		Elem: []AssignmentNode{
			AssignmentNode{K: "foo", Value: LiteralNode{Value: "foo"}},
			AssignmentNode{K: "bar", Value: LiteralNode{Value: "bar"}},
			AssignmentNode{K: "foo", Value: LiteralNode{Value: "baz"}},
		},
	}

	expected := []Node{
		LiteralNode{Value: "foo"},
		LiteralNode{Value: "baz"},
	}

	actual := n.Get("foo", false)

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("bad: %#v", actual)
	}
}
