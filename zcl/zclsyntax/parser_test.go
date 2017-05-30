package zclsyntax

import (
	"reflect"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/zclconf/go-zcl/zcl"
)

func TestParseConfig(t *testing.T) {
	tests := []struct {
		input     string
		diagCount int
		want      *Body
	}{
		{
			``,
			0,
			&Body{
				Attributes: Attributes{},
				Blocks:     Blocks{},
				SrcRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 1, Column: 1, Byte: 0},
				},
				EndRange: zcl.Range{
					Start: zcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   zcl.Pos{Line: 1, Column: 1, Byte: 0},
				},
			},
		},
	}

	prettyConfig := &pretty.Config{
		Diffable:          true,
		IncludeUnexported: true,
		PrintStringers:    true,
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			got, diags := ParseConfig([]byte(test.input), "", zcl.Pos{Byte: 0, Line: 1, Column: 1})
			if len(diags) != test.diagCount {
				t.Errorf("wrong number of diagnostics %d; want %d", len(diags), test.diagCount)
				for _, diag := range diags {
					t.Logf(" - %s", diag.Error())
				}
			}

			if !reflect.DeepEqual(got, test.want) {
				diff := prettyConfig.Compare(test.want, got)
				t.Errorf("wrong result\ninput: %s\ndiff:  %s", test.input, diff)
			}
		})
	}
}
