// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hclsyntax

import "testing"

func TestNameSuggestion(t *testing.T) {
	var keywords = []string{"false", "true", "null"}

	tests := []struct {
		Input, Want string
	}{
		{"true", "true"},
		{"false", "false"},
		{"null", "null"},
		{"bananas", ""},
		{"NaN", ""},
		{"Inf", ""},
		{"Infinity", ""},
		{"void", ""},
		{"undefined", ""},

		{"ture", "true"},
		{"tru", "true"},
		{"tre", "true"},
		{"treu", "true"},
		{"rtue", "true"},

		{"flase", "false"},
		{"fales", "false"},
		{"flse", "false"},
		{"fasle", "false"},
		{"fasel", "false"},
		{"flue", "false"},

		{"nil", "null"},
		{"nul", "null"},
		{"unll", "null"},
		{"nll", "null"},
	}

	for _, test := range tests {
		t.Run(test.Input, func(t *testing.T) {
			got := nameSuggestion(test.Input, keywords)
			if got != test.Want {
				t.Errorf(
					"wrong result\ninput: %q\ngot:   %q\nwant:  %q",
					test.Input, got, test.Want,
				)
			}
		})
	}
}
