// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package fuzzhclwrite

import (
	"io"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

func FuzzParseConfig(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		file, diags := hclwrite.ParseConfig(data, "<fuzz-conf>", hcl.Pos{Line: 1, Column: 1})

		if diags.HasErrors() {
			t.Logf("Error when parsing JSON %v", data)
			for _, diag := range diags {
				t.Logf("- %s", diag.Error())
			}
			return
		}

		_, err := file.WriteTo(io.Discard)

		if err != nil {
			t.Fatalf("error writing to file: %s", err)
		}
	})
}
