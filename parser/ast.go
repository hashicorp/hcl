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

// Node is an element in the parse tree.
type Node interface {
	String() string
	Type() NodeType
	Pos() scanner.Pos
	End() scanner.Pos
}

// IdentStatement represents an identifier.
type IdentStatement struct {
	Token scanner.Token
	Pos   scanner.Pos // position of the literal
	Value string
}

type BlockStatement struct {
	Lbrace scanner.Pos // position of "{"
	Rbrace scanner.Pos // position of "}"
	List   []Node      // the nodes in lexical order
}

// AssignStatement represents an assignment
type AssignStatement struct {
	Lhs    Node        // left hand side of the assignment
	Rhs    Node        // right hand side of the assignment
	Assign scanner.Pos // position of "="
}

// ListStatement represents a list
type ListStatement struct {
	Lbrack scanner.Pos // position of "["
	Rbrack scanner.Pos // position of "]"
	List   []Node      // the elements in lexical order
}

// ObjectStatment represents an object
type ObjectStatement struct {
	Idents []Node // the idents in elements in lexical order
	BlockStatement
}
