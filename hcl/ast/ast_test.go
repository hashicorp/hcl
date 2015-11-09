package ast

import (
	"reflect"
	"testing"

	"github.com/hashicorp/hcl/hcl/token"
)

func TestObjectListFilter(t *testing.T) {
	var cases = []struct {
		Filter []string
		Input  []*ObjectItem
		Output []*ObjectItem
	}{
		{
			[]string{"foo"},
			[]*ObjectItem{
				&ObjectItem{
					Keys: []*ObjectKey{
						&ObjectKey{
							Token: token.Token{Type: token.STRING, Text: `"foo"`},
						},
					},
				},
			},
			[]*ObjectItem{
				&ObjectItem{
					Keys: []*ObjectKey{},
				},
			},
		},

		{
			[]string{"foo"},
			[]*ObjectItem{
				&ObjectItem{
					Keys: []*ObjectKey{
						&ObjectKey{Token: token.Token{Type: token.STRING, Text: `"foo"`}},
						&ObjectKey{Token: token.Token{Type: token.STRING, Text: `"bar"`}},
					},
				},
				&ObjectItem{
					Keys: []*ObjectKey{
						&ObjectKey{Token: token.Token{Type: token.STRING, Text: `"baz"`}},
					},
				},
			},
			[]*ObjectItem{
				&ObjectItem{
					Keys: []*ObjectKey{
						&ObjectKey{Token: token.Token{Type: token.STRING, Text: `"bar"`}},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		input := &ObjectList{Items: tc.Input}
		expected := &ObjectList{Items: tc.Output}
		if actual := input.Filter(tc.Filter...); !reflect.DeepEqual(actual, expected) {
			t.Fatalf("in order: input, expected, actual\n\n%#v\n\n%#v\n\n%#v", input, expected, actual)
		}
	}
}
