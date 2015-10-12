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

// Node is an element in the abstract syntax tree.
type Node interface {
	node()
	String() string
	Pos() scanner.Pos
}

func (Source) node() {}
func (Ident) node()  {}

func (AssignStatement) node() {}
func (ObjectStatement) node() {}

func (LiteralType) node() {}
func (ObjectType) node()  {}
func (ListType) node()    {}

// Source represents a single HCL source file
type Source struct {
	nodes []Node
}

func (s *Source) add(node Node) {
	s.nodes = append(s.nodes, node)
}

func (s *Source) String() string {
	buf := ""
	for _, n := range s.nodes {
		buf += n.String()
	}

	return buf
}

func (s *Source) Pos() scanner.Pos {
	// always returns the uninitiliazed position
	return s.nodes[0].Pos()
}

// IdentStatement represents an identifier.
type Ident struct {
	token scanner.Token
}

func (i *Ident) String() string {
	return i.token.Text
}

func (i *Ident) Pos() scanner.Pos {
	return i.token.Pos
}

// AssignStatement represents an assignment
type AssignStatement struct {
	lhs    Node        // left hand side of the assignment
	rhs    Node        // right hand side of the assignment
	assign scanner.Pos // position of "="
}

func (a *AssignStatement) String() string {
	return a.lhs.String() + " = " + a.rhs.String()
}

func (a *AssignStatement) Pos() scanner.Pos {
	return a.lhs.Pos()
}

// ObjectStatment represents an object statement
type ObjectStatement struct {
	Idents []Node // the idents in elements in lexical order
	ObjectType
}

func (o *ObjectStatement) String() string {
	s := ""

	for i, n := range o.Idents {
		s += n.String()
		if i != len(o.Idents) {
			s += " "
		}
	}

	s += o.ObjectType.String()
	return s
}

func (o *ObjectStatement) Pos() scanner.Pos {
	return o.Idents[0].Pos()
}

// LiteralType represents a literal of basic type. Valid types are:
// scanner.NUMBER, scanner.FLOAT, scanner.BOOL and scanner.STRING
type LiteralType struct {
	token scanner.Token
}

func (l *LiteralType) String() string {
	return l.token.Text
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

func (l *ListType) String() string {
	s := "[\n"
	for _, n := range l.list {
		s += n.String() + ",\n"
	}

	s += "]"
	return s
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

func (b *ObjectType) String() string {
	s := "{\n"
	for _, n := range b.list {
		s += n.String() + "\n"
	}

	s += "}"
	return s
}

func (b *ObjectType) Pos() scanner.Pos {
	return b.lbrace
}
