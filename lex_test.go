package hcl

import (
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLex(t *testing.T) {
	cases := []struct {
		Input  string
		Output []int
	}{
		{
			"comment.hcl",
			[]int{IDENTIFIER, EQUAL, STRING, lexEOF},
		},
		{
			"structure.hcl",
			[]int{
				IDENTIFIER, IDENTIFIER, STRING, LEFTBRACE,
				IDENTIFIER, EQUAL, NUMBER, SEMICOLON,
				RIGHTBRACE, lexEOF,
			},
		},
	}

	for _, tc := range cases {
		d, err := ioutil.ReadFile(filepath.Join(fixtureDir, tc.Input))
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		l := &hclLex{Input: string(d)}
		var actual []int
		for {
			token := l.Lex(new(hclSymType))
			actual = append(actual, token)

			if token == lexEOF {
				break
			}

			if len(actual) > 500 {
				t.Fatalf("Input:%s\n\nExausted.", tc.Input)
			}
		}

		if !reflect.DeepEqual(actual, tc.Output) {
			t.Fatalf("Input: %s\n\nBad: %#v", tc.Input, actual)
		}
	}
}
