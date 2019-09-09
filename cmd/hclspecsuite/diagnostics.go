package main

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/hcl/v2"
)

func decodeJSONDiagnostics(src []byte) hcl.Diagnostics {
	type PosJSON struct {
		Line   int `json:"line"`
		Column int `json:"column"`
		Byte   int `json:"byte"`
	}
	type RangeJSON struct {
		Filename string  `json:"filename"`
		Start    PosJSON `json:"start"`
		End      PosJSON `json:"end"`
	}
	type DiagnosticJSON struct {
		Severity string     `json:"severity"`
		Summary  string     `json:"summary"`
		Detail   string     `json:"detail,omitempty"`
		Subject  *RangeJSON `json:"subject,omitempty"`
	}
	type DiagnosticsJSON struct {
		Diagnostics []DiagnosticJSON `json:"diagnostics"`
	}

	var raw DiagnosticsJSON
	var diags hcl.Diagnostics
	err := json.Unmarshal(src, &raw)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed to parse hcldec diagnostics result",
			Detail:   fmt.Sprintf("Sub-program hcldec produced invalid diagnostics: %s.", err),
		})
		return diags
	}

	if len(raw.Diagnostics) == 0 {
		return nil
	}

	diags = make(hcl.Diagnostics, 0, len(raw.Diagnostics))
	for _, rawDiag := range raw.Diagnostics {
		var severity hcl.DiagnosticSeverity
		switch rawDiag.Severity {
		case "error":
			severity = hcl.DiagError
		case "warning":
			severity = hcl.DiagWarning
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Failed to parse hcldec diagnostics result",
				Detail:   fmt.Sprintf("Diagnostic has unsupported severity %q.", rawDiag.Severity),
			})
			continue
		}

		diag := &hcl.Diagnostic{
			Severity: severity,
			Summary:  rawDiag.Summary,
			Detail:   rawDiag.Detail,
		}
		if rawDiag.Subject != nil {
			rawRange := rawDiag.Subject
			diag.Subject = &hcl.Range{
				Filename: rawRange.Filename,
				Start: hcl.Pos{
					Line:   rawRange.Start.Line,
					Column: rawRange.Start.Column,
					Byte:   rawRange.Start.Byte,
				},
				End: hcl.Pos{
					Line:   rawRange.End.Line,
					Column: rawRange.End.Column,
					Byte:   rawRange.End.Byte,
				},
			}
		}
		diags = append(diags, diag)
	}

	return diags
}

func severityString(severity hcl.DiagnosticSeverity) string {
	switch severity {
	case hcl.DiagError:
		return "error"
	case hcl.DiagWarning:
		return "warning"
	default:
		return "unsupported-severity"
	}
}

func rangeString(rng hcl.Range) string {
	return fmt.Sprintf(
		"from line %d column %d byte %d to line %d column %d byte %d",
		rng.Start.Line, rng.Start.Column, rng.Start.Byte,
		rng.End.Line, rng.End.Column, rng.End.Byte,
	)
}
