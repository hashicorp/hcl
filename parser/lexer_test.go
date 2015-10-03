package parser

import (
	"bytes"
	"fmt"
	"testing"
)

type token struct {
	tok  Token
	text string
}

func TestIdent(t *testing.T) {
	var identList = []token{
		{IDENT, "a"},
		{IDENT, "a0"},
		{IDENT, "foobar"},
		{IDENT, "abc123"},
		{IDENT, "LGTM"},
		{IDENT, "_"},
		{IDENT, "_abc123"},
		{IDENT, "abc123_"},
		{IDENT, "_abc_123_"},
		{IDENT, "_äöü"},
		{IDENT, "_本"},
		{IDENT, "äöü"},
		{IDENT, "本"},
		{IDENT, "a۰۱۸"},
		{IDENT, "foo६४"},
		{IDENT, "bar９８７６"},
	}

	// create artifical source code
	buf := new(bytes.Buffer)
	for _, ident := range identList {
		fmt.Fprintf(buf, " \t%s\n", ident.text)
	}

	l := NewLexer(buf)

	for _, ident := range identList {
		tok, lit := l.Scan()
		if tok != ident.tok {
			t.Errorf("tok = %s want %s for %s\n", tok, ident.tok, ident.text)
		}

		if lit != ident.text {
			t.Errorf("text = %s want %s", lit, ident.text)
		}

	}
}
