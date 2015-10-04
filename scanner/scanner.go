package scanner

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"unicode"

	"github.com/fatih/hcl/token"
)

// eof represents a marker rune for the end of the reader.
const eof = rune(0)

// Scanner defines a lexical scanner
type Scanner struct {
	src      *bytes.Buffer
	srcBytes []byte

	lastCharLen int // length of last character in bytes

	currPos Position // current position
	prevPos Position // previous position

	tokBuf bytes.Buffer // token text buffer
	tokPos int          // token text tail position (srcBuf index); valid if >= 0
	tokEnd int          // token text tail end (srcBuf index)

	// Error is called for each error encountered. If no Error
	// function is set, the error is reported to os.Stderr.
	Error func(pos Position, msg string)

	// ErrorCount is incremented by one for each error encountered.
	ErrorCount int
}

// NewScanner returns a new instance of Lexer. Even though src is an io.Reader,
// we fully consume the content.
func NewScanner(src io.Reader) (*Scanner, error) {
	buf, err := ioutil.ReadAll(src)
	if err != nil {
		return nil, err
	}

	b := bytes.NewBuffer(buf)
	return &Scanner{
		src:      b,
		srcBytes: b.Bytes(),
	}, nil
}

// next reads the next rune from the bufferred reader. Returns the rune(0) if
// an error occurs (or io.EOF is returned).
func (s *Scanner) next() rune {
	ch, size, err := s.src.ReadRune()
	if err != nil {
		return eof
	}

	// remember last position
	s.prevPos = s.currPos

	s.lastCharLen = size
	s.currPos.Offset += size
	s.currPos.Column += size

	if ch == '\n' {
		s.currPos.Line++
		s.currPos.Column = 0
	}

	return ch
}

func (s *Scanner) unread() {
	if err := s.src.UnreadRune(); err != nil {
		panic(err) // this is user fault, we should catch it
	}
	s.currPos = s.prevPos // put back last position
}

func (s *Scanner) peek() rune {
	peek, _, err := s.src.ReadRune()
	if err != nil {
		return eof
	}

	s.src.UnreadRune()
	return peek
}

// Scan scans the next token and returns the token.
func (s *Scanner) Scan() (tok token.Token) {
	ch := s.next()

	// skip white space
	for isWhitespace(ch) {
		ch = s.next()
	}

	// start the token position
	s.tokBuf.Reset()
	s.tokPos = s.currPos.Offset - s.lastCharLen

	if isLetter(ch) {
		tok = token.IDENT
		lit := s.scanIdentifier()
		if lit == "true" || lit == "false" {
			tok = token.BOOL
		}
	}

	if isDigit(ch) {
		// scanDigits()
		// TODO(arslan)
	}

	switch ch {
	case eof:
		tok = token.EOF
	case '"':
		tok = token.STRING
		s.scanString()
	}

	s.tokEnd = s.currPos.Offset
	return tok
}

func (s *Scanner) scanString() {
	for {
		// '"' opening already consumed
		// read character after quote
		ch := s.next()

		if ch == '\n' || ch < 0 || ch == eof {
			s.err("literal not terminated")
			return
		}

		if ch == '"' {
			break
		}

		if ch == '\\' {
			s.scanEscape()
		}
	}

	return
}

// scanEscape scans an escape sequence
func (s *Scanner) scanEscape() rune {
	// http://en.cppreference.com/w/cpp/language/escape
	ch := s.next() // read character after '/'
	switch ch {
	case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', '"':
		// nothing to do
	case '0', '1', '2', '3', '4', '5', '6', '7':
		// octal notation
		ch = s.scanDigits(ch, 8, 3)
	case 'x':
		// hexademical notation
		ch = s.scanDigits(s.next(), 16, 2)
	case 'u':
		// universal character name
		ch = s.scanDigits(s.next(), 16, 4)
	case 'U':
		// universal character name
		ch = s.scanDigits(s.next(), 16, 8)
	default:
		s.err("illegal char escape")
	}
	return ch
}

// scanDigits scans a rune with the given base for n times. For example an
// octan notation \184 would yield in scanDigits(ch, 8, 3)
func (s *Scanner) scanDigits(ch rune, base, n int) rune {
	for n > 0 && digitVal(ch) < base {
		ch = s.next()
		n--
	}
	if n > 0 {
		s.err("illegal char escape")
	}

	// we scanned all digits, put the last non digit char back
	s.unread()
	return ch
}

// scanIdentifier scans an identifier and returns the literal string
func (s *Scanner) scanIdentifier() string {
	offs := s.currPos.Offset - s.lastCharLen
	ch := s.next()
	for isLetter(ch) || isDigit(ch) {
		ch = s.next()
	}
	s.unread() // we got identifier, put back latest char

	// return string(s.srcBytes[offs:(s.currPos.Offset - s.lastCharLen)])
	return string(s.srcBytes[offs:s.currPos.Offset])
}

// TokenText returns the literal string corresponding to the most recently
// scanned token.
func (s *Scanner) TokenText() string {
	if s.tokPos < 0 {
		// no token text
		return ""
	}

	// part of the token text was saved in tokBuf: save the rest in
	// tokBuf as well and return its content
	s.tokBuf.Write(s.srcBytes[s.tokPos:s.tokEnd])
	s.tokPos = s.tokEnd // ensure idempotency of TokenText() call
	return s.tokBuf.String()
}

// Pos returns the position of the character immediately after the character or
// token returned by the last call to Scan.
func (s *Scanner) Pos() Position {
	return s.currPos
}

func (s *Scanner) err(msg string) {
	s.ErrorCount++
	if s.Error != nil {
		s.Error(s.currPos, msg)
		return
	}

	fmt.Fprintf(os.Stderr, "%s: %s\n", s.currPos, msg)
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

func digitVal(ch rune) int {
	switch {
	case '0' <= ch && ch <= '9':
		return int(ch - '0')
	case 'a' <= ch && ch <= 'f':
		return int(ch - 'a' + 10)
	case 'A' <= ch && ch <= 'F':
		return int(ch - 'A' + 10)
	}
	return 16 // larger than any legal digit val
}
