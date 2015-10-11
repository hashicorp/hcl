package parser

import (
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

// Parse returns the fully parsed source and returns the abstract syntax tree.
func (p *Parser) Parse() Node {
	defer un(trace(p, "ParseSource"))
	node := &Source{}

	for {
		// break if we hit the end
		if p.tok.Type == scanner.EOF {
			break
		}

		if n := p.parseStatement(); n != nil {
			node.add(n)
		}
	}

	return node
}

func (p *Parser) parseStatement() Node {
	defer un(trace(p, "ParseStatement"))

	tok := p.scan()

	if tok.Type.IsLiteral() {
		// found an object
		if p.prevTok.Type.IsLiteral() {
			return p.parseObject()
		}
		return p.parseIdent()
	}

	switch tok.Type {
	case scanner.LBRACE:
		return p.parseObject()
	case scanner.LBRACK:
		return p.parseList()
	case scanner.ASSIGN:
		return p.parseAssignment()
	}

	return nil
}

func (p *Parser) parseAssignment() Node {
	defer un(trace(p, "ParseAssignment"))
	return &AssignStatement{
		lhs: &Ident{
			token: p.prevTok,
		},
		assign: p.tok.Pos,
		rhs:    p.parseStatement(),
	}
}

func (p *Parser) parseIdent() Node {
	defer un(trace(p, "ParseIdent"))

	return &Ident{
		token: p.tok,
	}
}

func (p *Parser) parseObject() Node {
	return nil
}

func (p *Parser) parseList() Node {
	return nil
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
func (p *Parser) unscan() { p.n = 1 }

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
