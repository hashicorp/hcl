package hclsyntax

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
)

func TestCheckInvalidTokensTest(t *testing.T) {
	tests := []struct {
		Input       string
		WantSummary string
		WantDetail  string
	}{
		{
			`block “invalid” {}`,
			`Invalid character`,
			`"Curly quotes" are not valid here. These can sometimes be inadvertently introduced when sharing code via documents or discussion forums. It might help to replace the character with a "straight quote".`,
		},
		{
			`block 'invalid' {}`,
			`Invalid character`,
			`Single quotes are not valid. Use double quotes (") to enclose strings.`,
		},
		{
			"block `invalid` {}",
			`Invalid character`,
			"The \"`\" character is not valid. To create a multi-line string, use the \"heredoc\" syntax, like \"<<EOT\".",
		},
		{
			`foo = a & b`,
			`Unsupported operator`,
			`Bitwise operators are not supported. Did you mean boolean AND ("&&")?`,
		},
		{
			`foo = a | b`,
			`Unsupported operator`,
			`Bitwise operators are not supported. Did you mean boolean OR ("||")?`,
		},
		{
			`foo = ~a`,
			`Unsupported operator`,
			`Bitwise operators are not supported. Did you mean boolean NOT ("!")?`,
		},
	}

	for _, test := range tests {
		t.Run(test.Input, func(t *testing.T) {
			_, diags := LexConfig([]byte(test.Input), "", hcl.Pos{Line: 1, Column: 1})
			for _, diag := range diags {
				if diag.Severity == hcl.DiagError && diag.Summary == test.WantSummary && diag.Detail == test.WantDetail {
					return // success!
				}
			}
			// If we fall out here then we didn't find the diagnostic we were
			// looking for.
			t.Errorf("wrong errors\ngot:  %s\nwant: %s; %s", diags.Error(), test.WantSummary, test.WantDetail)
		})
	}
}
