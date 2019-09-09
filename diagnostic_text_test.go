package hcl

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestDiagnosticTextWriter(t *testing.T) {
	tests := []struct {
		Input *Diagnostic
		Want  string
	}{
		{
			&Diagnostic{
				Severity: DiagError,
				Summary:  "Splines not reticulated",
				Detail:   "All splines must be pre-reticulated.",
				Subject: &Range{
					Start: Pos{
						Byte:   0,
						Column: 1,
						Line:   1,
					},
					End: Pos{
						Byte:   3,
						Column: 4,
						Line:   1,
					},
				},
			},
			`Error: Splines not reticulated

  on  line 1, in hardcoded-context:
   1: foo = 1

All splines must be pre-reticulated.

`,
		},
		{
			&Diagnostic{
				Severity: DiagError,
				Summary:  "Unsupported attribute",
				Detail:   `"baz" is not a supported top-level attribute. Did you mean "bam"?`,
				Subject: &Range{
					Start: Pos{
						Byte:   16,
						Column: 1,
						Line:   3,
					},
					End: Pos{
						Byte:   19,
						Column: 4,
						Line:   3,
					},
				},
			},
			`Error: Unsupported attribute

  on  line 3, in hardcoded-context:
   3: baz = 3

"baz" is not a supported top-level
attribute. Did you mean "bam"?

`,
		},
		{
			&Diagnostic{
				Severity: DiagError,
				Summary:  "Unsupported attribute",
				Detail:   `"pizza" is not a supported attribute. Did you mean "pizzetta"?`,
				Subject: &Range{
					Start: Pos{
						Byte:   42,
						Column: 3,
						Line:   5,
					},
					End: Pos{
						Byte:   47,
						Column: 8,
						Line:   5,
					},
				},
				// This is actually not a great example of a context, but is here to test
				// whether we're able to show a multi-line context when needed.
				Context: &Range{
					Start: Pos{
						Byte:   24,
						Column: 1,
						Line:   4,
					},
					End: Pos{
						Byte:   60,
						Column: 2,
						Line:   6,
					},
				},
			},
			`Error: Unsupported attribute

  on  line 5, in hardcoded-context:
   4: block "party" {
   5:   pizza = "cheese"
   6: }

"pizza" is not a supported attribute.
Did you mean "pizzetta"?

`,
		},
		{
			&Diagnostic{
				Severity: DiagError,
				Summary:  "Test of including relevant variable values",
				Detail:   `This diagnostic includes an expression and an evalcontext.`,
				Subject: &Range{
					Start: Pos{
						Byte:   42,
						Column: 3,
						Line:   5,
					},
					End: Pos{
						Byte:   47,
						Column: 8,
						Line:   5,
					},
				},
				Expression: &diagnosticTestExpr{
					vars: []Traversal{
						{
							TraverseRoot{
								Name: "foo",
							},
						},
						{
							TraverseRoot{
								Name: "bar",
							},
							TraverseAttr{
								Name: "baz",
							},
						},
						{
							TraverseRoot{
								Name: "missing",
							},
						},
						{
							TraverseRoot{
								Name: "boz",
							},
						},
					},
				},
				EvalContext: &EvalContext{
					parent: &EvalContext{
						Variables: map[string]cty.Value{
							"foo": cty.StringVal("foo value"),
						},
					},
					Variables: map[string]cty.Value{
						"bar": cty.ObjectVal(map[string]cty.Value{
							"baz": cty.ListValEmpty(cty.String),
						}),
						"boz":    cty.NumberIntVal(5),
						"unused": cty.True,
					},
				},
			},
			`Error: Test of including relevant variable values

  on  line 5, in hardcoded-context:
   5:   pizza = "cheese"

with bar.baz as empty list of string,
     boz as 5,
     foo as "foo value".

This diagnostic includes an expression
and an evalcontext.

`,
		},
	}

	files := map[string]*File{
		"": &File{
			Bytes: []byte(testDiagnosticTextWriterSource),
			Nav:   &diagnosticTestNav{},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			bwr := &bytes.Buffer{}
			dwr := NewDiagnosticTextWriter(bwr, files, 40, false)
			err := dwr.WriteDiagnostic(test.Input)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			got := bwr.String()
			if got != test.Want {
				t.Errorf("wrong result\n\ngot:\n%swant:\n%s", got, test.Want)
			}
		})
	}
}

const testDiagnosticTextWriterSource = `foo = 1
bar = 2
baz = 3
block "party" {
  pizza = "cheese"
}
`

type diagnosticTestNav struct {
}

func (tn *diagnosticTestNav) ContextString(offset int) string {
	return "hardcoded-context"
}

type diagnosticTestExpr struct {
	vars []Traversal
	staticExpr
}

func (e *diagnosticTestExpr) Variables() []Traversal {
	return e.vars
}
