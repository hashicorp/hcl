package parser

import "github.com/fatih/hcl/scanner"

type Parser struct {
	sc *scanner.Scanner

	tok     scanner.Token // last read token
	prevTok scanner.Token // previous read token

	n int // buffer size (max = 1)
}

func NewParser(src []byte) *Parser {
	return &Parser{
		sc: scanner.New(src),
	}
}

// Parse returns the fully parsed source and returns the abstract syntax tree.
func (p *Parser) Parse() Node {
	tok := p.scan()

	node := Source{}

	switch tok.Type() {
	case scanner.IDENT:
		n := p.parseStatement()
		node.add(n)
	case scanner.EOF:
	}

	return node
}

func (p *Parser) parseStatement() Node {
	tok := p.scan()

	if tok.Type().IsLiteral() {
		return p.parseIdent()
	}

	switch tok.Type() {
	case scanner.LBRACE:
		return p.parseObject()
	case scanner.LBRACK:
		return p.parseList()
	case scanner.ASSIGN:
		return p.parseAssignment()
	}
	return nil
}

func (p *Parser) parseIdent() Node {
	return IdentStatement{
		token: p.tok,
	}
}

func (p *Parser) parseObject() Node {
	return nil
}

func (p *Parser) parseList() Node {
	return nil
}

func (p *Parser) parseAssignment() Node {
	return AssignStatement{
		lhs: IdentStatement{
			token: p.prevTok,
		},
		assign: p.tok.Pos(),
		rhs:    p.parseStatement(),
	}
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
