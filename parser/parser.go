package parser

import (
	"errors"
	"fmt"

	"github.com/fatih/hcl/scanner"
)

type Parser struct {
	sc *scanner.Scanner

	tok     scanner.Token // last read token
	prevTok scanner.Token // previous read token

	enableTrace bool
	indent      int
	n           int // buffer size (max = 1)
}

func New(src []byte) *Parser {
	return &Parser{
		sc: scanner.New(src),
	}
}

var errEofToken = errors.New("EOF token found")

// Parse returns the fully parsed source and returns the abstract syntax tree.
func (p *Parser) Parse() (Node, error) {
	defer un(trace(p, "ParseSource"))
	node := &ObjectList{}

	for {
		n, err := p.parseObjectItem()
		if err == errEofToken {
			break // we are finished
		}
		if err != nil {
			return nil, err
		}

		// we successfully parsed a node, add it to the final source node
		node.add(n)
	}

	return node, nil
}

func (p *Parser) parseObjectItem() (*ObjectItem, error) {
	defer un(trace(p, "ParseObjectItem"))

	tok := p.scan()
	fmt.Println(tok) // debug

	switch tok.Type {
	case scanner.ASSIGN:
		// return p.parseAssignment()
	case scanner.LBRACK:
		// return p.parseListType()
	case scanner.LBRACE:
		// return p.parseObjectTpe()
	case scanner.COMMENT:
		// implement comment
	case scanner.EOF:
		return nil, errEofToken
	}

	return nil, fmt.Errorf("not yet implemented: %s", tok.Type)
}

// parseIdent parses a generic identifier and returns a Ident AST
func (p *Parser) parseIdent() (*Ident, error) {
	defer un(trace(p, "ParseIdent"))

	return &Ident{
		token: p.tok,
	}, nil
}

// parseLiteralType parses a literal type and returns a LiteralType AST
func (p *Parser) parseLiteralType() (*LiteralType, error) {
	defer un(trace(p, "ParseLiteral"))

	return &LiteralType{
		token: p.tok,
	}, nil
}

// parseObjectType parses an object type and returns a ObjectType AST
func (p *Parser) parseObjectType() (*ObjectType, error) {
	return nil, errors.New("ObjectType is not implemented yet")
}

// parseListType parses a list type and returns a ListType AST
func (p *Parser) parseListType() (*ListType, error) {
	return nil, errors.New("ListType is not implemented yet")
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead.
func (p *Parser) scan() scanner.Token {
	// If we have a token on the buffer, then return it.
	if p.n != 0 {
		p.n = 0
		return p.tok
	}

	// store previous token
	p.prevTok = p.tok

	// Otherwise read the next token from the scanner and Save it to the buffer
	// in case we unscan later.
	p.tok = p.sc.Scan()
	return p.tok
}

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unscan() {
	p.n = 1
	p.tok = p.prevTok
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
