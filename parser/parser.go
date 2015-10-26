// Package parser implements a parser for HCL (HashiCorp Configuration
// Language)
package parser

import (
	"errors"
	"fmt"

	"github.com/fatih/hcl/ast"
	"github.com/fatih/hcl/scanner"
	"github.com/fatih/hcl/token"
)

type Parser struct {
	sc *scanner.Scanner

	tok      token.Token // last read token
	comments []*ast.Comment

	enableTrace bool
	indent      int
	n           int // buffer size (max = 1)
}

func newParser(src []byte) *Parser {
	return &Parser{
		sc: scanner.New(src),
	}
}

// Parse returns the fully parsed source and returns the abstract syntax tree.
func Parse(src []byte) (ast.Node, error) {
	p := newParser(src)
	return p.Parse()
}

var errEofToken = errors.New("EOF token found")

// Parse returns the fully parsed source and returns the abstract syntax tree.
func (p *Parser) Parse() (ast.Node, error) {
	return p.objectList()
}

func (p *Parser) objectList() (*ast.ObjectList, error) {
	defer un(trace(p, "ParseObjectList"))
	node := &ast.ObjectList{}

	for {
		n, err := p.next()
		if err == errEofToken {
			break // we are finished
		}

		// we don't return a nil, because might want to use already collected
		// items.
		if err != nil {
			return node, err
		}

		switch t := n.(type) {
		case *ast.ObjectItem:
			node.Add(t)
		case *ast.Comment:
			p.comments = append(p.comments, t)
		}
	}

	return node, nil
}

// next returns the next node
func (p *Parser) next() (ast.Node, error) {
	defer un(trace(p, "ParseNode"))

	tok := p.scan()

	switch tok.Type {
	case token.EOF:
		return nil, errEofToken
	case token.IDENT, token.STRING:
		p.unscan()
		return p.objectItem()
	case token.COMMENT:
		return &ast.Comment{
			Start: tok.Pos,
			Text:  tok.Text,
		}, nil
	default:
		return nil, fmt.Errorf("expected: IDENT | STRING | COMMENT got: %+v", tok.Type)
	}
}

// objectItem parses a single object item
func (p *Parser) objectItem() (*ast.ObjectItem, error) {
	defer un(trace(p, "ParseObjectItem"))

	keys, err := p.objectKey()
	if err != nil {
		return nil, err
	}

	switch p.tok.Type {
	case token.ASSIGN:
		// assignments
		o := &ast.ObjectItem{
			Keys:   keys,
			Assign: p.tok.Pos,
		}

		o.Val, err = p.object()
		if err != nil {
			return nil, err
		}
		return o, nil
	case token.LBRACE:
		// object or nested objects
		o := &ast.ObjectItem{
			Keys: keys,
		}

		o.Val, err = p.objectType()
		if err != nil {
			return nil, err
		}
		return o, nil
	}

	return nil, fmt.Errorf("not yet implemented: %s", p.tok.Type)
}

// objectKey parses an object key and returns a ObjectKey AST
func (p *Parser) objectKey() ([]*ast.ObjectKey, error) {
	keyCount := 0
	keys := make([]*ast.ObjectKey, 0)

	for {
		tok := p.scan()
		switch tok.Type {
		case token.EOF:
			return nil, errEofToken
		case token.ASSIGN:
			// assignment or object only, but not nested objects. this is not
			// allowed: `foo bar = {}`
			if keyCount > 1 {
				return nil, fmt.Errorf("nested object expected: LBRACE got: %s", p.tok.Type)
			}

			if keyCount == 0 {
				return nil, errors.New("no keys found!!!")
			}

			return keys, nil
		case token.LBRACE:
			// object
			return keys, nil
		case token.IDENT, token.STRING:
			keyCount++
			keys = append(keys, &ast.ObjectKey{Token: p.tok})
		case token.ILLEGAL:
			fmt.Println("illegal")
		default:
			return nil, fmt.Errorf("expected: IDENT | STRING | ASSIGN | LBRACE got: %s", p.tok.Type)
		}
	}
}

