package json

import (
	"testing"
)

func TestParse_nonObject(t *testing.T) {
	src := `true`
	file, diags := Parse([]byte(src), "")
	if len(diags) != 1 {
		t.Errorf("got %d diagnostics; want 1", len(diags))
	}
	if file != nil {
		t.Errorf("got non-nil File; want nil")
	}
}
