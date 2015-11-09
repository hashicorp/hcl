package ast

import "fmt"

// Walk traverses an AST in depth-first order: It starts by calling fn(node);
// node must not be nil.  If f returns true, Walk invokes f recursively for
// each of the non-nil children of node, followed by a call of f(nil).
func Walk(node Node, fn func(Node) bool) {
	if !fn(node) {
		return
	}

	switch n := node.(type) {
	case *File:
		Walk(n.Node, fn)
	case *ObjectList:
		for _, item := range n.Items {
			Walk(item, fn)
		}
	case *ObjectKey:
		// nothing to do
	case *ObjectItem:
		for _, k := range n.Keys {
			Walk(k, fn)
		}
		Walk(n.Val, fn)
	case *LiteralType:
		// nothing to do
	case *ListType:
		for _, l := range n.List {
			Walk(l, fn)
		}
	case *ObjectType:
		Walk(n.List, fn)
	default:
		fmt.Printf(" unknown type: %T\n", n)
	}

	fn(nil)
}
