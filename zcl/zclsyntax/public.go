package zclsyntax

import (
	"github.com/zclconf/go-zcl/zcl"
)

// ParseConfig parses the given buffer as a whole zcl config file, returning
// a Body representing its contents. If HasErrors called on the returned
// diagnostics returns true, the returned body is likely to be incomplete
// and should therefore be used with care.
func ParseConfig(src []byte, filename string, start zcl.Pos) (*Body, zcl.Diagnostics) {
	tokens := LexConfig(src, filename, start)
	peeker := newPeeker(tokens, false)
	parser := &parser{peeker: peeker}
	return parser.ParseBody(TokenEOF)
}

// ParseExpression parses the given buffer as a standalone zcl expression,
// returning it as an instance of Expression.
func ParseExpression(src []byte, filename string, start zcl.Pos) (*Expression, zcl.Diagnostics) {
	panic("ParseExpression is not yet implemented")
}

// ParseTemplate parses the given buffer as a standalone zcl template,
// returning it as an instance of Expression.
func ParseTemplate(src []byte, filename string, start zcl.Pos) (*Expression, zcl.Diagnostics) {
	panic("ParseTemplate is not yet implemented")
}

// LexConfig performs lexical analysis on the given buffer, treating it as a
// whole zcl config file, and returns the resulting tokens.
func LexConfig(src []byte, filename string, start zcl.Pos) Tokens {
	return scanTokens(src, filename, start, scanNormal)
}

// LexExpression performs lexical analysis on the given buffer, treating it as
// a standalone zcl expression, and returns the resulting tokens.
func LexExpression(src []byte, filename string, start zcl.Pos) Tokens {
	// This is actually just the same thing as LexConfig, since configs
	// and expressions lex in the same way.
	return scanTokens(src, filename, start, scanNormal)
}

// LexTemplate performs lexical analysis on the given buffer, treating it as a
// standalone zcl template, and returns the resulting tokens.
func LexTemplate(src []byte, filename string, start zcl.Pos) Tokens {
	return scanTokens(src, filename, start, scanTemplate)
}