// object parses any type of object, such as number, bool, string, object or
// list.
func (p *Parser) object() (ast.Node, error) {
	defer un(trace(p, "ParseType"))
	tok := p.scan()

	switch tok.Type {
	case token.NUMBER, token.FLOAT, token.BOOL, token.STRING:
		return p.literalType()
	case token.LBRACE:
		return p.objectType()
	case token.LBRACK:
		return p.listType()
	case token.COMMENT:
		// implement comment
	case token.EOF:
		return nil, errEofToken
	}

	return nil, fmt.Errorf("Unknown token: %+v", tok)
}

// ibjectType parses an object type and returns a ObjectType AST
func (p *Parser) objectType() (*ast.ObjectType, error) {
	defer un(trace(p, "ParseObjectType"))

	// we assume that the currently scanned token is a LBRACE
	o := &ast.ObjectType{
		Lbrace: p.tok.Pos,
	}

	l, err := p.objectList()

	// if we hit RBRACE, we are good to go (means we parsed all Items), if it's
	// not a RBRACE, it's an syntax error and we just return it.
	if err != nil && p.tok.Type != token.RBRACE {
		return nil, err
	}

	o.List = l
	o.Rbrace = p.tok.Pos // advanced via parseObjectList
	return o, nil
}

// listType parses a list type and returns a ListType AST
func (p *Parser) listType() (*ast.ListType, error) {
	defer un(trace(p, "ParseListType"))

	// we assume that the currently scanned token is a LBRACK
	l := &ast.ListType{
		Lbrack: p.tok.Pos,
	}

	for {
		tok := p.scan()
		switch tok.Type {
		case token.NUMBER, token.FLOAT, token.STRING:
			node, err := p.literalType()
			if err != nil {
				return nil, err
			}
			l.Add(node)
		case token.COMMA:
			// get next list item or we are at the end
			continue
		case token.COMMENT:
			// TODO(arslan): parse comment
			continue
		case token.BOOL:
			// TODO(arslan) should we support? not supported by HCL yet
		case token.LBRACK:
			// TODO(arslan) should we support nested lists? Even though it's
			// written in README of HCL, it's not a part of the grammar
			// (not defined in parse.y)
		case token.RBRACK:
			// finished
			l.Rbrack = p.tok.Pos
			return l, nil
		default:
			return nil, fmt.Errorf("unexpected token while parsing list: %s", tok.Type)
		}

	}
}

// literalType parses a literal type and returns a LiteralType AST
func (p *Parser) literalType() (*ast.LiteralType, error) {
	defer un(trace(p, "ParseLiteral"))

	return &ast.LiteralType{
		Token: p.tok,
	}, nil
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead.
func (p *Parser) scan() token.Token {
	// If we have a token on the buffer, then return it.
	if p.n != 0 {
		p.n = 0
		return p.tok
	}

	// Otherwise read the next token from the scanner and Save it to the buffer
	// in case we unscan later.
	p.tok = p.sc.Scan()
	return p.tok
}

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unscan() {
	p.n = 1
}

// ----------------------------------------------------------------------------
// Parsing support

func (p *Parser) printTrace(a ...interface{}) {
	if !p.enableTrace {
		return
	}

	const dots = ". . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . "
	const n = len(dots)
	fmt.Printf("%5d:%3d: ", p.tok.Pos.Line, p.tok.Pos.Column)

	i := 2 * p.indent
	for i > n {
		fmt.Print(dots)
		i -= n
	}
	// i <= n
	fmt.Print(dots[0:i])
	fmt.Println(a...)
}

func trace(p *Parser, msg string) *Parser {
	p.printTrace(msg, "(")
	p.indent++
	return p
}

// Usage pattern: defer un(trace(p, "..."))
func un(p *Parser) {
	p.indent--
	p.printTrace(")")
}
