package printer

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/hashicorp/hcl/hcl/parser"
	"github.com/hashicorp/hcl/testhelper"
)

var update = flag.Bool("update", false, "update golden files")

const (
	dataDir = "testdata"
)

type entry struct {
	source, golden string
	filters        []Filter
}

// Use go test -update to create/update the respective golden files.
var data = []entry{
	{"complexhcl.input", "complexhcl.golden", nil},
	{"list.input", "list.golden", nil},
	{"list_comment.input", "list_comment.golden", nil},
	{"comment.input", "comment.golden", nil},
	{"comment_crlf.input", "comment.golden", nil},
	{"comment_aligned.input", "comment_aligned.golden", nil},
	{"comment_array.input", "comment_array.golden", nil},
	{"comment_end_file.input", "comment_end_file.golden", nil},
	{"comment_multiline_indent.input", "comment_multiline_indent.golden", nil},
	{"comment_multiline_no_stanza.input", "comment_multiline_no_stanza.golden", nil},
	{"comment_multiline_stanza.input", "comment_multiline_stanza.golden", nil},
	{"comment_newline.input", "comment_newline.golden", nil},
	{"comment_object_multi.input", "comment_object_multi.golden", nil},
	{"comment_standalone.input", "comment_standalone.golden", nil},
	{"empty_block.input", "empty_block.golden", nil},
	{"list_of_objects.input", "list_of_objects.golden", nil},
	{"multiline_string.input", "multiline_string.golden", nil},
	{"object_singleline.input", "object_singleline.golden", nil},
	{"object_with_heredoc.input", "object_with_heredoc.golden", nil},
	{"object_with_heredoc.input", "object_with_heredoc.golden", nil},
	{"object_filter.input", "object_filter.golden", []Filter{&testhelper.TestFilter{}}},
}

func TestFiles(t *testing.T) {
	for _, e := range data {
		source := filepath.Join(dataDir, e.source)
		golden := filepath.Join(dataDir, e.golden)
		t.Run(e.source, func(t *testing.T) {
			check(t, source, golden, e.filters)
		})
	}
}

func check(t *testing.T, source, golden string, filters []Filter) {
	src, err := ioutil.ReadFile(source)
	if err != nil {
		t.Error(err)
		return
	}

	res, err := format(src, filters)
	if err != nil {
		t.Error(err)
		return
	}

	// update golden files if necessary
	if *update {
		if err := ioutil.WriteFile(golden, res, 0644); err != nil {
			t.Error(err)
		}
		return
	}

	// get golden
	gld, err := ioutil.ReadFile(golden)
	if err != nil {
		t.Error(err)
		return
	}

	// formatted source and golden must be the same
	if err := diff(source, golden, res, gld); err != nil {
		t.Error(err)
		return
	}
}

// diff compares a and b.
func diff(aname, bname string, a, b []byte) error {
	var buf bytes.Buffer // holding long error message

	// compare lengths
	if len(a) != len(b) {
		fmt.Fprintf(&buf, "\nlength changed: len(%s) = %d, len(%s) = %d", aname, len(a), bname, len(b))
	}

	// compare contents
	line := 1
	offs := 1
	for i := 0; i < len(a) && i < len(b); i++ {
		ch := a[i]
		if ch != b[i] {
			fmt.Fprintf(&buf, "\n%s:%d:%d: %q", aname, line, i-offs+1, lineAt(a, offs))
			fmt.Fprintf(&buf, "\n%s:%d:%d: %q", bname, line, i-offs+1, lineAt(b, offs))
			fmt.Fprintf(&buf, "\n\n")
			break
		}
		if ch == '\n' {
			line++
			offs = i + 1
		}
	}

	if buf.Len() > 0 {
		return errors.New(buf.String())
	}
	return nil
}

// format parses src, prints the corresponding AST, verifies the resulting
// src is syntactically correct, and returns the resulting src or an error
// if any.
func format(src []byte, filters []Filter) ([]byte, error) {
	formatted, err := Format(src, filters)
	if err != nil {
		return nil, err
	}

	// make sure formatted output is syntactically correct
	if _, err := parser.Parse(formatted); err != nil {
		return nil, fmt.Errorf("parse: %s\n%s", err, formatted)
	}

	return formatted, nil
}

// lineAt returns the line in text starting at offset offs.
func lineAt(text []byte, offs int) []byte {
	i := offs
	for i < len(text) && text[i] != '\n' {
		i++
	}
	return text[offs:i]
}
