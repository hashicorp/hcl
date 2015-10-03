package parser

import (
	"bytes"
	"fmt"
	"testing"
)

var f100 = "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"

type token struct {
	tok  Token
	text string
}

func TestBool(t *testing.T) {
	var tokenList = []token{
		{BOOL, "true"},
		{BOOL, "false"},
	}

	// create artifical source code
	buf := new(bytes.Buffer)
	for _, ident := range tokenList {
		fmt.Fprintf(buf, " \t%s\n", ident.text)
	}

	l, err := NewLexer(buf)
	if err != nil {
		t.Fatal(err)
	}

	for _, ident := range tokenList {
		tok := l.Scan()
		if tok != ident.tok {
			t.Errorf("tok = %s want %s for %s\n", tok, ident.tok, ident.text)
		}

		if l.TokenText() != ident.text {
			t.Errorf("text = %s want %s", lit, ident.text)
		}

	}
}

func TestIdent(t *testing.T) {
	var tokenList = []token{
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
	for _, ident := range tokenList {
		fmt.Fprintf(buf, " \t%s\n", ident.text)
	}

	l, err := NewLexer(buf)
	if err != nil {
		t.Fatal(err)
	}

	for _, ident := range tokenList {
		tok := l.Scan()
		if tok != ident.tok {
			t.Errorf("tok = %s want %s for %s\n", tok, ident.tok, ident.text)
		}

		if l.TokenText() != ident.text {
			t.Errorf("text = %s want %s", lit, ident.text)
		}

	}
}

func TestString(t *testing.T) {
	var tokenList = []token{
		{STRING, `" "`},
		{STRING, `"a"`},
		{STRING, `"本"`},
		// {STRING, `"\a"`},
		// {STRING, `"\b"`},
		// {STRING, `"\f"`},
		// {STRING, `"\n"`},
		// {STRING, `"\r"`},
		// {STRING, `"\t"`},
		// {STRING, `"\v"`},
		// {STRING, `"\""`},
		// {STRING, `"\000"`},
		// {STRING, `"\777"`},
		// {STRING, `"\x00"`},
		// {STRING, `"\xff"`},
		// {STRING, `"\u0000"`},
		// {STRING, `"\ufA16"`},
		// {STRING, `"\U00000000"`},
		// {STRING, `"\U0000ffAB"`},
		// {STRING, `"` + f100 + `"`},
	}

	// create artifical source code
	buf := new(bytes.Buffer)
	for _, ident := range tokenList {
		fmt.Fprintf(buf, " \t%s\n", ident.text)
	}

	l, err := NewLexer(buf)
	if err != nil {
		t.Fatal(err)
	}

	for _, ident := range tokenList {
		tok := l.Scan()
		if tok != ident.tok {
			t.Errorf("tok = %s want %s for %s\n", tok, ident.tok, ident.text)
		}

		if l.TokenText() != ident.text {
			t.Errorf("text = %s want %s", lit, ident.text)
		}

	}
}
