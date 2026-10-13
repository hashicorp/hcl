// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package hclwrite

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func TestExpressionVariablesObjectCons(t *testing.T) {
	tests := []struct {
		src        string
		wantVars   []string
		wantRename string
	}{
		{
			src:        "a = { k = old.x }\n",
			wantVars:   []string{"old.x"},
			wantRename: "a = { k = new.x }\n",
		},
		{
			src:        "a = {\n  k = old.x\n  j = old.y\n}\n",
			wantVars:   []string{"old.x", "old.y"},
			wantRename: "a = {\n  k = new.x\n  j = new.y\n}\n",
		},
		{
			src:        "a = { k = { j = old.x } }\n",
			wantVars:   []string{"old.x"},
			wantRename: "a = { k = { j = new.x } }\n",
		},
		{
			src:        "a = { k = [old.x] }\n",
			wantVars:   []string{"old.x"},
			wantRename: "a = { k = [new.x] }\n",
		},
		{
			src:        "a = { k = f(old.x) }\n",
			wantVars:   []string{"old.x"},
			wantRename: "a = { k = f(new.x) }\n",
		},
		{
			src:        "a = { \"k\" = old.x }\n",
			wantVars:   []string{"old.x"},
			wantRename: "a = { \"k\" = new.x }\n",
		},
		{
			src:        "a = { old = 1 }\n",
			wantVars:   nil,
			wantRename: "a = { old = 1 }\n",
		},
		{
			src:        "a = { (old.x) = 1 }\n",
			wantVars:   []string{"old.x"},
			wantRename: "a = { (new.x) = 1 }\n",
		},
		{
			src:        "a = { k = 1 }\n",
			wantVars:   nil,
			wantRename: "a = { k = 1 }\n",
		},
	}

	for _, test := range tests {
		t.Run(test.src, func(t *testing.T) {
			f, diags := ParseConfig([]byte(test.src), "", hcl.InitialPos)
			if diags.HasErrors() {
				t.Fatalf("unexpected diagnostics: %s", diags.Error())
			}
			expr := f.Body().GetAttribute("a").Expr()

			var got []string
			for _, traversal := range expr.Variables() {
				got = append(got, strings.TrimSpace(string(traversal.BuildTokens(nil).Bytes())))
			}
			if diff := cmp.Diff(test.wantVars, got); diff != "" {
				t.Errorf("wrong variables\n%s", diff)
			}

			nativeFile, nativeDiags := hclsyntax.ParseConfig([]byte(test.src), "", hcl.InitialPos)
			if nativeDiags.HasErrors() {
				t.Fatalf("unexpected native diagnostics: %s", nativeDiags.Error())
			}
			nativeExpr := nativeFile.Body.(*hclsyntax.Body).Attributes["a"].Expr
			if got, want := len(expr.Variables()), len(nativeExpr.Variables()); got != want {
				t.Errorf("hclwrite found %d variables but hclsyntax found %d", got, want)
			}

			expr.RenameVariablePrefix([]string{"old"}, []string{"new"})
			if got := string(f.Bytes()); got != test.wantRename {
				t.Errorf("wrong result after rename\ngot:  %q\nwant: %q", got, test.wantRename)
			}
		})
	}
}
