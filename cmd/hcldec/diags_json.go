package main

import (
	"encoding/json"
	"io"

	"github.com/hashicorp/hcl/v2"
)

type jsonDiagWriter struct {
	w     io.Writer
	diags hcl.Diagnostics
}

var _ hcl.DiagnosticWriter = &jsonDiagWriter{}

func (wr *jsonDiagWriter) WriteDiagnostic(diag *hcl.Diagnostic) error {
	wr.diags = append(wr.diags, diag)
	return nil
}

func (wr *jsonDiagWriter) WriteDiagnostics(diags hcl.Diagnostics) error {
	wr.diags = append(wr.diags, diags...)
	return nil
}

func (wr *jsonDiagWriter) Flush() error {
	if len(wr.diags) == 0 {
		return nil
	}

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

	diagsJSON := make([]DiagnosticJSON, 0, len(wr.diags))
	for _, diag := range wr.diags {
		var diagJSON DiagnosticJSON

		switch diag.Severity {
		case hcl.DiagError:
			diagJSON.Severity = "error"
		case hcl.DiagWarning:
			diagJSON.Severity = "warning"
		default:
			diagJSON.Severity = "(unknown)" // should never happen
		}

		diagJSON.Summary = diag.Summary
		diagJSON.Detail = diag.Detail
		if diag.Subject != nil {
			diagJSON.Subject = &RangeJSON{}
			sJSON := diagJSON.Subject
			rng := diag.Subject
			sJSON.Filename = rng.Filename
			sJSON.Start.Line = rng.Start.Line
			sJSON.Start.Column = rng.Start.Column
			sJSON.Start.Byte = rng.Start.Byte
			sJSON.End.Line = rng.End.Line
			sJSON.End.Column = rng.End.Column
			sJSON.End.Byte = rng.End.Byte
		}

		diagsJSON = append(diagsJSON, diagJSON)
	}

	src, err := json.MarshalIndent(DiagnosticsJSON{diagsJSON}, "", "  ")
	if err != nil {
		return err
	}
	_, err = wr.w.Write(src)
	wr.w.Write([]byte{'\n'})
	return err
}

type flusher interface {
	Flush() error
}

func flush(maybeFlusher interface{}) error {
	if f, ok := maybeFlusher.(flusher); ok {
		return f.Flush()
	}
	return nil
}
