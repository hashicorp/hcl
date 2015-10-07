package scanner

import (
	"bytes"
	"fmt"
	"testing"
)

var f100 = "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"

type tokenPair struct {
	tok  TokenType
	text string
}

var tokenLists = map[string][]tokenPair{
	"comment": []tokenPair{
		{COMMENT, "//"},
		{COMMENT, "////"},
		{COMMENT, "// comment"},
		{COMMENT, "// /* comment */"},
		{COMMENT, "// // comment //"},
		{COMMENT, "//" + f100},
		{COMMENT, "#"},
		{COMMENT, "##"},
		{COMMENT, "# comment"},
		{COMMENT, "# /* comment */"},
		{COMMENT, "# # comment #"},
		{COMMENT, "#" + f100},
		{COMMENT, "/**/"},
		{COMMENT, "/***/"},
		{COMMENT, "/* comment */"},
		{COMMENT, "/* // comment */"},
		{COMMENT, "/* /* comment */"},
		{COMMENT, "/*\n comment\n*/"},
		{COMMENT, "/*" + f100 + "*/"},
	},
	"operator": []tokenPair{
		{LBRACK, "["},
		{LBRACE, "{"},
		{COMMA, ","},
		{PERIOD, "."},
		{RBRACK, "]"},
		{RBRACE, "}"},
		{ASSIGN, "="},
		{ADD, "+"},
		{SUB, "-"},
	},
	"bool": []tokenPair{
		{BOOL, "true"},
		{BOOL, "false"},
	},
	"ident": []tokenPair{
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
	},
	"string": []tokenPair{
		{STRING, `" "`},
		{STRING, `"a"`},
		{STRING, `"本"`},
		{STRING, `"\a"`},
		{STRING, `"\b"`},
		{STRING, `"\f"`},
		{STRING, `"\n"`},
		{STRING, `"\r"`},
		{STRING, `"\t"`},
		{STRING, `"\v"`},
		{STRING, `"\""`},
		{STRING, `"\000"`},
		{STRING, `"\777"`},
		{STRING, `"\x00"`},
		{STRING, `"\xff"`},
		{STRING, `"\u0000"`},
		{STRING, `"\ufA16"`},
		{STRING, `"\U00000000"`},
		{STRING, `"\U0000ffAB"`},
		{STRING, `"` + f100 + `"`},
	},
	"number": []tokenPair{
		{NUMBER, "0"},
		{NUMBER, "1"},
		{NUMBER, "9"},
		{NUMBER, "42"},
		{NUMBER, "1234567890"},
		{NUMBER, "00"},
		{NUMBER, "01"},
		{NUMBER, "07"},
		{NUMBER, "042"},
		{NUMBER, "01234567"},
		{NUMBER, "0x0"},
		{NUMBER, "0x1"},
		{NUMBER, "0xf"},
		{NUMBER, "0x42"},
		{NUMBER, "0x123456789abcDEF"},
		{NUMBER, "0x" + f100},
		{NUMBER, "0X0"},
		{NUMBER, "0X1"},
		{NUMBER, "0XF"},
		{NUMBER, "0X42"},
		{NUMBER, "0X123456789abcDEF"},
		{NUMBER, "0X" + f100},
		{NUMBER, "0e0"},
		{NUMBER, "1e0"},
		{NUMBER, "42e0"},
		{NUMBER, "01234567890e0"},
		{NUMBER, "0E0"},
		{NUMBER, "1E0"},
		{NUMBER, "42E0"},
		{NUMBER, "01234567890E0"},
		{NUMBER, "0e+10"},
		{NUMBER, "1e-10"},
		{NUMBER, "42e+10"},
		{NUMBER, "01234567890e-10"},
		{NUMBER, "0E+10"},
		{NUMBER, "1E-10"},
		{NUMBER, "42E+10"},
		{NUMBER, "01234567890E-10"},
	},
	"float": []tokenPair{
		{FLOAT, "0."},
		{FLOAT, "1."},
		{FLOAT, "42."},
		{FLOAT, "01234567890."},
		{FLOAT, ".0"},
		{FLOAT, ".1"},
		{FLOAT, ".42"},
		{FLOAT, ".0123456789"},
		{FLOAT, "0.0"},
		{FLOAT, "1.0"},
		{FLOAT, "42.0"},
		{FLOAT, "01234567890.0"},
		{FLOAT, "01.8e0"},
		{FLOAT, "1.4e0"},
		{FLOAT, "42.2e0"},
		{FLOAT, "01234567890.12e0"},
		{FLOAT, "0.E0"},
		{FLOAT, "1.12E0"},
		{FLOAT, "42.123E0"},
		{FLOAT, "01234567890.213E0"},
		{FLOAT, "0.2e+10"},
		{FLOAT, "1.2e-10"},
		{FLOAT, "42.54e+10"},
		{FLOAT, "01234567890.98e-10"},
		{FLOAT, "0.1E+10"},
		{FLOAT, "1.1E-10"},
		{FLOAT, "42.1E+10"},
		{FLOAT, "01234567890.1E-10"},
	},
}

