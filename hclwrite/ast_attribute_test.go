// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package hclwrite

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
)

func TestAttributeLeadComments(t *testing.T) {
	tests := map[string]struct {
		src  string
		want string
	}{
		"basic comment": {`
# comment
test_attribute = foo
`,
			`# comment
`,
		},
		"basic multiline comment": {`
# multi-line
# comment (singe comment formatting)
test_attribute = foo
`,
			`# multi-line
# comment (singe comment formatting)
`,
		},
		"go formatted comment": {
			`
// comment
test_attribute = foo
`,
			`// comment
`,
		},
		"go formatted multi-line comment": {`
/* 
	go-style multi-line 
	comment 
*/
test_attribute = foo
`,
			``, //unsupported
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			f, diags := ParseConfig([]byte(test.src), "", hcl.Pos{Line: 1, Column: 1})
			if len(diags) != 0 {
				for _, diag := range diags {
					t.Logf("- %s", diag.Error())
				}
				t.Fatalf("unexpected diagnostics")
			}
			attr := f.Body().Attributes()["test_attribute"]
			got := string(attr.LeadComments().Bytes())
			if got != test.want {
				t.Errorf("wrong result\ngot:  %s\nwant: %s", got, test.want)
			}
		})
	}
}

func TestAttributeLineComments(t *testing.T) {
	tests := []struct {
		src  string
		want string
	}{
		{
			`
test_attribute = foo # comment
`,
			` # comment
`,
		},
		{
			`
test_attribute = foo // comment
`,
			` // comment
`,
		},
		{
			`
test_attribute = foo # multi-line
					 # comment (invalid)
`,
			` # multi-line
`, // known limitation: any extra comment lines (when using single-comment syntax) are dropped
		},
		{
			`
test_attribute = foo /* multi-line
                        comment in a weird place
                     */
`,
			` /* multi-line
                        comment in a weird place
                     */`, // note that all the whitespaces are returned
		},
	}

	for _, test := range tests {
		t.Run(test.want, func(t *testing.T) {
			f, diags := ParseConfig([]byte(test.src), "", hcl.Pos{Line: 1, Column: 1})
			if len(diags) != 0 {
				for _, diag := range diags {
					t.Logf("- %s", diag.Error())
				}
				t.Fatalf("unexpected diagnostics")
			}
			attr := f.Body().Attributes()["test_attribute"]
			got := string(attr.LineComments().Bytes())
			if got != test.want {
				t.Errorf("wrong result\ngot:  %s\nwant: %s", got, test.want)
			}
		})
	}
}
