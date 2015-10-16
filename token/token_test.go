package token

import "testing"

func TestTypeString(t *testing.T) {
	var tokens = []struct {
		tt  Type
		str string
	}{
		{ILLEGAL, "ILLEGAL"},
		{EOF, "EOF"},
		{COMMENT, "COMMENT"},
		{IDENT, "IDENT"},
		{NUMBER, "NUMBER"},
		{FLOAT, "FLOAT"},
		{BOOL, "BOOL"},
		{STRING, "STRING"},
		{LBRACK, "LBRACK"},
		{LBRACE, "LBRACE"},
		{COMMA, "COMMA"},
		{PERIOD, "PERIOD"},
		{RBRACK, "RBRACK"},
		{RBRACE, "RBRACE"},
		{ASSIGN, "ASSIGN"},
		{ADD, "ADD"},
		{SUB, "SUB"},
	}

	for _, token := range tokens {
		if token.tt.String() != token.str {
			t.Errorf("want: %q got:%q\n", token.str, token.tt)

		}
	}

}
