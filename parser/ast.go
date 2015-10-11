package parser

import "github.com/fatih/hcl/scanner"

type NodeType int

const (
	Unknown NodeType = 0
	Number
	Float
	Bool
	String
	List
	Object
)

// Node is an element in the abstract syntax  tree.
type Node interface {
	node()
	String() string
	Pos() scanner.Pos
}

func (IdentStatement) node()  {}
func (BlockStatement) node()  {}
func (AssignStatement) node() {}
func (ListStatement) node()   {}
func (ObjectStatement) node() {}

// IdentStatement represents an identifier.
type IdentStatement struct {
	pos   scanner.Pos // position of the literal
	token scanner.Token
	value string
}

func (i IdentStatement) String() string {
	return i.value
}

func (i IdentStatement) Pos() scanner.Pos {
	return i.pos
}

type BlockStatement struct {
	lbrace scanner.Pos // position of "{"
	rbrace scanner.Pos // position of "}"
	list   []Node      // the nodes in lexical order
}

func (b BlockStatement) String() string {
	s := "{\n"
	for _, n := range b.list {
		s += n.String() + "\n"
	}

	s += "}"
	return s
}

func (b BlockStatement) Pos() scanner.Pos {
	return b.lbrace
}

// AssignStatement represents an assignment
type AssignStatement struct {
	lhs    Node        // left hand side of the assignment
	rhs    Node        // right hand side of the assignment
	assign scanner.Pos // position of "="
}

func (a AssignStatement) String() string {
	return a.lhs.String() + " = " + a.rhs.String()
}

func (a AssignStatement) Pos() scanner.Pos {
	return a.lhs.Pos()
}

// ListStatement represents a list
type ListStatement struct {
	lbrack scanner.Pos // position of "["
	rbrack scanner.Pos // position of "]"
	list   []Node      // the elements in lexical order
}

func (l ListStatement) String() string {
	s := "[\n"
	for _, n := range l.list {
		s += n.String() + ",\n"
	}

	s += "]"
	return s
}

func (l ListStatement) Pos() scanner.Pos {
	return l.lbrack
}

// ObjectStatment represents an object
type ObjectStatement struct {
	Idents []Node // the idents in elements in lexical order
	BlockStatement
}

func (o ObjectStatement) String() string {
	s := ""

	for i, n := range o.Idents {
		s += n.String()
		if i != len(o.Idents) {
			s += " "
		}
	}

	s += o.BlockStatement.String()
	return s
}

func (o ObjectStatement) Pos() scanner.Pos {
	return o.Idents[0].Pos()
}
