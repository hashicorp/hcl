// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package hcl

import "testing"

func TestDiagnosticError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		diag *Diagnostic
		want string
	}{
		{
			name: "with subject",
			diag: &Diagnostic{
				Severity: DiagError,
				Summary:  "Invalid attribute",
				Detail:   "The attribute is not expected here.",
				Subject: &Range{
					Filename: "example.hcl",
					Start:    Pos{Byte: 0, Line: 1, Column: 1},
					End:      Pos{Byte: 3, Line: 1, Column: 4},
				},
			},
			want: "example.hcl:1,1-4: Invalid attribute; The attribute is not expected here.",
		},
		{
			name: "nil subject",
			diag: &Diagnostic{
				Severity: DiagError,
				Summary:  "Configuration file not found",
				Detail:   `The configuration file foo.hcl does not exist.`,
			},
			want: "Configuration file not found; The configuration file foo.hcl does not exist.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.diag.Error()
			if got != tt.want {
				t.Fatalf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}
