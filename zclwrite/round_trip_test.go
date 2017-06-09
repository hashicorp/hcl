package zclwrite

import (
	"bytes"
	"testing"

	"github.com/zclconf/go-zcl/zcl"
)

func TestRoundTrip(t *testing.T) {
	tests := []string{
		``,
		`foo = 1
`,
		`
foobar = 1
baz    = 1
`,
		`
# this file is awesome

# tossed salads and scrambled eggs
foobar = 1
baz    = 1

# and they all lived happily ever after
`,
	}

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			src := []byte(test)
			file, diags := parse(src, "", zcl.Pos{Line: 1, Column: 1})
			if len(diags) != 0 {
				for _, diag := range diags {
					t.Logf(" - %s", diag.Error())
				}
				t.Fatalf("unexpected diagnostics")
			}

			wr := &bytes.Buffer{}
			n, err := file.WriteTo(wr)
			if n != len(test) {
				t.Errorf("wrong number of bytes %d; want %d", n, len(test))
			}
			if err != nil {
				t.Fatalf("error from WriteTo")
			}

			result := wr.Bytes()

			if !bytes.Equal(result, src) {
				t.Errorf("wrong result\nresult:\n%s\ninput:\n%s", result, src)
			}
		})
	}
}
