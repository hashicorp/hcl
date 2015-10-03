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
type Scanner struct {
	src *bufio.Reader // input
	ch  rune          // current character
}

// NewLexer returns a new instance of Lexer.
func NewLexer(src io.Reader) *Scanner {
	return &Scanner{
		src: bufio.NewReader(src),
	}
}

// next reads the next rune from the bufferred reader. Returns the rune(0) if
// an error occurs (or io.EOF is returned).
func (s *Scanner) next() rune {
	var err error
	s.ch, _, err = s.src.ReadRune()
	if err != nil {
		return eof
	}

	return s.ch
}

// unread places the previously read rune back on the reader.
func (s *Scanner) unread() { _ = s.src.UnreadRune() }

func (s *Scanner) peek() rune {
	prev := s.ch
	peekCh := s.next()
	s.unread()
	s.ch = prev
	return peekCh
}

// Scan scans the next token and returns the token and it's literal string.
func (s *Scanner) Scan() (tok Token, lit string) {
	ch := s.next()

	// skip white space
	for isWhitespace(ch) {
		ch = s.next()
	}

	// identifier
	if isLetter(ch) {
		return s.scanIdentifier()
	}

	switch ch {
	case eof:
		return EOF, ""
	}

	return 0, ""
}

func (s *Scanner) scanIdentifier() (Token, string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer

	for isLetter(s.ch) || isDigit(s.ch) {
		buf.WriteRune(s.ch)
		s.next()
	}

	return IDENT, buf.String()
}

// Pos returns the position of the character immediately after the character or
// token returned by the last call to Next or Scan.
func (s *Scanner) Pos() Position {
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
