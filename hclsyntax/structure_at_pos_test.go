package hclsyntax

import (
	"reflect"
	"testing"

	"github.com/hashicorp/hcl/v2"
)

func TestBlocksAtPos(t *testing.T) {
	tests := map[string]struct {
		Src       string
		Pos       hcl.Pos
		WantTypes []string
	}{
		"empty": {
			``,
			hcl.Pos{Byte: 0},
			nil,
		},
		"spaces": {
			`    `,
			hcl.Pos{Byte: 1},
			nil,
		},
		"single in header": {
			`foo {}`,
			hcl.Pos{Byte: 1},
			[]string{"foo"},
		},
		"single in body": {
			`foo {    }`,
			hcl.Pos{Byte: 7},
			[]string{"foo"},
		},
		"single in body with unselected nested": {
			`
			foo {

				bar {

				}
			}
			`,
			hcl.Pos{Byte: 10},
			[]string{"foo"},
		},
		"single in body with unselected sibling": {
			`
			foo {  }
			bar {  }
			`,
			hcl.Pos{Byte: 10},
			[]string{"foo"},
		},
		"selected nested two levels": {
			`
			foo {
				bar {

				}
			}
			`,
			hcl.Pos{Byte: 20},
			[]string{"foo", "bar"},
		},
		"selected nested three levels": {
			`
			foo {
				bar {
					baz {

					}
				}
			}
			`,
			hcl.Pos{Byte: 31},
			[]string{"foo", "bar", "baz"},
		},
		"selected nested three levels with unselected sibling after": {
			`
			foo {
				bar {
					baz {

					}
				}
				not_wanted {}
			}
			`,
			hcl.Pos{Byte: 31},
			[]string{"foo", "bar", "baz"},
		},
		"selected nested three levels with unselected sibling before": {
			`
			foo {
				not_wanted {}
				bar {
					baz {

					}
				}
			}
			`,
			hcl.Pos{Byte: 49},
			[]string{"foo", "bar", "baz"},
		},
		"unterminated": {
			`foo {    `,
			hcl.Pos{Byte: 7},
			[]string{"foo"},
		},
		"unterminated nested": {
			`
			foo {
				bar {
			}
			`,
			hcl.Pos{Byte: 16},
			[]string{"foo", "bar"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			f, diags := ParseConfig([]byte(test.Src), "", hcl.Pos{Line: 1, Column: 1})
			for _, diag := range diags {
				// We intentionally ignore diagnostics here because we should be
				// able to work with the incomplete configuration that results
				// when the parser does its recovery behavior. However, we do
				// log them in case it's helpful to someone debugging a failing
				// test.
				t.Logf(diag.Error())
			}

			blocks := f.BlocksAtPos(test.Pos)
			outermost := f.OutermostBlockAtPos(test.Pos)
			innermost := f.InnermostBlockAtPos(test.Pos)

			gotTypes := make([]string, len(blocks))
			for i, block := range blocks {
				gotTypes[i] = block.Type
			}

			if len(test.WantTypes) == 0 {
				if len(gotTypes) != 0 {
					t.Errorf("wrong block types\ngot:  %#v\nwant: (none)", gotTypes)
				}
				if outermost != nil {
					t.Errorf("wrong outermost type\ngot:  %#v\nwant: (none)", outermost.Type)
				}
				if innermost != nil {
					t.Errorf("wrong innermost type\ngot:  %#v\nwant: (none)", innermost.Type)
				}
				return
			}

			if !reflect.DeepEqual(gotTypes, test.WantTypes) {
				if len(gotTypes) != 0 {
					t.Errorf("wrong block types\ngot:  %#v\nwant: %#v", gotTypes, test.WantTypes)
				}
			}
			if got, want := outermost.Type, test.WantTypes[0]; got != want {
				t.Errorf("wrong outermost type\ngot:  %#v\nwant: %#v", got, want)
			}
			if got, want := innermost.Type, test.WantTypes[len(test.WantTypes)-1]; got != want {
				t.Errorf("wrong innermost type\ngot:  %#v\nwant: %#v", got, want)
			}
		})
	}
}

