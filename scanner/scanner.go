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
	src *bytes.Buffer

	// Source Buffer
	srcBuf []byte

	// Source Position
	srcPos  Position // current position
	prevPos Position // previous position

	lastCharLen int // length of last character in bytes
	lastLineLen int // length of last line in characters (for correct column reporting)

	tokBuf   bytes.Buffer // token text buffer
	tokStart int          // token text start position
	tokEnd   int          // token text end  position

	// Error is called for each error encountered. If no Error
	// function is set, the error is reported to os.Stderr.
	Error func(pos Position, msg string)

	// ErrorCount is incremented by one for each error encountered.
	ErrorCount int

	// Start position of most recently scanned token; set by Scan.
	// Calling Init or Next invalidates the position (Line == 0).
	// The Filename field is always left untouched by the Scanner.
	// If an error is reported (via Error) and Position is invalid,
	// the scanner is not inside a token. Call Pos to obtain an error
	// position in that case.
	tokPos Position
}

// NewScanner returns a new instance of Lexer. Even though src is an io.Reader,
// we fully consume the content.
func NewScanner(src io.Reader) (*Scanner, error) {
	buf, err := ioutil.ReadAll(src)
	if err != nil {
		return nil, err
	}

	b := bytes.NewBuffer(buf)
	s := &Scanner{
		src:    b,
		srcBuf: b.Bytes(),
	}

	// srcPosition always starts with 1
	s.srcPos.Line = 1

	return s, nil
}

// next reads the next rune from the bufferred reader. Returns the rune(0) if
// an error occurs (or io.EOF is returned).
func (s *Scanner) next() rune {
	ch, size, err := s.src.ReadRune()
	if err != nil {
		return eof
	}

	// remember last position
	s.prevPos = s.srcPos
	s.lastCharLen = size

	s.srcPos.Offset += size
	s.srcPos.Column++
	if ch == '\n' {
		s.srcPos.Line++
		s.srcPos.Column = 0
		s.lastLineLen = s.srcPos.Column
	}

	// debug
	// fmt.Printf("ch: %q, off:column: %d:%d\n", ch, s.srcPos.Offset, s.srcPos.Column)
	return ch
}

func (s *Scanner) unread() {
	if err := s.src.UnreadRune(); err != nil {
		panic(err) // this is user fault, we should catch it
	}
	s.srcPos = s.prevPos // put back last position
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

	// token text markings
	s.tokBuf.Reset()
	s.tokStart = s.srcPos.Offset - s.lastCharLen

	// token position, initial next() is moving the offset by one(size of rune
	// actually), though we are interested with the starting point
	s.tokPos.Offset = s.srcPos.Offset - s.lastCharLen

	if s.srcPos.Column > 0 {
		// common case: last character was not a '\n'
		s.tokPos.Line = s.srcPos.Line
		s.tokPos.Column = s.srcPos.Column
	} else {
		// last character was a '\n'
		// (we cannot be at the beginning of the source
		// since we have called next() at least once)
		s.tokPos.Line = s.srcPos.Line - 1
		s.tokPos.Column = s.lastLineLen
	}

	switch {
	case isLetter(ch):
		tok = token.IDENT
		lit := s.scanIdentifier()
		if lit == "true" || lit == "false" {
			tok = token.BOOL
		}
	case isDecimal(ch):
		tok = s.scanNumber(ch)
	default:
		switch ch {
		case eof:
			tok = token.EOF
		case '"':
			tok = token.STRING
			s.scanString()
		case '#', '/':
			tok = token.COMMENT
			s.scanComment(ch)
		case '.':
			tok = token.PERIOD
			ch = s.peek()
			if isDecimal(ch) {
				tok = token.FLOAT
				ch = s.scanMantissa(ch)
				ch = s.scanExponent(ch)
			}
		case '[':
			tok = token.LBRACK
		case ']':
			tok = token.RBRACK
		case '{':
			tok = token.LBRACE
		case '}':
			tok = token.RBRACE
		case ',':
			tok = token.COMMA
		case '=':
			tok = token.ASSIGN
		case '+':
			tok = token.ADD
		case '-':
			tok = token.SUB
		}
	}

	s.tokEnd = s.srcPos.Offset
	return tok
}

func (s *Scanner) scanComment(ch rune) {
	// look for /* - style comments
	if ch == '/' && s.peek() == '*' {
		for {
			if ch < 0 {
				s.err("comment not terminated")
				break
			}

			ch0 := ch
			ch = s.next()
			if ch0 == '*' && ch == '/' {
				break
			}
		}
	}

	// single line comments
	if ch == '#' || ch == '/' {
		ch = s.next()
		for ch != '\n' && ch >= 0 {
			ch = s.next()
		}
		s.unread()
		return
	}
}

