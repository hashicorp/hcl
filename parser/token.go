package parser

import "strconv"

// Token is the set of lexical tokens of the HCL (HashiCorp Configuration Language)
type Token int

const (
	// Special tokens
	ILLEGAL Token = iota
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

	EPLUS  // e
	EMINUS // e-
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

	LBRACK: "[",
	LBRACE: "{",
	COMMA:  ",",
	PERIOD: ".",

	RBRACK: "]",
	RBRACE: "}",

	ASSIGN: "=",
	ADD:    "+",
	SUB:    "-",

	EPLUS:  "e",
	EMINUS: "e-",
}

// String returns the string corresponding to the token tok.
// For operators, delimiters, and keywords the string is the actual
// token character sequence (e.g., for the token ADD, the string is
// "+"). For all other tokens the string corresponds to the token
// constant name (e.g. for the token IDENT, the string is "IDENT").
func (t Token) String() string {
	s := ""
	if 0 <= t && t < Token(len(tokens)) {
		s = tokens[t]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(t)) + ")"
	}
	return s
}

// IsLiteral returns true for tokens corresponding to identifiers and basic
// type literals; it returns false otherwise.
func (t Token) IsLiteral() bool { return literal_beg < t && t < literal_end }

// IsOperator returns true for tokens corresponding to operators and
// delimiters; it returns false otherwise.
func (t Token) IsOperator() bool { return operator_beg < t && t < operator_end }
