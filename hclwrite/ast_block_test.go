package hclwrite

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
)

func TestBlockType(t *testing.T) {
	tests := []struct {
		src  string
		want string
	}{
		{
			`
service {
  attr0 = "val0"
}
`,
			"service",
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s", test.want), func(t *testing.T) {
			f, diags := ParseConfig([]byte(test.src), "", hcl.Pos{Line: 1, Column: 1})
			if len(diags) != 0 {
				for _, diag := range diags {
					t.Logf("- %s", diag.Error())
				}
				t.Fatalf("unexpected diagnostics")
			}

			block := f.Body().Blocks()[0]
			got := string(block.Type())
			if got != test.want {
				t.Errorf("wrong result\ngot:  %s\nwant: %s", got, test.want)
			}
		})
	}
}

func TestBlockLabels(t *testing.T) {
	tests := []struct {
		src  string
		want []string
	}{
		{
			`
nolabel {
}
`,
			[]string{},
		},
		{
			`
quoted "label1" {
}
`,
			[]string{"label1"},
		},
		{
			`
quoted "label1" "label2" {
}
`,
			[]string{"label1", "label2"},
		},
		{
			`
unquoted label1 {
}
`,
			[]string{"label1"},
		},
		{
			`
escape "\u0041" {
}
`,
			[]string{"\u0041"},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s", strings.Join(test.want, " ")), func(t *testing.T) {
			f, diags := ParseConfig([]byte(test.src), "", hcl.Pos{Line: 1, Column: 1})
			if len(diags) != 0 {
				for _, diag := range diags {
					t.Logf("- %s", diag.Error())
				}
				t.Fatalf("unexpected diagnostics")
			}

			block := f.Body().Blocks()[0]
			got := block.Labels()
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.want)
			}
		})
	}
}
