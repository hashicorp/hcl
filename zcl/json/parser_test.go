package json

import (
	"reflect"
	"testing"

	"github.com/apparentlymart/go-zcl/zcl"
	"github.com/davecgh/go-spew/spew"
)

func TestParse(t *testing.T) {
	tests := []struct {
		Input     string
		Want      node
		DiagCount int
	}{
		{
			`true`,
			&booleanVal{
				Value: true,
				SrcRange: zcl.Range{
					Filename: "",
					Start:    zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:      zcl.Pos{Line: 1, Column: 5, Byte: 4},
				},
			},
			0,
		},
		{
			`false`,
			&booleanVal{
				Value: false,
				SrcRange: zcl.Range{
					Filename: "",
					Start:    zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:      zcl.Pos{Line: 1, Column: 6, Byte: 5},
				},
			},
			0,
		},
		{
			`null`,
			&nullVal{
				SrcRange: zcl.Range{
					Filename: "",
					Start:    zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:      zcl.Pos{Line: 1, Column: 5, Byte: 4},
				},
			},
			0,
		},
		{
			`undefined`,
			nil,
			1,
		},
		{
			`flase`,
			nil,
			1,
		},
	}

	for _, test := range tests {
		t.Run(test.Input, func(t *testing.T) {
			got, diag := parseFileContent([]byte(test.Input), "")

			if len(diag) != test.DiagCount {
				t.Errorf("got %d diagnostics; want %d\n%s", len(diag), test.DiagCount, spew.Sdump(diag))
			}

			if !reflect.DeepEqual(got, test.Want) {
				t.Errorf(
					"wrong result\ninput: %s\ngot:   %#v\nwant:  %#v",
					test.Input, got, test.Want,
				)
			}
		})
	}
}
