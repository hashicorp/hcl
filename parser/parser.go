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

	tok     token.Token // last read token
	prevTok token.Token // previous read token

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
func (p *Parser) Parse() (ast.Node, error) {
	defer un(trace(p, "ParseObjectList"))
	node := &ast.ObjectList{}

	for {
		n, err := p.parseObjectItem()
		if err == errEofToken {
			break // we are finished
		}
		if err != nil {
			return nil, err
		}

		// we successfully parsed a node, add it to the final source node
		node.Add(n)
	}

	return node, nil
}

// parseObjectItem parses a single object item
func (p *Parser) parseObjectItem() (*ast.ObjectItem, error) {
	defer un(trace(p, "ParseObjectItem"))

	keys, err := p.parseObjectKey()
	if err != nil {
		return nil, err
	}

	// either an assignment or object
	switch p.tok.Type {
	case token.ASSIGN:
		o := &ast.ObjectItem{
			Keys:   keys,
			Assign: p.tok.Pos,
		}

		o.Val, err = p.parseType()
		if err != nil {
			return nil, err
		}

		return o, nil
	case token.LBRACE:
		if len(keys) > 1 {
			// nested object
			fmt.Println("nested object")
		}

		// object
		fmt.Println("object")
	}

	return nil, fmt.Errorf("not yet implemented: %s", p.tok.Type)
}

// parseType parses any type of Type, such as number, bool, string, object or
// list.
func (p *Parser) parseType() (ast.Node, error) {
	defer un(trace(p, "ParseType"))
	tok := p.scan()

	switch tok.Type {
	case token.NUMBER, token.FLOAT, token.BOOL, token.STRING:
		return p.parseLiteralType()
	case token.LBRACE:
		return p.parseObjectType()
	case token.LBRACK:
		return p.parseListType()
	case token.COMMENT:
		// implement comment
	case token.EOF:
		return nil, errEofToken
	}

	return nil, errors.New("ParseType is not implemented yet")
}

// parseObjectKey parses an object key and returns a ObjectKey AST
func (p *Parser) parseObjectKey() ([]*ast.ObjectKey, error) {
	tok := p.scan()
	if tok.Type == token.EOF {
		return nil, errEofToken
	}

	keys := make([]*ast.ObjectKey, 0)

	switch tok.Type {
	case token.IDENT, token.STRING:
		// add first found token
		keys = append(keys, &ast.ObjectKey{Token: tok})
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
		case token.ASSIGN:
			// assignment or object only, but not nested objects. this is not
			// allowed: `foo bar = {}`
			if nestedObj {
				return nil, fmt.Errorf("nested object expected: LBRACE got: %s", tok.Type)
			}

			return keys, nil
		case token.LBRACE:
			// object
			return keys, nil
		case token.IDENT, token.STRING:
			// nested object
			nestedObj = true
			keys = append(keys, &ast.ObjectKey{Token: tok})
		default:
			return nil, fmt.Errorf("expected: IDENT | STRING | ASSIGN | LBRACE got: %s", tok.Type)
		}
	}
}

// parseListType parses a list type and returns a ListType AST
func (p *Parser) parseListType() (*ast.ListType, error) {
	defer un(trace(p, "ParseListType"))

	l := &ast.ListType{
		Lbrack: p.tok.Pos,
	}

	for {
		tok := p.scan()
		switch tok.Type {
		case token.NUMBER, token.FLOAT, token.STRING:
			node, err := p.parseLiteralType()
			if err != nil {
				return nil, err
			}
			l.Add(node)
		case token.COMMA:
			// get next list item or we are at the end
			continue
		case token.BOOL:
			// TODO(arslan) should we support? not supported by HCL yet
		case token.LBRACK:
			// TODO(arslan) should we support nested lists?
		case token.RBRACK:
			// finished
			l.Rbrack = p.tok.Pos
			return l, nil
		default:
			return nil, fmt.Errorf("unexpected token while parsing list: %s", tok.Type)
		}

	}
}

// parseLiteralType parses a literal type and returns a LiteralType AST
func (p *Parser) parseLiteralType() (*ast.LiteralType, error) {
	defer un(trace(p, "ParseLiteral"))

	return &ast.LiteralType{
		Token: p.tok,
	}, nil
}

// parseObjectType parses an object type and returns a ObjectType AST
func (p *Parser) parseObjectType() (*ast.ObjectType, error) {
	defer un(trace(p, "ParseObjectYpe"))

	return nil, errors.New("ObjectType is not implemented yet")
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead.
func (p *Parser) scan() token.Token {
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
