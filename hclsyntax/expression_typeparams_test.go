//go:build go1.18
// +build go1.18

package hclsyntax

import (
	"fmt"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// This file contains some additional tests that only make sense when using
// a Go compiler which supports type parameters (Go 1.18 or later).

func TestExpressionDiagnosticExtra(t *testing.T) {
	tests := []struct {
		input  string
		ctx    *hcl.EvalContext
		assert func(t *testing.T, diags hcl.Diagnostics)
	}{
		// Error messages describing inconsistent result types for conditional expressions.
		{
			"boop()",
			&hcl.EvalContext{
				Functions: map[string]function.Function{
					"boop": function.New(&function.Spec{
						Type: function.StaticReturnType(cty.String),
						Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
							return cty.DynamicVal, fmt.Errorf("the expected error")
						},
					}),
				},
			},
			func(t *testing.T, diags hcl.Diagnostics) {
				try := func(diags hcl.Diagnostics) {
					t.Helper()
					for _, diag := range diags {
						extra, ok := hcl.DiagnosticExtra[FunctionCallDiagExtra](diag)
						if !ok {
							continue
						}

						if got, want := extra.CalledFunctionName(), "boop"; got != want {
							t.Errorf("wrong called function name %q; want %q", got, want)
						}
						err := extra.FunctionCallError()
						if err == nil {
							t.Fatal("FunctionCallError returned nil")
						}
						if got, want := err.Error(), "the expected error"; got != want {
							t.Errorf("wrong error message\ngot:  %q\nwant: %q", got, want)
						}

						return
					}
					t.Fatalf("None of the returned diagnostics implement FunctionCallDiagError\n%s", diags.Error())
				}

				t.Run("unwrapped", func(t *testing.T) {
					try(diags)
				})

				// It should also work if we wrap up the "extras" in wrapper types.
				for _, diag := range diags {
					diag.Extra = diagnosticExtraWrapper{diag.Extra}
				}
				t.Run("wrapped", func(t *testing.T) {
					try(diags)
				})
			},
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			var diags hcl.Diagnostics
			expr, parseDiags := ParseExpression([]byte(test.input), "", hcl.Pos{Line: 1, Column: 1, Byte: 0})
			diags = append(diags, parseDiags...)
			_, valDiags := expr.Value(test.ctx)
			diags = append(diags, valDiags...)

			if !diags.HasErrors() {
				t.Fatal("unexpected success")
			}

			test.assert(t, diags)
		})
	}
}

type diagnosticExtraWrapper struct {
	wrapped interface{}
}

var _ hcl.DiagnosticExtraUnwrapper = diagnosticExtraWrapper{}

func (w diagnosticExtraWrapper) UnwrapDiagnosticExtra() interface{} {
	return w.wrapped
}
