package fuzzhclsyntax

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func FuzzParseTemplate(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		_, diags := hclsyntax.ParseTemplate(data, "<fuzz-tmpl>", hcl.Pos{Line: 1, Column: 1})

		if diags.HasErrors() {
			t.Logf("Error when parsing template %v", data)
			for _, diag := range diags {
				t.Logf("- %s", diag.Error())
			}
		}
	})
}

func FuzzParseTraversalAbs(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		_, diags := hclsyntax.ParseTraversalAbs(data, "<fuzz-trav>", hcl.Pos{Line: 1, Column: 1})

		if diags.HasErrors() {
			t.Logf("Error when parsing traversal %v", data)
			for _, diag := range diags {
				t.Logf("- %s", diag.Error())
			}
		}
	})
}

func FuzzParseExpression(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		_, diags := hclsyntax.ParseExpression(data, "<fuzz-expr>", hcl.Pos{Line: 1, Column: 1})

		if diags.HasErrors() {
			t.Logf("Error when parsing expression %v", data)
			for _, diag := range diags {
				t.Logf("- %s", diag.Error())
			}
		}
	})
}

func FuzzParseConfig(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		_, diags := hclsyntax.ParseConfig(data, "<fuzz-conf>", hcl.Pos{Line: 1, Column: 1})

		if diags.HasErrors() {
			t.Logf("Error when parsing config %v", data)
			for _, diag := range diags {
				t.Logf("- %s", diag.Error())
			}
		}
	})
}
