package token

import (
	"fmt"
	"strconv"

	hclstrconv "github.com/hashicorp/hcl/hcl/strconv"
	hcltoken "github.com/hashicorp/hcl/hcl/token"
)

// Token defines a single HCL token which can be obtained via the Scanner
type Token struct {
	Type Type
	Pos  Pos
	Text string
}

// Type is the set of lexical tokens of the HCL (HashiCorp Configuration Language)
type Type int

const (
	// Special tokens
	ILLEGAL Type = iota
	EOF

	identifier_beg
	literal_beg
	NUMBER // 12345
	FLOAT  // 123.45
	BOOL   // true,false
	STRING // "abc"
	NULL   // null
	literal_end
	identifier_end

	operator_beg
	LBRACK // [
	LBRACE // {
	COMMA  // ,
	PERIOD // .
	COLON  // :

	RBRACK // ]
	RBRACE // }

	operator_end
)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",

	EOF: "EOF",

	NUMBER: "NUMBER",
	FLOAT:  "FLOAT",
	BOOL:   "BOOL",
	STRING: "STRING",
	NULL:   "NULL",

	LBRACK: "LBRACK",
	LBRACE: "LBRACE",
	COMMA:  "COMMA",
	PERIOD: "PERIOD",
	COLON:  "COLON",

	RBRACK: "RBRACK",
	RBRACE: "RBRACE",
}

// String returns the string corresponding to the token tok.
func (t Type) String() string {
	s := ""
	if 0 <= t && t < Type(len(tokens)) {
		s = tokens[t]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(t)) + ")"
	}
	return s
}

// IsIdentifier returns true for tokens corresponding to identifiers and basic
// type literals; it returns false otherwise.
func (t Type) IsIdentifier() bool { return identifier_beg < t && t < identifier_end }

// IsLiteral returns true for tokens corresponding to basic type literals; it
// returns false otherwise.
func (t Type) IsLiteral() bool { return literal_beg < t && t < literal_end }

// IsOperator returns true for tokens corresponding to operators and
// delimiters; it returns false otherwise.
func (t Type) IsOperator() bool { return operator_beg < t && t < operator_end }

// String returns the token's literal text. Note that this is only
// applicable for certain token types, such as token.IDENT,
// token.STRING, etc..
func (t Token) String() string {
	return fmt.Sprintf("%s %s %s", t.Pos.String(), t.Type.String(), t.Text)
}

// Value returns the properly typed value for this token. The type of
// the returned interface{} is guaranteed based on the Type field.
//
// This can only be called for literal types. If it is called for any other
// type, this will panic.
func (t Token) Value() interface{} {
	switch t.Type {
	case BOOL:
		if t.Text == "true" {
			return true
		} else if t.Text == "false" {
			return false
		}

		panic("unknown bool value: " + t.Text)
	case FLOAT:
		v, err := strconv.ParseFloat(t.Text, 64)
		if err != nil {
			panic(err)
		}

		return float64(v)
	case NULL:
		return nil
	case NUMBER:
		v, err := strconv.ParseInt(t.Text, 0, 64)
		if err != nil {
			panic(err)
		}

		return int64(v)
	case STRING:
		v, err := hclstrconv.Unquote(t.Text)
		if err != nil {
			panic(fmt.Sprintf("unquote %s err: %s", t.Text, err))
		}

		return v
	default:
		panic(fmt.Sprintf("unimplemented Value for type: %s", t.Type))
	}
}

// HCLToken converts this token to an HCL token.
//
// The token type must be a literal type or this will panic.
func (t Token) HCLToken() hcltoken.Token {
	switch t.Type {
	case BOOL:
		return hcltoken.Token{Type: hcltoken.BOOL, Text: t.Text}
	case FLOAT:
		return hcltoken.Token{Type: hcltoken.FLOAT, Text: t.Text}
	case NULL:
		return hcltoken.Token{Type: hcltoken.STRING, Text: ""}
	case NUMBER:
		return hcltoken.Token{Type: hcltoken.NUMBER, Text: t.Text}
	case STRING:
		return hcltoken.Token{Type: hcltoken.STRING, Text: t.Text}
	default:
		panic(fmt.Sprintf("unimplemented HCLToken for type: %s", t.Type))
	}
}
