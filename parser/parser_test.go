package parser

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/fatih/hcl/scanner"
)

func TestObjectKey(t *testing.T) {
	keys := []struct {
		exp []scanner.TokenType
		src string
	}{
		{[]scanner.TokenType{scanner.IDENT}, `foo {}`},
		{[]scanner.TokenType{scanner.IDENT}, `foo = {}`},
		{[]scanner.TokenType{scanner.IDENT}, `foo = "${var.bar}`},
		{[]scanner.TokenType{scanner.STRING}, `"foo" {}`},
		{[]scanner.TokenType{scanner.STRING}, `"foo" = {}`},
		{[]scanner.TokenType{scanner.STRING}, `"foo" = "${var.bar}`},
		{[]scanner.TokenType{scanner.IDENT, scanner.IDENT}, `foo bar {}`},
		{[]scanner.TokenType{scanner.IDENT, scanner.STRING}, `foo "bar" {}`},
		{[]scanner.TokenType{scanner.STRING, scanner.IDENT}, `"foo" bar {}`},
		{[]scanner.TokenType{scanner.IDENT, scanner.IDENT, scanner.IDENT}, `foo bar baz {}`},
	}

	for _, k := range keys {
		p := New([]byte(k.src))
		keys, err := p.parseObjectKey()
		if err != nil {
			t.Fatal(err)
		}

		tokens := []scanner.TokenType{}
		for _, o := range keys {
			tokens = append(tokens, o.token.Type)
		}

		equals(t, k.exp, tokens)
	}

	errKeys := []struct {
		src string
	}{
		{`foo 12 {}`},
		{`foo bar = {}`},
		{`foo []`},
		{`12 {}`},
	}

	for _, k := range errKeys {
		p := New([]byte(k.src))
		_, err := p.parseObjectKey()
		if err == nil {
			t.Errorf("case '%s' should give an error", k.src)
		}
	}
}

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}
