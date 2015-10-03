package parser

// Token is the set of lexical tokens of the HCL (HashiCorp Configuration Language)
type Token int

const (
	// Special tokens
	ILLEGAL Token = iota
	EOF
	COMMENT
	NEWLINE

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

// IsLiteral returns true for tokens corresponding to identifiers and basic
// type literals; it returns false otherwise.
func (t Token) IsLiteral() bool { return literal_beg < t && t < literal_end }

// IsOperator returns true for tokens corresponding to operators and
// delimiters; it returns false otherwise.
func (t Token) IsOperator() bool { return operator_beg < t && t < operator_end }
