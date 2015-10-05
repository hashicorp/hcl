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
		fmt.Fprintf(buf, "%s\n", ident.text)
	}

	s, err := NewScanner(buf)
	if err != nil {
		t.Fatal(err)
	}

	for _, ident := range tokenList {
		tok := s.Scan()
		if tok != ident.tok {
			t.Errorf("tok = %q want %q for %q\n", tok, ident.tok, ident.text)
		}

		if s.TokenText() != ident.text {
			t.Errorf("text = %q want %q", s.TokenText(), ident.text)
		}

	}
}

func TestPosition(t *testing.T) {
	t.SkipNow()
	// create artifical source code
	buf := new(bytes.Buffer)
	for _, list := range tokenLists {
		for _, ident := range list {
			fmt.Fprintf(buf, "\t\t\t\t%s\n", ident.text)
		}
	}

	s, err := NewScanner(buf)
	if err != nil {
		t.Fatal(err)
	}

	s.Scan()
	pos := Position{"", 4, 1, 5}
	for _, list := range tokenLists {
		for _, k := range list {
			curPos := s.Pos()
			fmt.Printf("[%q] s = %+v:%+v\n", k.text, curPos.Offset, curPos.Column)
			if curPos.Offset != pos.Offset {
				t.Errorf("offset = %d, want %d for %q", curPos.Offset, pos.Offset, k.text)
			}
			if curPos.Line != pos.Line {
				t.Errorf("line = %d, want %d for %q", curPos.Line, pos.Line, k.text)
			}
			if curPos.Column != pos.Column {
				t.Errorf("column = %d, want %d for %q", curPos.Column, pos.Column, k.text)
			}
			pos.Offset += 4 + len(k.text) + 1     // 4 tabs + token bytes + newline
			pos.Line += countNewlines(k.text) + 1 // each token is on a new line
			s.Scan()
		}
	}
	// make sure there were no token-internal errors reported by scanner
	if s.ErrorCount != 0 {
		t.Errorf("%d errors", s.ErrorCount)
	}
}

