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

// parseObjectItem parses a single object item
func (p *Parser) parseObjectItem() (*ObjectItem, error) {
	defer un(trace(p, "ParseObjectItem"))

	keys, err := p.parseObjectKey()
	if err != nil {
		return nil, err
	}

	// either an assignment or object
	switch p.tok.Type {
	case scanner.ASSIGN:
	case scanner.LBRACE:
		if len(keys) > 1 {
			// nested object
		}

		// object or nested object
	}

	switch len(keys) {
	case 1:
		// assignment or object
	default:
		// nested object
	}

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

// parseObjectKey parses an object key and returns a ObjectKey AST
func (p *Parser) parseObjectKey() ([]*ObjectKey, error) {
	tok := p.scan()

	keys := make([]*ObjectKey, 0)

	switch tok.Type {
	case scanner.IDENT, scanner.STRING:
		// add first found token
		keys = append(keys, &ObjectKey{token: tok})
	default:
		return nil, fmt.Errorf("expected: IDENT | STRING got: %s", tok.Type)
	}

	nestedObj := false

	// we have three casses
	// 1. assignment: KEY = NODE
	// 2. object: KEY { }
	// 2. nested object: KEY KEY2 ... KEYN {}
	for {
		tok := p.scan()
		switch tok.Type {
		case scanner.ASSIGN:
			// assignment or object, but not nested objects
			if nestedObj {
				return nil, fmt.Errorf("nested object expected: LBRACE got: %s", tok.Type)
			}

			return keys, nil
		case scanner.LBRACE:
			// object
			return keys, nil
		case scanner.IDENT, scanner.STRING:
			// nested object
			nestedObj = true
			keys = append(keys, &ObjectKey{token: tok})
		default:
			return nil, fmt.Errorf("expected: IDENT | STRING | ASSIGN | LBRACE got: %s", tok.Type)
		}
	}
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
