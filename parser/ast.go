package parser

import "github.com/fatih/hcl/scanner"

// Node is an element in the abstract syntax tree.
type Node interface {
	node()
	Pos() scanner.Pos
}

func (ObjectList) node()  {}
func (ObjectItem) node()  {}
func (ObjectType) node()  {}
func (LiteralType) node() {}
func (ListType) node()    {}
func (Ident) node()       {}

// ObjectList represents a list of ObjectItems. An HCL file itself is an
// ObjectList.
type ObjectList struct {
	items []*ObjectItem
}

func (o *ObjectList) add(item *ObjectItem) {
	o.items = append(o.items, item)
}

func (o *ObjectList) Pos() scanner.Pos {
	// always returns the uninitiliazed position
	return o.items[0].Pos()
}

// ObjectItem represents a HCL Object Item. An item is represented with a key
// (or keys). It can be an assignment or an object (both normal and nested)
type ObjectItem struct {
	// key is either an Identifier or a String. The slice is only one lenght
	// long, however if it's a nested object it'll can be larger than one. In
	// that case "assign" is invalid as there is no assignments for a nested
	// object.
	key []Ident

	// assign contains the position of "=", if any
	assign scanner.Pos

	// val is the item itself. It can be an object,list, number, bool or a
	// string. If key lenght is larger than one, val can be only of type
	// Object.
	val Node
}

func (o *ObjectItem) Pos() scanner.Pos {
	return o.key[0].Pos()
}

// IdentStatement represents an identifier.
type Ident struct {
	token scanner.Token
}

func (i *Ident) Pos() scanner.Pos {
	return i.token.Pos
}

// LiteralType represents a literal of basic type. Valid types are:
// scanner.NUMBER, scanner.FLOAT, scanner.BOOL and scanner.STRING
type LiteralType struct {
	token scanner.Token
}

// isValid() returns true if the underlying identifier satisfies one of the
// valid types.
func (l *LiteralType) isValid() bool {
	switch l.token.Type {
	case scanner.NUMBER, scanner.FLOAT, scanner.BOOL, scanner.STRING:
		return true
	default:
		return false
	}
}

func (l *LiteralType) Pos() scanner.Pos {
	return l.token.Pos
}

// ListStatement represents a HCL List type
type ListType struct {
	lbrack scanner.Pos // position of "["
	rbrack scanner.Pos // position of "]"
	list   []Node      // the elements in lexical order
}

func (l *ListType) Pos() scanner.Pos {
	return l.lbrack
}

// ObjectType represents a HCL Object Type
type ObjectType struct {
	lbrace scanner.Pos // position of "{"
	rbrace scanner.Pos // position of "}"
	list   []Node      // the nodes in lexical order
}

func (b *ObjectType) Pos() scanner.Pos {
	return b.lbrace
}