var orderedTokenLists = []string{
	"comment",
	"operator",
	"bool",
	"ident",
	"string",
	"number",
	"float",
}

func TestPosition(t *testing.T) {
	// create artifical source code
	buf := new(bytes.Buffer)

	for _, listName := range orderedTokenLists {
		for _, ident := range tokenLists[listName] {
			fmt.Fprintf(buf, "\t\t\t\t%s\n", ident.text)
		}
	}

	s := NewScanner(buf.Bytes())

	pos := Pos{"", 4, 1, 5}
	s.Scan()
	for _, listName := range orderedTokenLists {

		for _, k := range tokenLists[listName] {
			curPos := s.tokPos
			// fmt.Printf("[%q] s = %+v:%+v\n", k.text, curPos.Offset, curPos.Column)

			if curPos.Offset != pos.Offset {
				t.Fatalf("offset = %d, want %d for %q", curPos.Offset, pos.Offset, k.text)
			}
			if curPos.Line != pos.Line {
				t.Fatalf("line = %d, want %d for %q", curPos.Line, pos.Line, k.text)
			}
			if curPos.Column != pos.Column {
				t.Fatalf("column = %d, want %d for %q", curPos.Column, pos.Column, k.text)
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

func TestRealExample(t *testing.T) {
	complexHCL := `// This comes from Terraform, as a test
	variable "foo" {
	    default = "bar"
	    description = "bar"
	}

	provider "aws" {
	  access_key = "foo"
	  secret_key = "bar"
	}

	resource "aws_security_group" "firewall" {
	    count = 5
	}

	resource aws_instance "web" {
	    ami = "${var.foo}"
	    security_groups = [
	        "foo",
	        "${aws_security_group.firewall.foo}"
	    ]

	    network_interface {
	        device_index = 0
	        description = "Main network interface"
	    }
	}`

	literals := []struct {
		token   TokenType
		literal string
	}{
		{COMMENT, `// This comes from Terraform, as a test`},
		{IDENT, `variable`},
		{STRING, `"foo"`},
		{LBRACE, `{`},
		{IDENT, `default`},
		{ASSIGN, `=`},
		{STRING, `"bar"`},
		{IDENT, `description`},
		{ASSIGN, `=`},
		{STRING, `"bar"`},
		{RBRACE, `}`},
		{IDENT, `provider`},
		{STRING, `"aws"`},
		{LBRACE, `{`},
		{IDENT, `access_key`},
		{ASSIGN, `=`},
		{STRING, `"foo"`},
		{IDENT, `secret_key`},
		{ASSIGN, `=`},
		{STRING, `"bar"`},
		{RBRACE, `}`},
		{IDENT, `resource`},
		{STRING, `"aws_security_group"`},
		{STRING, `"firewall"`},
		{LBRACE, `{`},
		{IDENT, `count`},
		{ASSIGN, `=`},
		{NUMBER, `5`},
		{RBRACE, `}`},
		{IDENT, `resource`},
		{IDENT, `aws_instance`},
		{STRING, `"web"`},
		{LBRACE, `{`},
		{IDENT, `ami`},
		{ASSIGN, `=`},
		{STRING, `"${var.foo}"`},
		{IDENT, `security_groups`},
		{ASSIGN, `=`},
		{LBRACK, `[`},
		{STRING, `"foo"`},
		{COMMA, `,`},
		{STRING, `"${aws_security_group.firewall.foo}"`},
		{RBRACK, `]`},
		{IDENT, `network_interface`},
		{LBRACE, `{`},
		{IDENT, `device_index`},
		{ASSIGN, `=`},
		{NUMBER, `0`},
		{IDENT, `description`},
		{ASSIGN, `=`},
		{STRING, `"Main network interface"`},
		{RBRACE, `}`},
		{RBRACE, `}`},
		{EOF, ``},
	}

	s := NewScanner([]byte(complexHCL))
	for _, l := range literals {
		tok := s.Scan()
		if l.token != tok.Type() {
			t.Errorf("got: %s want %s for %s\n", tok, l.token, tok.String())
		}

		if l.literal != tok.String() {
			t.Errorf("got: %s want %s\n", tok, l.literal)
		}
	}

}

func TestError(t *testing.T) {
	testError(t, "\x80", "1:1", "illegal UTF-8 encoding", ILLEGAL)
	testError(t, "\xff", "1:1", "illegal UTF-8 encoding", ILLEGAL)

	testError(t, "ab\x80", "1:3", "illegal UTF-8 encoding", IDENT)
	testError(t, "abc\xff", "1:4", "illegal UTF-8 encoding", IDENT)

	testError(t, `"ab`+"\x80", "1:4", "illegal UTF-8 encoding", STRING)
	testError(t, `"abc`+"\xff", "1:5", "illegal UTF-8 encoding", STRING)

	testError(t, `01238`, "1:6", "illegal octal number", NUMBER)
	testError(t, `01238123`, "1:9", "illegal octal number", NUMBER)
	testError(t, `0x`, "1:3", "illegal hexadecimal number", NUMBER)
	testError(t, `0xg`, "1:3", "illegal hexadecimal number", NUMBER)
	testError(t, `'aa'`, "1:1", "illegal char", ILLEGAL)

	testError(t, `"`, "1:2", "literal not terminated", STRING)
	testError(t, `"abc`, "1:5", "literal not terminated", STRING)
	testError(t, `"abc`+"\n", "1:5", "literal not terminated", STRING)
	testError(t, `/*/`, "1:4", "comment not terminated", COMMENT)
}

func testError(t *testing.T, src, pos, msg string, tok TokenType) {
	s := NewScanner([]byte(src))

	errorCalled := false
	s.Error = func(p Pos, m string) {
		if !errorCalled {
			if pos != p.String() {
				t.Errorf("pos = %q, want %q for %q", p, pos, src)
			}

			if m != msg {
				t.Errorf("msg = %q, want %q for %q", m, msg, src)
			}
			errorCalled = true
		}
	}

	tk := s.Scan()
	if tk.Type() != tok {
		t.Errorf("tok = %s, want %s for %q", tk, tok, src)
	}
	if !errorCalled {
		t.Errorf("error handler not called for %q", src)
	}
	if s.ErrorCount == 0 {
		t.Errorf("count = %d, want > 0 for %q", s.ErrorCount, src)
	}
}

func testTokenList(t *testing.T, tokenList []tokenPair) {
	// create artifical source code
	buf := new(bytes.Buffer)
	for _, ident := range tokenList {
		fmt.Fprintf(buf, "%s\n", ident.text)
	}

	s := NewScanner(buf.Bytes())
	for _, ident := range tokenList {
		tok := s.Scan()
		if tok.Type() != ident.tok {
			t.Errorf("tok = %q want %q for %q\n", tok, ident.tok, ident.text)
		}

		if tok.String() != ident.text {
			t.Errorf("text = %q want %q", tok.String(), ident.text)
		}

	}
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
