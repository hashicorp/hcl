// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package printer

import (
	"bytes"
	"strings"
	"testing"
)

// TestFormatTrailingListComments verifies that formatting preserves comments
// inside their original lists, including template directives.
func TestFormatTrailingListComments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		source string
		marker string
	}{
		{
			name: "customer policy with surrounding attributes",
			source: `path "auth/token/create/${role}" {
  allowed_parameters = {
    "meta" = []
    "entity_alias" = [
      # %{ for alias in aliases ~}
      "${alias}",
      # %{ endfor ~}
    ]
    "display_name" = []
    "num_uses" = []
  }
}
`,
			marker: "# %{ endfor ~}",
		},
		{
			name: "last element inline comment",
			source: `values = [
  "first", # Keep inline.
]
`,
			marker: "# Keep inline.",
		},
		{
			name: "template directive",
			source: `aliases = [
  # %{ for alias in aliases ~}
  "${alias}",
  # %{ endfor ~}
]
`,
			marker: "# %{ endfor ~}",
		},
		{
			name: "ordinary trailing comment",
			source: `values = [
  "first",
  # Keep this inside the list.
]
`,
			marker: "# Keep this inside the list.",
		},
		{
			name: "empty list with comment",
			source: `values = [
  # No values configured.
]
`,
			marker: "# No values configured.",
		},
		{
			name: "nested list",
			source: `values = [
  [
    "first",
    # Inner list comment.
  ],
]
`,
			marker: "# Inner list comment.",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			formatted, err := Format([]byte(tt.source))
			if err != nil {
				t.Fatal(err)
			}

			output := string(formatted)
			comment := strings.Index(output, tt.marker)
			if comment < 0 {
				t.Fatalf("comment missing:\n%s", output)
			}
			openingBracket := strings.LastIndex(output[:comment], "[")
			closingBracket := -1
			if openingBracket >= 0 {
				if offset := strings.Index(output[openingBracket:], "]"); offset >= 0 {
					closingBracket = openingBracket + offset
				}
			}
			if comment < 0 || closingBracket < 0 || comment > closingBracket {
				t.Fatalf("comment must remain before the list's closing bracket:\n%s", output)
			}
			if strings.Count(output, tt.marker) != 1 {
				t.Fatalf("comment must appear exactly once:\n%s", output)
			}

			second, err := Format(formatted)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(formatted, second) {
				t.Fatalf("formatting must be idempotent:\nfirst:\n%s\nsecond:\n%s", formatted, second)
			}
		})
	}
}
