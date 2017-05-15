package json

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/apparentlymart/go-zcl/zcl"
)

func TestScan(t *testing.T) {
	tests := []struct {
		Input string
		Want  []token
	}{
		{
			``,
			[]token{
				{
					Type: tokenEOF,
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   0,
							Line:   1,
							Column: 1,
						},
						End: zcl.Pos{
							Byte:   0,
							Line:   1,
							Column: 1,
						},
					},
				},
			},
		},
		{
			`   `,
			[]token{
				{
					Type: tokenEOF,
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   3,
							Line:   1,
							Column: 4,
						},
						End: zcl.Pos{
							Byte:   3,
							Line:   1,
							Column: 4,
						},
					},
				},
			},
		},
		{
			`{}`,
			[]token{
				{
					Type:  tokenBraceO,
					Bytes: []byte(`{`),
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   0,
							Line:   1,
							Column: 1,
						},
						End: zcl.Pos{
							Byte:   1,
							Line:   1,
							Column: 2,
						},
					},
				},
				{
					Type:  tokenBraceC,
					Bytes: []byte(`}`),
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   1,
							Line:   1,
							Column: 2,
						},
						End: zcl.Pos{
							Byte:   2,
							Line:   1,
							Column: 3,
						},
					},
				},
				{
					Type: tokenEOF,
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   2,
							Line:   1,
							Column: 3,
						},
						End: zcl.Pos{
							Byte:   2,
							Line:   1,
							Column: 3,
						},
					},
				},
			},
		},
		{
			`][`,
			[]token{
				{
					Type:  tokenBrackC,
					Bytes: []byte(`]`),
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   0,
							Line:   1,
							Column: 1,
						},
						End: zcl.Pos{
							Byte:   1,
							Line:   1,
							Column: 2,
						},
					},
				},
				{
					Type:  tokenBrackO,
					Bytes: []byte(`[`),
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   1,
							Line:   1,
							Column: 2,
						},
						End: zcl.Pos{
							Byte:   2,
							Line:   1,
							Column: 3,
						},
					},
				},
				{
					Type: tokenEOF,
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   2,
							Line:   1,
							Column: 3,
						},
						End: zcl.Pos{
							Byte:   2,
							Line:   1,
							Column: 3,
						},
					},
				},
			},
		},
		{
			`:,`,
			[]token{
				{
					Type:  tokenColon,
					Bytes: []byte(`:`),
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   0,
							Line:   1,
							Column: 1,
						},
						End: zcl.Pos{
							Byte:   1,
							Line:   1,
							Column: 2,
						},
					},
				},
				{
					Type:  tokenComma,
					Bytes: []byte(`,`),
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   1,
							Line:   1,
							Column: 2,
						},
						End: zcl.Pos{
							Byte:   2,
							Line:   1,
							Column: 3,
						},
					},
				},
				{
					Type: tokenEOF,
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   2,
							Line:   1,
							Column: 3,
						},
						End: zcl.Pos{
							Byte:   2,
							Line:   1,
							Column: 3,
						},
					},
				},
			},
		},
		{
			`1`,
			[]token{
				{
					Type:  tokenNumber,
					Bytes: []byte(`1`),
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   0,
							Line:   1,
							Column: 1,
						},
						End: zcl.Pos{
							Byte:   1,
							Line:   1,
							Column: 2,
						},
					},
				},
				{
					Type: tokenEOF,
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   1,
							Line:   1,
							Column: 2,
						},
						End: zcl.Pos{
							Byte:   1,
							Line:   1,
							Column: 2,
						},
					},
				},
			},
		},
		{
			`  1`,
			[]token{
				{
					Type:  tokenNumber,
					Bytes: []byte(`1`),
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   2,
							Line:   1,
							Column: 3,
						},
						End: zcl.Pos{
							Byte:   3,
							Line:   1,
							Column: 4,
						},
					},
				},
				{
					Type: tokenEOF,
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   3,
							Line:   1,
							Column: 4,
						},
						End: zcl.Pos{
							Byte:   3,
							Line:   1,
							Column: 4,
						},
					},
				},
			},
		},
		{
			`  12`,
			[]token{
				{
					Type:  tokenNumber,
					Bytes: []byte(`12`),
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   2,
							Line:   1,
							Column: 3,
						},
						End: zcl.Pos{
							Byte:   4,
							Line:   1,
							Column: 5,
						},
					},
				},
				{
					Type: tokenEOF,
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   4,
							Line:   1,
							Column: 5,
						},
						End: zcl.Pos{
							Byte:   4,
							Line:   1,
							Column: 5,
						},
					},
				},
			},
		},
		{
			`1 2`,
			[]token{
				{
					Type:  tokenNumber,
					Bytes: []byte(`1`),
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   0,
							Line:   1,
							Column: 1,
						},
						End: zcl.Pos{
							Byte:   1,
							Line:   1,
							Column: 2,
						},
					},
				},
				{
					Type:  tokenNumber,
					Bytes: []byte(`2`),
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   2,
							Line:   1,
							Column: 3,
						},
						End: zcl.Pos{
							Byte:   3,
							Line:   1,
							Column: 4,
						},
					},
				},
				{
					Type: tokenEOF,
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   3,
							Line:   1,
							Column: 4,
						},
						End: zcl.Pos{
							Byte:   3,
							Line:   1,
							Column: 4,
						},
					},
				},
			},
		},
		{
			"\n1\n 2",
			[]token{
				{
					Type:  tokenNumber,
					Bytes: []byte(`1`),
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   1,
							Line:   2,
							Column: 1,
						},
						End: zcl.Pos{
							Byte:   2,
							Line:   2,
							Column: 2,
						},
					},
				},
				{
					Type:  tokenNumber,
					Bytes: []byte(`2`),
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   4,
							Line:   3,
							Column: 2,
						},
						End: zcl.Pos{
							Byte:   5,
							Line:   3,
							Column: 3,
						},
					},
				},
				{
					Type: tokenEOF,
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   5,
							Line:   3,
							Column: 3,
						},
						End: zcl.Pos{
							Byte:   5,
							Line:   3,
							Column: 3,
						},
					},
				},
			},
		},
		{
			`-1 2.5`,
			[]token{
				{
					Type:  tokenNumber,
					Bytes: []byte(`-1`),
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   0,
							Line:   1,
							Column: 1,
						},
						End: zcl.Pos{
							Byte:   2,
							Line:   1,
							Column: 3,
						},
					},
				},
				{
					Type:  tokenNumber,
					Bytes: []byte(`2.5`),
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   3,
							Line:   1,
							Column: 4,
						},
						End: zcl.Pos{
							Byte:   6,
							Line:   1,
							Column: 7,
						},
					},
				},
				{
					Type: tokenEOF,
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   6,
							Line:   1,
							Column: 7,
						},
						End: zcl.Pos{
							Byte:   6,
							Line:   1,
							Column: 7,
						},
					},
				},
			},
		},
		{
			`true`,
			[]token{
				{
					Type:  tokenKeyword,
					Bytes: []byte(`true`),
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   0,
							Line:   1,
							Column: 1,
						},
						End: zcl.Pos{
							Byte:   4,
							Line:   1,
							Column: 5,
						},
					},
				},
				{
					Type: tokenEOF,
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   4,
							Line:   1,
							Column: 5,
						},
						End: zcl.Pos{
							Byte:   4,
							Line:   1,
							Column: 5,
						},
					},
				},
			},
		},
		{
			`""`,
			[]token{
				{
					Type:  tokenString,
					Bytes: []byte(`""`),
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   0,
							Line:   1,
							Column: 1,
						},
						End: zcl.Pos{
							Byte:   2,
							Line:   1,
							Column: 3,
						},
					},
				},
				{
					Type: tokenEOF,
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   2,
							Line:   1,
							Column: 3,
						},
						End: zcl.Pos{
							Byte:   2,
							Line:   1,
							Column: 3,
						},
					},
				},
			},
		},
		{
			`"hello"`,
			[]token{
				{
					Type:  tokenString,
					Bytes: []byte(`"hello"`),
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   0,
							Line:   1,
							Column: 1,
						},
						End: zcl.Pos{
							Byte:   7,
							Line:   1,
							Column: 8,
						},
					},
				},
				{
					Type: tokenEOF,
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   7,
							Line:   1,
							Column: 8,
						},
						End: zcl.Pos{
							Byte:   7,
							Line:   1,
							Column: 8,
						},
					},
				},
			},
		},
		{
			`"he\"llo"`,
			[]token{
				{
					Type:  tokenString,
					Bytes: []byte(`"he\"llo"`),
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   0,
							Line:   1,
							Column: 1,
						},
						End: zcl.Pos{
							Byte:   9,
							Line:   1,
							Column: 10,
						},
					},
				},
				{
					Type: tokenEOF,
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   9,
							Line:   1,
							Column: 10,
						},
						End: zcl.Pos{
							Byte:   9,
							Line:   1,
							Column: 10,
						},
					},
				},
			},
		},
		{
			`"hello\\" 1`,
			[]token{
				{
					Type:  tokenString,
					Bytes: []byte(`"hello\\"`),
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   0,
							Line:   1,
							Column: 1,
						},
						End: zcl.Pos{
							Byte:   9,
							Line:   1,
							Column: 10,
						},
					},
				},
				{
					Type:  tokenNumber,
					Bytes: []byte(`1`),
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   10,
							Line:   1,
							Column: 11,
						},
						End: zcl.Pos{
							Byte:   11,
							Line:   1,
							Column: 12,
						},
					},
				},
				{
					Type: tokenEOF,
					Range: zcl.Range{
						Start: zcl.Pos{
							Byte:   11,
							Line:   1,
							Column: 12,
						},
						End: zcl.Pos{
							Byte:   11,
							Line:   1,
							Column: 12,
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Input, func(t *testing.T) {
			buf := []byte(test.Input)
			start := pos{
				Filename: "",
				Pos: zcl.Pos{
					Byte:   0,
					Line:   1,
					Column: 1,
				},
			}
			got := scan(buf, start)

			if !reflect.DeepEqual(got, test.Want) {
				errMsg := &bytes.Buffer{}
				errMsg.WriteString("wrong result\ngot:\n")
				if len(got) == 0 {
					errMsg.WriteString("  (empty slice)\n")
				}
				for _, tok := range got {
					fmt.Fprintf(errMsg, "  - %#v\n", tok)
				}
				errMsg.WriteString("want:\n")
				if len(test.Want) == 0 {
					errMsg.WriteString("  (empty slice)\n")
				}
				for _, tok := range test.Want {
					fmt.Fprintf(errMsg, "  - %#v\n", tok)
				}
				t.Error(errMsg.String())
			}
		})
	}
}
