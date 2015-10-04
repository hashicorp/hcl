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

func testTokenList(t *testing.T, tokenList []tokenPair) {
	// create artifical source code
	buf := new(bytes.Buffer)
	for _, ident := range tokenList {
		fmt.Fprintf(buf, " \t%s\n", ident.text)
	}

	s, err := NewScanner(buf)
	if err != nil {
		t.Fatal(err)
	}

	for _, ident := range tokenList {
		tok := s.Scan()
		if tok != ident.tok {
			t.Errorf("tok = %s want %s for %s\n", tok, ident.tok, ident.text)
		}

		if s.TokenText() != ident.text {
			t.Errorf("text = %s want %s", s.TokenText(), ident.text)
		}

	}
}

func TestBool(t *testing.T) {
	var tokenList = []tokenPair{
		{token.BOOL, "true"},
		{token.BOOL, "false"},
	}

	testTokenList(t, tokenList)
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

	testTokenList(t, tokenList)
}

func TestString(t *testing.T) {
	var tokenList = []tokenPair{
		{token.STRING, `" "`},
		{token.STRING, `"a"`},
		{token.STRING, `"本"`},
		{token.STRING, `"\a"`},
		{token.STRING, `"\b"`},
		{token.STRING, `"\f"`},
		{token.STRING, `"\n"`},
		{token.STRING, `"\r"`},
		{token.STRING, `"\t"`},
		{token.STRING, `"\v"`},
		{token.STRING, `"\""`},
		{token.STRING, `"\000"`},
		{token.STRING, `"\777"`},
		{token.STRING, `"\x00"`},
		{token.STRING, `"\xff"`},
		{token.STRING, `"\u0000"`},
		{token.STRING, `"\ufA16"`},
		{token.STRING, `"\U00000000"`},
		{token.STRING, `"\U0000ffAB"`},
		{token.STRING, `"` + f100 + `"`},
	}

	testTokenList(t, tokenList)
}

func TestNumber(t *testing.T) {
	var tokenList = []tokenPair{
		{token.NUMBER, "0"},
		{token.NUMBER, "1"},
		{token.NUMBER, "9"},
		{token.NUMBER, "42"},
		{token.NUMBER, "1234567890"},
		{token.NUMBER, "00"},
		{token.NUMBER, "01"},
		{token.NUMBER, "07"},
		{token.NUMBER, "042"},
		{token.NUMBER, "01234567"},
		{token.NUMBER, "0x0"},
		{token.NUMBER, "0x1"},
		{token.NUMBER, "0xf"},
		{token.NUMBER, "0x42"},
		{token.NUMBER, "0x123456789abcDEF"},
		{token.NUMBER, "0x" + f100},
		{token.NUMBER, "0X0"},
		{token.NUMBER, "0X1"},
		{token.NUMBER, "0XF"},
		{token.NUMBER, "0X42"},
		{token.NUMBER, "0X123456789abcDEF"},
		{token.NUMBER, "0X" + f100},
		// {token.FLOAT, "0."},
		// {token.FLOAT, "1."},
		// {token.FLOAT, "42."},
		// {token.FLOAT, "01234567890."},
		// {token.FLOAT, ".0"},
		// {token.FLOAT, ".1"},
		// {token.FLOAT, ".42"},
		// {token.FLOAT, ".0123456789"},
		// {token.FLOAT, "0.0"},
		// {token.FLOAT, "1.0"},
		// {token.FLOAT, "42.0"},
		// {token.FLOAT, "01234567890.0"},
		// {token.FLOAT, "0e0"},
		// {token.FLOAT, "1e0"},
		// {token.FLOAT, "42e0"},
		// {token.FLOAT, "01234567890e0"},
		// {token.FLOAT, "0E0"},
		// {token.FLOAT, "1E0"},
		// {token.FLOAT, "42E0"},
		// {token.FLOAT, "01234567890E0"},
		// {token.FLOAT, "0e+10"},
		// {token.FLOAT, "1e-10"},
		// {token.FLOAT, "42e+10"},
		// {token.FLOAT, "01234567890e-10"},
		// {token.FLOAT, "0E+10"},
		// {token.FLOAT, "1E-10"},
		// {token.FLOAT, "42E+10"},
		// {token.FLOAT, "01234567890E-10"},
	}

	testTokenList(t, tokenList)
}
