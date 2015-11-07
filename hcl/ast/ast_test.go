package ast

import (
	"reflect"
	"testing"

	"github.com/hashicorp/hcl/hcl/token"
)

func TestObjectListPrefix(t *testing.T) {
	var cases = []struct {
		Prefix []string
		Input  []*ObjectItem
		Output []*ObjectItem
	}{
		{
			[]string{"foo"},
			[]*ObjectItem{
				&ObjectItem{
					Keys: []*ObjectKey{
						&ObjectKey{
							Token: token.Token{Type: token.STRING, Text: `foo`},
						},
					},
				},
			},
			nil,
		},

		{
			[]string{"foo"},
			[]*ObjectItem{
				&ObjectItem{
					Keys: []*ObjectKey{
						&ObjectKey{Token: token.Token{Type: token.STRING, Text: `foo`}},
						&ObjectKey{Token: token.Token{Type: token.STRING, Text: `bar`}},
					},
				},
				&ObjectItem{
					Keys: []*ObjectKey{
						&ObjectKey{Token: token.Token{Type: token.STRING, Text: `baz`}},
					},
				},
			},
			[]*ObjectItem{
				&ObjectItem{
					Keys: []*ObjectKey{
						&ObjectKey{Token: token.Token{Type: token.STRING, Text: `bar`}},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		input := &ObjectList{Items: tc.Input}
		expected := &ObjectList{Items: tc.Output}
		if actual := input.Prefix(tc.Prefix...); !reflect.DeepEqual(actual, expected) {
			t.Fatalf("in order: input, expected, actual\n\n%#v\n\n%#v\n\n%#v", input, expected, actual)
		}
	}
}
