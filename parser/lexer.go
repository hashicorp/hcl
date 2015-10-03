package parser

import (
	"bufio"
	"bytes"
	"io"
	"unicode"
)

// eof represents a marker rune for the end of the reader.
const eof = rune(0)

// Lexer defines a lexical scanner
type Lexer struct {
	src *bufio.Reader // input
	ch  rune          // current character
}

// NewLexer returns a new instance of Lexer.
func NewLexer(src io.Reader) *Lexer {
	return &Lexer{
		src: bufio.NewReader(src),
	}
}

// next reads the next rune from the bufferred reader. Returns the rune(0) if
// an error occurs (or io.EOF is returned).
func (l *Lexer) next() rune {
	var err error
	l.ch, _, err = l.src.ReadRune()
	if err != nil {
		return eof
	}

	return l.ch
}

// unread places the previously read rune back on the reader.
func (l *Lexer) unread() { _ = l.src.UnreadRune() }

func (l *Lexer) peek() rune {
	prev := l.ch
	peekCh := l.next()
	l.unread()
	l.ch = prev
	return peekCh
}

// Scan scans the next token and returns the token and it's literal string.
func (l *Lexer) Scan() (tok Token, lit string) {
	ch := l.next()

	// skip white space
	for isWhitespace(ch) {
		ch = l.next()
	}

	// identifier
	if isLetter(ch) {
		return l.scanIdentifier()
	}

	switch ch {
	case eof:
		return EOF, ""
	}

	return 0, ""
}

func (l *Lexer) scanIdentifier() (Token, string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer

	for isLetter(l.ch) || isDigit(l.ch) {
		buf.WriteRune(l.ch)
		l.next()
	}

	return IDENT, buf.String()
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
