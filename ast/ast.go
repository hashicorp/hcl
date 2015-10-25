// Package ast declares the types used to represent syntax trees for HCL
// (HashiCorp Configuration Language)
package ast

import "github.com/fatih/hcl/token"

// Node is an element in the abstract syntax tree.
type Node interface {
	node()
	Pos() token.Pos
}

func (ObjectList) node() {}
func (ObjectKey) node()  {}
func (ObjectItem) node() {}

func (Comment) node()     {}
func (ObjectType) node()  {}
func (LiteralType) node() {}
func (ListType) node()    {}

// ObjectList represents a list of ObjectItems. An HCL file itself is an
// ObjectList.
type ObjectList struct {
	Items []*ObjectItem
}

func (o *ObjectList) Add(item *ObjectItem) {
	o.Items = append(o.Items, item)
}

func (o *ObjectList) Pos() token.Pos {
	// always returns the uninitiliazed position
	return o.Items[0].Pos()
}

// ObjectItem represents a HCL Object Item. An item is represented with a key
// (or keys). It can be an assignment or an object (both normal and nested)
type ObjectItem struct {
	// keys is only one length long if it's of type assignment. If it's a
	// nested object it can be larger than one. In that case "assign" is
	// invalid as there is no assignments for a nested object.
	Keys []*ObjectKey

	// assign contains the position of "=", if any
	Assign token.Pos

	// val is the item itself. It can be an object,list, number, bool or a
	// string. If key length is larger than one, val can be only of type
	// Object.
	Val Node
}

func (o *ObjectItem) Pos() token.Pos {
	return o.Keys[0].Pos()
}

// ObjectKeys are either an identifier or of type string.
type ObjectKey struct {
	Token token.Token
}

func (o *ObjectKey) Pos() token.Pos {
	return o.Token.Pos
}

// LiteralType represents a literal of basic type. Valid types are:
// token.NUMBER, token.FLOAT, token.BOOL and token.STRING
type LiteralType struct {
	Token token.Token
}

func (l *LiteralType) Pos() token.Pos {
	return l.Token.Pos
}

// ListStatement represents a HCL List type
type ListType struct {
	Lbrack token.Pos // position of "["
	Rbrack token.Pos // position of "]"
	List   []Node    // the elements in lexical order
}

func (l *ListType) Pos() token.Pos {
	return l.Lbrack
}

func (l *ListType) Add(node Node) {
	l.List = append(l.List, node)
}

// ObjectType represents a HCL Object Type
type ObjectType struct {
	Lbrace token.Pos   // position of "{"
	Rbrace token.Pos   // position of "}"
	List   *ObjectList // the nodes in lexical order
}

func (o *ObjectType) Pos() token.Pos {
	return o.Lbrace
}

// Comment node represents a single //, # style or /*- style commment
type Comment struct {
	Start token.Pos // position of / or #
	Text  string
}

func (c *Comment) Pos() token.Pos {
	return c.Start
}
