package zclsyntax

import (
	"github.com/zclconf/go-zcl/zcl"
)

// ParseConfig parses the given buffer as a whole zcl config file, returning
// a *zcl.File representing its contents. If HasErrors called on the returned
// diagnostics returns true, the returned body is likely to be incomplete
// and should therefore be used with care.
//
// The body in the returned file has dynamic type *zclsyntax.Body, so callers
// may freely type-assert this to get access to the full zclsyntax API in
// situations where detailed access is required. However, most common use-cases
// should be served using the zcl.Body interface to ensure compatibility with
// other configurationg syntaxes, such as JSON.
func ParseConfig(src []byte, filename string, start zcl.Pos) (*zcl.File, zcl.Diagnostics) {
	tokens := LexConfig(src, filename, start)
	peeker := newPeeker(tokens, false)
	parser := &parser{peeker: peeker}
	body, diags := parser.ParseBody(TokenEOF)
	return &zcl.File{
		Body:  body,
		Bytes: src,
	}, diags
}

// ParseExpression parses the given buffer as a standalone zcl expression,
// returning it as an instance of Expression.
func ParseExpression(src []byte, filename string, start zcl.Pos) (Expression, zcl.Diagnostics) {
	tokens := LexExpression(src, filename, start)
	peeker := newPeeker(tokens, false)
	parser := &parser{peeker: peeker}

	// Bare expressions are always parsed in  "ignore newlines" mode, as if
	// they were wrapped in parentheses.
	parser.PushIncludeNewlines(false)

	expr, diags := parser.ParseExpression()

	next := parser.Peek()
	if next.Type != TokenEOF && !parser.recovery {
		diags = append(diags, &zcl.Diagnostic{
			Severity: zcl.DiagError,
			Summary:  "Extra characters after expression",
			Detail:   "An expression was successfully parsed, but extra characters were found after it.",
			Subject:  &next.Range,
		})
	}

	return expr, diags
}

// ParseTemplate parses the given buffer as a standalone zcl template,
// returning it as an instance of Expression.
func ParseTemplate(src []byte, filename string, start zcl.Pos) (Expression, zcl.Diagnostics) {
	tokens := LexTemplate(src, filename, start)
	peeker := newPeeker(tokens, false)
	parser := &parser{peeker: peeker}
	return parser.ParseTemplate(TokenEOF)
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