// scanNumber scans a HCL number definition starting with the given rune
func (s *Scanner) scanNumber(ch rune) token.Token {
	if ch == '0' {
		// check for hexadecimal, octal or float
		ch = s.next()
		if ch == 'x' || ch == 'X' {
			// hexadecimal
			ch = s.next()
			found := false
			for isHexadecimal(ch) {
				ch = s.next()
				found = true
			}
			s.unread()

			if !found {
				s.err("illegal hexadecimal number")
			}

			return token.NUMBER
		}

		// now it's either something like: 0421(octal) or 0.1231(float)
		illegalOctal := false
		for isDecimal(ch) {
			ch = s.next()
			if ch == '8' || ch == '9' {
				// this is just a possibility. For example 0159 is illegal, but
				// 0159.23 is valid. So we mark a possible illegal octal. If
				// the next character is not a period, we'll print the error.
				illegalOctal = true

			}

		}

		// literals of form 01e10 are treates as Numbers in HCL, which differs from Go.
		if ch == 'e' || ch == 'E' {
			ch = s.scanExponent(ch)
			return token.NUMBER
		}

		if ch == '.' {
			ch = s.scanFraction(ch)

			if ch == 'e' || ch == 'E' {
				ch = s.next()
				ch = s.scanExponent(ch)
			}
			return token.FLOAT
		}

		if illegalOctal {
			s.err("illegal octal number")
		}

		s.unread()
		return token.NUMBER
	}

	s.scanMantissa(ch)
	ch = s.next() // seek forward
	// literals of form 1e10 are treates as Numbers in HCL, which differs from Go.
	if ch == 'e' || ch == 'E' {
		ch = s.scanExponent(ch)
		return token.NUMBER
	}

	if ch == '.' {
		ch = s.scanFraction(ch)
		if ch == 'e' || ch == 'E' {
			ch = s.next()
			ch = s.scanExponent(ch)
		}
		return token.FLOAT
	}

	s.unread()
	return token.NUMBER
}

// scanMantissa scans the mantissa begining from the rune. It returns the next
// non decimal rune. It's used to determine wheter it's a fraction or exponent.
func (s *Scanner) scanMantissa(ch rune) rune {
	scanned := false
	for isDecimal(ch) {
		ch = s.next()
		scanned = true
	}

	if scanned {
		s.unread()
	}
	return ch
}

func (s *Scanner) scanFraction(ch rune) rune {
	if ch == '.' {
		ch = s.peek() // we peek just to see if we can move forward
		ch = s.scanMantissa(ch)
	}
	return ch
}

func (s *Scanner) scanExponent(ch rune) rune {
	if ch == 'e' || ch == 'E' {
		ch = s.next()
		if ch == '-' || ch == '+' {
			ch = s.next()
		}
		ch = s.scanMantissa(ch)
	}
	return ch
}

// scanString scans a quoted string
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
	offs := s.srcPos.Offset - s.lastCharLen
	ch := s.next()
	for isLetter(ch) || isDigit(ch) {
		ch = s.next()
	}
	s.unread() // we got identifier, put back latest char

	return string(s.srcBuf[offs:s.srcPos.Offset])
}

// TokenText returns the literal string corresponding to the most recently
// scanned token.
func (s *Scanner) TokenText() string {
	if s.tokStart < 0 {
		// no token text
		return ""
	}

	// part of the token text was saved in tokBuf: save the rest in
	// tokBuf as well and return its content
	s.tokBuf.Write(s.srcBuf[s.tokStart:s.tokEnd])
	s.tokStart = s.tokEnd // ensure idempotency of TokenText() call
	return s.tokBuf.String()
}

// Pos returns the position of the character immediately after the character or
// token returned by the last call to Scan.
func (s *Scanner) Pos() (pos Position) {
	return s.tokPos
}

func (s *Scanner) err(msg string) {
	s.ErrorCount++
	if s.Error != nil {
		s.Error(s.srcPos, msg)
		return
	}

	fmt.Fprintf(os.Stderr, "%s: %s\n", s.srcPos, msg)
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch >= 0x80 && unicode.IsLetter(ch)
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9' || ch >= 0x80 && unicode.IsDigit(ch)
}

func isOctal(ch rune) bool {
	return '0' <= ch && ch <= '7'
}

func isDecimal(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

func isHexadecimal(ch rune) bool {
	return '0' <= ch && ch <= '9' || 'a' <= ch && ch <= 'f' || 'A' <= ch && ch <= 'F'
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