func TestAttributeAtPos(t *testing.T) {
	tests := map[string]struct {
		Src      string
		Pos      hcl.Pos
		WantName string
	}{
		"empty": {
			``,
			hcl.Pos{Byte: 0},
			"",
		},
		"top-level": {
			`foo = 1`,
			hcl.Pos{Byte: 0},
			"foo",
		},
		"top-level with ignored sibling after": {
			`
			foo = 1
			bar = 2
			`,
			hcl.Pos{Byte: 6},
			"foo",
		},
		"top-level ignored sibling before": {
			`
			foo = 1
			bar = 2
			`,
			hcl.Pos{Byte: 17},
			"bar",
		},
		"nested": {
			`
			foo {
				bar = 2
			}
			`,
			hcl.Pos{Byte: 17},
			"bar",
		},
		"nested in unterminated block": {
			`
			foo {
				bar = 2
			`,
			hcl.Pos{Byte: 17},
			"bar",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			f, diags := ParseConfig([]byte(test.Src), "", hcl.Pos{Line: 1, Column: 1})
			for _, diag := range diags {
				// We intentionally ignore diagnostics here because we should be
				// able to work with the incomplete configuration that results
				// when the parser does its recovery behavior. However, we do
				// log them in case it's helpful to someone debugging a failing
				// test.
				t.Logf(diag.Error())
			}

			got := f.AttributeAtPos(test.Pos)

			if test.WantName == "" {
				if got != nil {
					t.Errorf("wrong attribute name\ngot:  %#v\nwant: (none)", got.Name)
				}
				return
			}

			if got == nil {
				t.Fatalf("wrong attribute name\ngot:  (none)\nwant: %#v", test.WantName)
			}

			if got.Name != test.WantName {
				t.Errorf("wrong attribute name\ngot:  %#v\nwant: %#v", got.Name, test.WantName)
			}
		})
	}
}

func TestOutermostExprAtPos(t *testing.T) {
	tests := map[string]struct {
		Src     string
		Pos     hcl.Pos
		WantSrc string
	}{
		"empty": {
			``,
			hcl.Pos{Byte: 0},
			``,
		},
		"simple bool": {
			`a = true`,
			hcl.Pos{Byte: 6},
			`true`,
		},
		"simple reference": {
			`a = blah`,
			hcl.Pos{Byte: 6},
			`blah`,
		},
		"attribute reference": {
			`a = blah.foo`,
			hcl.Pos{Byte: 6},
			`blah.foo`,
		},
		"parens": {
			`a = (1 + 1)`,
			hcl.Pos{Byte: 6},
			`(1 + 1)`,
		},
		"tuple cons": {
			`a = [1, 2, 3]`,
			hcl.Pos{Byte: 5},
			`[1, 2, 3]`,
		},
		"function call": {
			`a = foom("a")`,
			hcl.Pos{Byte: 10},
			`foom("a")`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			inputSrc := []byte(test.Src)
			f, diags := ParseConfig(inputSrc, "", hcl.Pos{Line: 1, Column: 1})
			for _, diag := range diags {
				// We intentionally ignore diagnostics here because we should be
				// able to work with the incomplete configuration that results
				// when the parser does its recovery behavior. However, we do
				// log them in case it's helpful to someone debugging a failing
				// test.
				t.Logf(diag.Error())
			}

			gotExpr := f.OutermostExprAtPos(test.Pos)
			var gotSrc string
			if gotExpr != nil {
				rng := gotExpr.Range()
				gotSrc = string(rng.SliceBytes(inputSrc))
			}

			if test.WantSrc == "" {
				if gotExpr != nil {
					t.Errorf("wrong expression source\ngot:  %s\nwant: (none)", gotSrc)
				}
				return
			}

			if gotExpr == nil {
				t.Fatalf("wrong expression source\ngot:  (none)\nwant: %s", test.WantSrc)
			}

			if gotSrc != test.WantSrc {
				t.Errorf("wrong expression source\ngot:  %#v\nwant: %#v", gotSrc, test.WantSrc)
			}
		})
	}
}
