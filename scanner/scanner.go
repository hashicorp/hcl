package scanner

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
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
	// '"' opening already consumed
	ch := s.next() // read character after quote
	for ch != '"' {
		if ch == '\n' || ch < 0 {
			log.Println("[ERROR] literal not terminated")
			return
		}

		if ch == '\\' {
			// scanEscape
			return
		} else {
			ch = s.next()
		}
	}

	return
}

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