var tokenLists = map[string][]tokenPair{
	// "comment": []tokenPair{
	// 	{token.COMMENT, "//"},
	// 	{token.COMMENT, "////"},
	// 	{token.COMMENT, "// comment"},
	// 	{token.COMMENT, "// /* comment */"},
	// 	{token.COMMENT, "// // comment //"},
	// 	{token.COMMENT, "//" + f100},
	// 	{token.COMMENT, "#"},
	// 	{token.COMMENT, "##"},
	// 	{token.COMMENT, "# comment"},
	// 	{token.COMMENT, "# /* comment */"},
	// 	{token.COMMENT, "# # comment #"},
	// 	{token.COMMENT, "#" + f100},
	// 	{token.COMMENT, "/**/"},
	// 	{token.COMMENT, "/***/"},
	// 	{token.COMMENT, "/* comment */"},
	// 	{token.COMMENT, "/* // comment */"},
	// 	{token.COMMENT, "/* /* comment */"},
	// 	{token.COMMENT, "/*\n comment\n*/"},
	// 	{token.COMMENT, "/*" + f100 + "*/"},
	// },
	// "operator": []tokenPair{
	// 	{token.LBRACK, "["},
	// 	{token.LBRACE, "{"},
	// 	{token.COMMA, ","},
	// 	{token.PERIOD, "."},
	// 	{token.RBRACK, "]"},
	// 	{token.RBRACE, "}"},
	// 	{token.ASSIGN, "="},
	// 	{token.ADD, "+"},
	// 	{token.SUB, "-"},
	// },
	// "bool": []tokenPair{
	// 	{token.BOOL, "true"},
	// 	{token.BOOL, "false"},
	// },

	"ident": []tokenPair{
		{token.IDENT, "a"},
		{token.IDENT, "a0"},
		{token.IDENT, "foobar"},
		{token.IDENT, "abc123"},
		{token.IDENT, "LGTM"},
		{token.IDENT, "_"},
		{token.IDENT, "_abc123"},
		{token.IDENT, "abc123_"},
		{token.IDENT, "_abc_123_"},
		// {token.IDENT, "_äöü"},
		// {token.IDENT, "_本"},
		// {token.IDENT, "äöü"},
		// {token.IDENT, "本"},
		// {token.IDENT, "a۰۱۸"},
		// {token.IDENT, "foo६४"},
		// {token.IDENT, "bar９８７６"},
	},
	// "string": []tokenPair{
	// 	{token.STRING, `" "`},
	// 	{token.STRING, `"a"`},
	// 	{token.STRING, `"本"`},
	// 	{token.STRING, `"\a"`},
	// 	{token.STRING, `"\b"`},
	// 	{token.STRING, `"\f"`},
	// 	{token.STRING, `"\n"`},
	// 	{token.STRING, `"\r"`},
	// 	{token.STRING, `"\t"`},
	// 	{token.STRING, `"\v"`},
	// 	{token.STRING, `"\""`},
	// 	{token.STRING, `"\000"`},
	// 	{token.STRING, `"\777"`},
	// 	{token.STRING, `"\x00"`},
	// 	{token.STRING, `"\xff"`},
	// 	{token.STRING, `"\u0000"`},
	// 	{token.STRING, `"\ufA16"`},
	// 	{token.STRING, `"\U00000000"`},
	// 	{token.STRING, `"\U0000ffAB"`},
	// 	{token.STRING, `"` + f100 + `"`},
	// },
	"number": []tokenPair{
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
		{token.NUMBER, "0e0"},
		{token.NUMBER, "1e0"},
		{token.NUMBER, "42e0"},
		{token.NUMBER, "01234567890e0"},
		{token.NUMBER, "0E0"},
		{token.NUMBER, "1E0"},
		{token.NUMBER, "42E0"},
		{token.NUMBER, "01234567890E0"},
		{token.NUMBER, "0e+10"},
		{token.NUMBER, "1e-10"},
		{token.NUMBER, "42e+10"},
		{token.NUMBER, "01234567890e-10"},
		{token.NUMBER, "0E+10"},
		{token.NUMBER, "1E-10"},
		{token.NUMBER, "42E+10"},
		{token.NUMBER, "01234567890E-10"},
	},
	"float": []tokenPair{
		{token.FLOAT, "0."},
		{token.FLOAT, "1."},
		{token.FLOAT, "42."},
		{token.FLOAT, "01234567890."},
		{token.FLOAT, ".0"},
		{token.FLOAT, ".1"},
		{token.FLOAT, ".42"},
		{token.FLOAT, ".0123456789"},
		{token.FLOAT, "0.0"},
		{token.FLOAT, "1.0"},
		{token.FLOAT, "42.0"},
		{token.FLOAT, "01234567890.0"},
		{token.FLOAT, "01.8e0"},
		{token.FLOAT, "1.4e0"},
		{token.FLOAT, "42.2e0"},
		{token.FLOAT, "01234567890.12e0"},
		{token.FLOAT, "0.E0"},
		{token.FLOAT, "1.12E0"},
		{token.FLOAT, "42.123E0"},
		{token.FLOAT, "01234567890.213E0"},
		{token.FLOAT, "0.2e+10"},
		{token.FLOAT, "1.2e-10"},
		{token.FLOAT, "42.54e+10"},
		{token.FLOAT, "01234567890.98e-10"},
		{token.FLOAT, "0.1E+10"},
		{token.FLOAT, "1.1E-10"},
		{token.FLOAT, "42.1E+10"},
		{token.FLOAT, "01234567890.1E-10"},
	},
}

func TestComment(t *testing.T) {
	testTokenList(t, tokenLists["comment"])
}

func TestOperator(t *testing.T) {
	testTokenList(t, tokenLists["operator"])
}

func TestBool(t *testing.T) {
	testTokenList(t, tokenLists["bool"])
}

func TestIdent(t *testing.T) {
	testTokenList(t, tokenLists["ident"])
}

func TestString(t *testing.T) {
	testTokenList(t, tokenLists["string"])
}

func TestNumber(t *testing.T) {
	testTokenList(t, tokenLists["number"])
}

func TestFloat(t *testing.T) {
	testTokenList(t, tokenLists["float"])
}

func countNewlines(s string) int {
	n := 0
	for _, ch := range s {
		if ch == '\n' {
			n++
		}
	}
	return n
}
