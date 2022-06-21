package fuzzjson

import (
	"testing"

	"github.com/hashicorp/hcl/v2/json"
)

func FuzzParse(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		_, diags := json.Parse(data, "<fuzz-conf>")

		if diags.HasErrors() {
			t.Logf("Error when parsing JSON %v", data)
			for _, diag := range diags {
				t.Logf("- %s", diag.Error())
			}
		}
	})
}
