package scanner

import (
	"fmt"
	"strconv"
)

// Token defines a single HCL token which can be obtained via the Scanner
type Token struct {
	Type TokenType
	Pos  Pos
	Text string
}

// TokenType is the set of lexical tokens of the HCL (HashiCorp Configuration Language)
type TokenType int

const (
	// Special tokens
	ILLEGAL TokenType = iota
	EOF
	COMMENT

	literal_beg
	IDENT  // literals
	NUMBER // 12345
	FLOAT  // 123.45
	BOOL   // true,false
	STRING // "abc"
	literal_end

	operator_beg
	LBRACK // [
	LBRACE // {
	COMMA  // ,
	PERIOD // .

	RBRACK // ]
	RBRACE // }

	ASSIGN // =
	ADD    // +
	SUB    // -
	operator_end
)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",

	EOF:     "EOF",
	COMMENT: "COMMENT",

	IDENT:  "IDENT",
	NUMBER: "NUMBER",
	FLOAT:  "FLOAT",
	BOOL:   "BOOL",
	STRING: "STRING",

	LBRACK: "LBRACK",
	LBRACE: "LBRACE",
	COMMA:  "COMMA",
	PERIOD: "PERIOD",

	RBRACK: "RBRACK",
	RBRACE: "RBRACE",

	ASSIGN: "ASSIGN",
	ADD:    "ADD",
	SUB:    "SUB",
}

// String returns the string corresponding to the token tok.
func (t TokenType) String() string {
	s := ""
	if 0 <= t && t < TokenType(len(tokens)) {
		s = tokens[t]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(t)) + ")"
	}
	return s
}

// IsLiteral returns true for tokens corresponding to identifiers and basic
// type literals; it returns false otherwise.
func (t TokenType) IsLiteral() bool { return literal_beg < t && t < literal_end }

// IsOperator returns true for tokens corresponding to operators and
// delimiters; it returns false otherwise.
func (t TokenType) IsOperator() bool { return operator_beg < t && t < operator_end }

// String returns the token's literal text. Note that this is only
// applicable for certain token types, such as token.IDENT,
// token.STRING, etc..
func (t Token) String() string {
	return fmt.Sprintf("%s %s %s", t.Pos.String(), t.Type.String(), t.Text)
}
