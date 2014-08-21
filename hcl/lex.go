package hcl

import (
	"bytes"
	"fmt"
	"strconv"
	"unicode"
	"unicode/utf8"
)

// The parser expects the lexer to return 0 on EOF.
const lexEOF = 0

// The parser uses the type <prefix>Lex as a lexer.  It must provide
// the methods Lex(*<prefix>SymType) int and Error(string).
type hclLex struct {
	Input string

	lastNumber bool
	pos        int
	width      int
	col, line  int
	err        error
}

// The parser calls this method to get each new token.
func (x *hclLex) Lex(yylval *hclSymType) int {
	for {
		c := x.next()
		if c == lexEOF {
			return lexEOF
		}

		// Ignore all whitespace except a newline which we handle
		// specially later.
		if unicode.IsSpace(c) {
			continue
		}

		// Consume all comments
		switch c {
		case '#':
			fallthrough
		case '/':
			// Starting comment
			if !x.consumeComment(c) {
				return lexEOF
			}
			continue
		}

		// If it is a number, lex the number
		if c >= '0' && c <= '9' {
			x.lastNumber = true
			x.backup()
			return x.lexNumber(yylval)
		}

		// This is a hacky way to find 'e' and lex it, but it works.
		if x.lastNumber {
			switch c {
			case 'e':
				fallthrough
			case 'E':
				switch x.next() {
				case '+':
					return EPLUS
				case '-':
					return EMINUS
				default:
					x.backup()
					return EPLUS
				}
			}
		}
		x.lastNumber = false

		switch c {
		case '.':
			return PERIOD
		case '-':
			return MINUS
		case ',':
			return COMMA
		case '=':
			return EQUAL
		case '[':
			return LEFTBRACKET
		case ']':
			return RIGHTBRACKET
		case '{':
			return LEFTBRACE
		case '}':
			return RIGHTBRACE
		case '"':
			return x.lexString(yylval)
		default:
			x.backup()
			return x.lexId(yylval)
		}
	}
}

func (x *hclLex) consumeComment(c rune) bool {
	single := c == '#'
	if !single {
		c = x.next()
		if c != '/' && c != '*' {
			x.backup()
			x.createErr(fmt.Sprintf("comment expected, got '%c'", c))
			return false
		}

		single = c == '/'
	}

	nested := 1
	for {
		c = x.next()
		if c == lexEOF {
			x.backup()
			return true
		}

		// Single line comments continue until a '\n'
		if single {
			if c == '\n' {
				return true
			}

			continue
		}

		// Multi-line comments continue until a '*/'
		switch c {
		case '/':
			c = x.next()
			if c == '*' {
				nested++
			} else {
				x.backup()
			}
		case '*':
			c = x.next()
			if c == '/' {
				nested--
			} else {
				x.backup()
			}
		default:
			// Continue
		}

		// If we're done with the comment, return!
		if nested == 0 {
			return true
		}
	}
}

// lexId lexes an identifier
func (x *hclLex) lexId(yylval *hclSymType) int {
	var b bytes.Buffer
	for {
		c := x.next()
		if c == lexEOF {
			break
		}

		if unicode.IsSpace(c) {
			x.backup()
			break
		}

		if _, err := b.WriteRune(c); err != nil {
			return lexEOF
		}
	}

	yylval.str = b.String()

	switch yylval.str {
	case "true":
		yylval.b = true
		return BOOL
	case "false":
		yylval.b = false
		return BOOL
	}

	return IDENTIFIER
}

// lexNumber lexes out a number
func (x *hclLex) lexNumber(yylval *hclSymType) int {
	var b bytes.Buffer
	for {
		c := x.next()
		if c == lexEOF {
			break
		}

		// No more numeric characters
		if c < '0' || c > '9' {
			x.backup()
			break
		}

		if _, err := b.WriteRune(c); err != nil {
			x.createErr(fmt.Sprintf("Internal error: %s", err))
			return lexEOF
		}
	}

	v, err := strconv.ParseInt(b.String(), 0, 0)
	if err != nil {
		x.createErr(fmt.Sprintf("Expected number: %s", err))
		return lexEOF
	}

	yylval.num = int(v)
	return NUMBER
}

// lexString extracts a string from the input
func (x *hclLex) lexString(yylval *hclSymType) int {
	var b bytes.Buffer
	for {
		c := x.next()
		if c == lexEOF {
			break
		}

		// String end
		if c == '"' {
			break
		}

		if _, err := b.WriteRune(c); err != nil {
			return lexEOF
		}
	}

	yylval.str = b.String()
	return STRING
}

// Return the next rune for the lexer.
func (x *hclLex) next() rune {
	if int(x.pos) >= len(x.Input) {
		x.width = 0
		return lexEOF
	}

	r, w := utf8.DecodeRuneInString(x.Input[x.pos:])
	x.width = w
	x.pos += x.width

	x.col += 1
	if x.line == 0 {
		x.line = 1
	}
	if r == '\n' {
		x.line += 1
		x.col = 0
	}

	return r
}

// peek returns but does not consume the next rune in the input
func (x *hclLex) peek() rune {
	r := x.next()
	x.backup()
	return r
}

// backup steps back one rune. Can only be called once per next.
func (x *hclLex) backup() {
	x.col -= 1
	x.pos -= x.width
}

// createErr records the given error
func (x *hclLex) createErr(msg string) {
	x.err = fmt.Errorf("Line %d, column %d: %s", x.line, x.col, msg)
}

// The parser calls this method on a parse error.
func (x *hclLex) Error(s string) {
	x.createErr(s)
}
