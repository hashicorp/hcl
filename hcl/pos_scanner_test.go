package hcl

import (
	"bufio"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestPosScanner(t *testing.T) {
	tests := map[string]struct {
		Input    string
		Want     []Range
		WantToks [][]byte
	}{
		"empty": {
			"",
			[]Range{},
			[][]byte{},
		},
		"single line": {
			"hello",
			[]Range{
				{
					Start: Pos{Byte: 0, Line: 1, Column: 1},
					End:   Pos{Byte: 5, Line: 1, Column: 6},
				},
			},
			[][]byte{
				[]byte("hello"),
			},
		},
		"single line with trailing UNIX newline": {
			"hello\n",
			[]Range{
				{
					Start: Pos{Byte: 0, Line: 1, Column: 1},
					End:   Pos{Byte: 5, Line: 1, Column: 6},
				},
			},
			[][]byte{
				[]byte("hello"),
			},
		},
		"single line with trailing Windows newline": {
			"hello\r\n",
			[]Range{
				{
					Start: Pos{Byte: 0, Line: 1, Column: 1},
					End:   Pos{Byte: 5, Line: 1, Column: 6},
				},
			},
			[][]byte{
				[]byte("hello"),
			},
		},
		"two lines with UNIX newline": {
			"hello\nworld",
			[]Range{
				{
					Start: Pos{Byte: 0, Line: 1, Column: 1},
					End:   Pos{Byte: 5, Line: 1, Column: 6},
				},
				{
					Start: Pos{Byte: 6, Line: 2, Column: 1},
					End:   Pos{Byte: 11, Line: 2, Column: 6},
				},
			},
			[][]byte{
				[]byte("hello"),
				[]byte("world"),
			},
		},
		"two lines with Windows newline": {
			"hello\r\nworld",
			[]Range{
				{
					Start: Pos{Byte: 0, Line: 1, Column: 1},
					End:   Pos{Byte: 5, Line: 1, Column: 6},
				},
				{
					Start: Pos{Byte: 7, Line: 2, Column: 1},
					End:   Pos{Byte: 12, Line: 2, Column: 6},
				},
			},
			[][]byte{
				[]byte("hello"),
				[]byte("world"),
			},
		},
		"blank line with UNIX newlines": {
			"hello\n\nworld",
			[]Range{
				{
					Start: Pos{Byte: 0, Line: 1, Column: 1},
					End:   Pos{Byte: 5, Line: 1, Column: 6},
				},
				{
					Start: Pos{Byte: 6, Line: 2, Column: 1},
					End:   Pos{Byte: 6, Line: 2, Column: 1},
				},
				{
					Start: Pos{Byte: 7, Line: 3, Column: 1},
					End:   Pos{Byte: 12, Line: 3, Column: 6},
				},
			},
			[][]byte{
				[]byte("hello"),
				[]byte(""),
				[]byte("world"),
			},
		},
		"blank line with Windows newlines": {
			"hello\r\n\r\nworld",
			[]Range{
				{
					Start: Pos{Byte: 0, Line: 1, Column: 1},
					End:   Pos{Byte: 5, Line: 1, Column: 6},
				},
				{
					Start: Pos{Byte: 7, Line: 2, Column: 1},
					End:   Pos{Byte: 7, Line: 2, Column: 1},
				},
				{
					Start: Pos{Byte: 9, Line: 3, Column: 1},
					End:   Pos{Byte: 14, Line: 3, Column: 6},
				},
			},
			[][]byte{
				[]byte("hello"),
				[]byte(""),
				[]byte("world"),
			},
		},
		"two lines with combiner and UNIX newline": {
			"foo \U0001f469\U0001f3ff bar\nbaz",
			[]Range{
				{
					Start: Pos{Byte: 0, Line: 1, Column: 1},
					End:   Pos{Byte: 16, Line: 1, Column: 10},
				},
				{
					Start: Pos{Byte: 17, Line: 2, Column: 1},
					End:   Pos{Byte: 20, Line: 2, Column: 4},
				},
			},
			[][]byte{
				[]byte("foo \U0001f469\U0001f3ff bar"),
				[]byte("baz"),
			},
		},
		"two lines with combiner and Windows newline": {
			"foo \U0001f469\U0001f3ff bar\r\nbaz",
			[]Range{
				{
					Start: Pos{Byte: 0, Line: 1, Column: 1},
					End:   Pos{Byte: 16, Line: 1, Column: 10},
				},
				{
					Start: Pos{Byte: 18, Line: 2, Column: 1},
					End:   Pos{Byte: 21, Line: 2, Column: 4},
				},
			},
			[][]byte{
				[]byte("foo \U0001f469\U0001f3ff bar"),
				[]byte("baz"),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			src := []byte(test.Input)
			sc := NewRangeScanner(src, "", bufio.ScanLines)
			got := make([]Range, 0)
			gotToks := make([][]byte, 0)
			for sc.Scan() {
				got = append(got, sc.Range())
				gotToks = append(gotToks, sc.Bytes())
			}
			if sc.Err() != nil {
				t.Fatalf("unexpected error: %s", sc.Err())
			}
			if !reflect.DeepEqual(got, test.Want) {
				t.Errorf("incorrect ranges\ngot: %swant: %s", spew.Sdump(got), spew.Sdump(test.Want))
			}
			if !reflect.DeepEqual(gotToks, test.WantToks) {
				t.Errorf("incorrect tokens\ngot: %swant: %s", spew.Sdump(gotToks), spew.Sdump(test.WantToks))
			}
		})
	}
}
