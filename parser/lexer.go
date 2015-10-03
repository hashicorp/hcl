package parser

import (
	"bufio"
	"io"
	"unicode"
)

// eof represents a marker rune for the end of the reader.
const eof = rune(0)

// Lexer defines a lexical scanner
type Lexer struct {
	r *bufio.Reader

	// Start position of most recently scanned token; set by Scan.
	// Calling Init or Next invalidates the position (Line == 0).
	// The Filename field is always left untouched by the Scanner.
	// If an error is reported (via Error) and Position is invalid,
	// the scanner is not inside a token. Call Pos to obtain an error
	// position in that case.
	Position
}

// NewLexer returns a new instance of Lexer.
func NewLexer(r io.Reader) *Lexer {
	return &Lexer{
		r: bufio.NewReader(r),
	}
}

// next reads the next rune from the bufferred reader.  Returns the rune(0) if
// an error occurs (or io.EOF is returned).
func (l *Lexer) next() rune {
	ch, _, err := l.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

// unread places the previously read rune back on the reader.
func (l *Lexer) unread() { _ = l.r.UnreadRune() }

// Scan scans the next token and returns the token and it's literal string.
func (l *Lexer) Scan() (tok Token, lit string) {
	ch := l.next()

	if isWhitespace(ch) {
		ch = l.next()
	}

	return 0, ""
}

func (l *Lexer) skipWhitespace() {
	l.next()
}

// Pos returns the position of the character immediately after the character or
// token returned by the last call to Next or Scan.
func (l *Lexer) Pos() Position {
	return Position{}
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// isEndOfLine reports whether r is an end-of-line character.
func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch >= 0x80 && unicode.IsLetter(ch)
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9' || ch >= 0x80 && unicode.IsDigit(ch)
}

// isWhitespace returns true if the rune is a space, tab, newline or carriage return
func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}
