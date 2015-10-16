package parser

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/fatih/hcl/ast"
	"github.com/fatih/hcl/token"
)

func TestParseType(t *testing.T) {
	src := `foo {
		fatih = "true"
}`

	p := New([]byte(src))
	p.enableTrace = true

	node, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	ast.Walk(node, func(n ast.Node) bool {
		if list, ok := n.(*ast.ObjectList); ok {
			for _, l := range list.Items {
				fmt.Printf("l = %+v\n", l)
				for _, k := range l.Keys {
					fmt.Printf("key = %+v\n", k)
				}
				fmt.Printf("val = %+v\n", l.Val)
			}
		}
		return true
	})
}

func TestObjectKey(t *testing.T) {
	keys := []struct {
		exp []token.Type
		src string
	}{
		{[]token.Type{token.IDENT}, `foo {}`},
		{[]token.Type{token.IDENT}, `foo = {}`},
		{[]token.Type{token.IDENT}, `foo = bar`},
		{[]token.Type{token.IDENT}, `foo = 123`},
		{[]token.Type{token.IDENT}, `foo = "${var.bar}`},
		{[]token.Type{token.STRING}, `"foo" {}`},
		{[]token.Type{token.STRING}, `"foo" = {}`},
		{[]token.Type{token.STRING}, `"foo" = "${var.bar}`},
		{[]token.Type{token.IDENT, token.IDENT}, `foo bar {}`},
		{[]token.Type{token.IDENT, token.STRING}, `foo "bar" {}`},
		{[]token.Type{token.STRING, token.IDENT}, `"foo" bar {}`},
		{[]token.Type{token.IDENT, token.IDENT, token.IDENT}, `foo bar baz {}`},
	}

	for _, k := range keys {
		p := New([]byte(k.src))
		keys, err := p.parseObjectKey()
		if err != nil {
			t.Fatal(err)
		}

		tokens := []token.Type{}
		for _, o := range keys {
			tokens = append(tokens, o.Token.Type)
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
