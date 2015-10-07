package parser

import "github.com/fatih/hcl/scanner"

type Parser struct {
	sc *scanner.Scanner
}

func NewParser(src []byte) *Parser {
	return &Parser{
		sc: scanner.NewScanner(src),
	}
}

func (p *Parser) Parse() {
}
