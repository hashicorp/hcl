package parser

import "github.com/fatih/hcl/scanner"

type Parser struct {
	sc  *scanner.Scanner
	buf struct {
		tok scanner.Token // last read token
		n   int           // buffer size (max = 1)
	}
}

func NewParser(src []byte) *Parser {
	return &Parser{
		sc: scanner.New(src),
	}
}

func (p *Parser) Parse() Node {
	tok := p.scan()

	switch tok.Type() {
	case scanner.IDENT:
		// p.parseStatement()
	case scanner.EOF:
	}

	return nil
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead.
func (p *Parser) scan() scanner.Token {
	// If we have a token on the buffer, then return it.
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok
	}

	// Otherwise read the next token from the scanner and Save it to the buffer
	// in case we unscan later.
	p.buf.tok = p.sc.Scan()

	return p.buf.tok
}

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unread() { p.buf.n = 1 }
