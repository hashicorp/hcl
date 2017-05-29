package zclsyntax

import (
	"github.com/zclconf/go-zcl/zcl"
)

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
