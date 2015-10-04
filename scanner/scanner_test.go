package scanner

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/fatih/hcl/token"
)

var f100 = "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"

type tokenPair struct {
	tok  token.Token
	text string
}

func TestBool(t *testing.T) {
	var tokenList = []tokenPair{
		{token.BOOL, "true"},
		{token.BOOL, "false"},
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
			t.Errorf("text = %s want %s", l.TokenText(), ident.text)
		}

	}
}

func TestIdent(t *testing.T) {
	var tokenList = []tokenPair{
		{token.IDENT, "a"},
		{token.IDENT, "a0"},
		{token.IDENT, "foobar"},
		{token.IDENT, "abc123"},
		{token.IDENT, "LGTM"},
		{token.IDENT, "_"},
		{token.IDENT, "_abc123"},
		{token.IDENT, "abc123_"},
		{token.IDENT, "_abc_123_"},
		{token.IDENT, "_äöü"},
		{token.IDENT, "_本"},
		{token.IDENT, "äöü"},
		{token.IDENT, "本"},
		{token.IDENT, "a۰۱۸"},
		{token.IDENT, "foo६४"},
		{token.IDENT, "bar９８７６"},
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
			t.Errorf("text = %s want %s", l.TokenText(), ident.text)
		}

	}
}

func TestString(t *testing.T) {
	var tokenList = []tokenPair{
		{token.STRING, `" "`},
		{token.STRING, `"a"`},
		{token.STRING, `"本"`},
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
			t.Errorf("text = %s want %s", l.TokenText(), ident.text)
		}

	}
}
