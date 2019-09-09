package hclsyntax

import (
	"testing"
)

func TestValidIdentifier(t *testing.T) {
	tests := []struct {
		Input string
		Want  bool
	}{
		{"", false},
		{"hello", true},
		{"hello.world", false},
		{"hello ", false},
		{" hello", false},
		{"hello\n", false},
		{"hello world", false},
		{"aws_instance", true},
		{"aws.instance", false},
		{"foo-bar", true},
		{"foo--bar", true},
		{"foo_", true},
		{"foo-", true},
		{"_foobar", true},
		{"-foobar", false},
		{"blah1", true},
		{"blah1blah", true},
		{"1blah1blah", false},
		{"héllo", true}, // combining acute accent
		{"Χαίρετε", true},
		{"звать", true},
		{"今日は", true},
		{"\x80", false},  // UTF-8 continuation without an introducer
		{"a\x80", false}, // UTF-8 continuation after a non-introducer
	}

	for _, test := range tests {
		t.Run(test.Input, func(t *testing.T) {
			got := ValidIdentifier(test.Input)
			if got != test.Want {
				t.Errorf("wrong result %#v; want %#v", got, test.Want)
			}
		})
	}
}
