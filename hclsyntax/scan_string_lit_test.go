package hclsyntax

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestScanStringLit(t *testing.T) {
	tests := []struct {
		Input        string
		WantQuoted   []string
		WantUnquoted []string
	}{
		{
			``,
			[]string{},
			[]string{},
		},
		{
			`hello`,
			[]string{`hello`},
			[]string{`hello`},
		},
		{
			`hello world`,
			[]string{`hello world`},
			[]string{`hello world`},
		},
		{
			`hello\nworld`,
			[]string{`hello`, `\n`, `world`},
			[]string{`hello\nworld`},
		},
		{
			`hello\🥁world`,
			[]string{`hello`, `\🥁`, `world`},
			[]string{`hello\🥁world`},
		},
		{
			`hello\uabcdworld`,
			[]string{`hello`, `\uabcd`, `world`},
			[]string{`hello\uabcdworld`},
		},
		{
			`hello\uabcdabcdworld`,
			[]string{`hello`, `\uabcd`, `abcdworld`},
			[]string{`hello\uabcdabcdworld`},
		},
		{
			`hello\uabcworld`,
			[]string{`hello`, `\uabc`, `world`},
			[]string{`hello\uabcworld`},
		},
		{
			`hello\U01234567world`,
			[]string{`hello`, `\U01234567`, `world`},
			[]string{`hello\U01234567world`},
		},
		{
			`hello\U012345670123world`,
			[]string{`hello`, `\U01234567`, `0123world`},
			[]string{`hello\U012345670123world`},
		},
		{
			`hello\Uabcdworld`,
			[]string{`hello`, `\Uabcd`, `world`},
			[]string{`hello\Uabcdworld`},
		},
		{
			`hello\Uabcworld`,
			[]string{`hello`, `\Uabc`, `world`},
			[]string{`hello\Uabcworld`},
		},
		{
			`hello\uworld`,
			[]string{`hello`, `\u`, `world`},
			[]string{`hello\uworld`},
		},
		{
			`hello\Uworld`,
			[]string{`hello`, `\U`, `world`},
			[]string{`hello\Uworld`},
		},
		{
			`hello\u`,
			[]string{`hello`, `\u`},
			[]string{`hello\u`},
		},
		{
			`hello\U`,
			[]string{`hello`, `\U`},
			[]string{`hello\U`},
		},
		{
			`hello\`,
			[]string{`hello`, `\`},
			[]string{`hello\`},
		},
		{
			`hello$${world}`,
			[]string{`hello`, `$${`, `world}`},
			[]string{`hello`, `$${`, `world}`},
		},
		{
			`hello$$world`,
			[]string{`hello`, `$$`, `world`},
			[]string{`hello`, `$$`, `world`},
		},
		{
			`hello$world`,
			[]string{`hello`, `$`, `world`},
			[]string{`hello`, `$`, `world`},
		},
		{
			`hello$`,
			[]string{`hello`, `$`},
			[]string{`hello`, `$`},
		},
		{
			`hello$${`,
			[]string{`hello`, `$${`},
			[]string{`hello`, `$${`},
		},
		{
			`hello%%{world}`,
			[]string{`hello`, `%%{`, `world}`},
			[]string{`hello`, `%%{`, `world}`},
		},
		{
			`hello%%world`,
			[]string{`hello`, `%%`, `world`},
			[]string{`hello`, `%%`, `world`},
		},
		{
			`hello%world`,
			[]string{`hello`, `%`, `world`},
			[]string{`hello`, `%`, `world`},
		},
		{
			`hello%`,
			[]string{`hello`, `%`},
			[]string{`hello`, `%`},
		},
		{
			`hello%%{`,
			[]string{`hello`, `%%{`},
			[]string{`hello`, `%%{`},
		},
		{
			`hello\${world}`,
			[]string{`hello`, `\$`, `{world}`},
			[]string{`hello\`, `$`, `{world}`},
		},
		{
			`hello\%{world}`,
			[]string{`hello`, `\%`, `{world}`},
			[]string{`hello\`, `%`, `{world}`},
		},
		{
			"hello\nworld",
			[]string{`hello`, "\n", `world`},
			[]string{`hello`, "\n", `world`},
		},
		{
			"hello\rworld",
			[]string{`hello`, "\r", `world`},
			[]string{`hello`, "\r", `world`},
		},
		{
			"hello\r\nworld",
			[]string{`hello`, "\r\n", `world`},
			[]string{`hello`, "\r\n", `world`},
		},
	}

	for _, test := range tests {
		t.Run(test.Input, func(t *testing.T) {
			t.Run("quoted", func(t *testing.T) {
				slices := scanStringLit([]byte(test.Input), true)
				got := make([]string, len(slices))
				for i, slice := range slices {
					got[i] = string(slice)
				}
				if !reflect.DeepEqual(got, test.WantQuoted) {
					t.Errorf("wrong result\ngot: %swant: %s", spew.Sdump(got), spew.Sdump(test.WantQuoted))
				}
			})
			t.Run("unquoted", func(t *testing.T) {
				slices := scanStringLit([]byte(test.Input), false)
				got := make([]string, len(slices))
				for i, slice := range slices {
					got[i] = string(slice)
				}
				if !reflect.DeepEqual(got, test.WantUnquoted) {
					t.Errorf("wrong result\ngot: %swant: %s", spew.Sdump(got), spew.Sdump(test.WantUnquoted))
				}
			})
		})
	}
}
