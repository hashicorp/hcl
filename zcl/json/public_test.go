package json

import (
	"testing"

	"github.com/apparentlymart/go-cty/cty"
	"github.com/apparentlymart/go-zcl/zcl"
)

func TestParse_nonObject(t *testing.T) {
	src := `true`
	file, diags := Parse([]byte(src), "")
	if len(diags) != 1 {
		t.Errorf("got %d diagnostics; want 1", len(diags))
	}
	if file == nil {
		t.Errorf("got nil File; want actual file")
	}
	if file.Body == nil {
		t.Fatalf("got nil Body; want actual body")
	}
	if file.Body.(*body).obj == nil {
		t.Errorf("got nil Body object; want placeholder object")
	}
}

func TestParseTemplateWithHIL(t *testing.T) {
	src := `{"greeting": "hello ${\"world\"}"}`
	file, diags := ParseWithHIL([]byte(src), "")
	if len(diags) != 0 {
		t.Errorf("got %d diagnostics on parse; want 0", len(diags))
	}
	if file == nil {
		t.Errorf("got nil File; want actual file")
	}
	if file.Body == nil {
		t.Fatalf("got nil Body; want actual body")
	}
	attrs, diags := file.Body.JustAttributes()
	if len(diags) != 0 {
		t.Errorf("got %d diagnostics on decode; want 0", len(diags))
	}

	val, diags := attrs["greeting"].Expr.Value(&zcl.EvalContext{})
	if len(diags) != 0 {
		t.Errorf("got %d diagnostics on eval; want 0", len(diags))
	}

	if !val.RawEquals(cty.StringVal("hello world")) {
		t.Errorf("wrong result %#v; want %#v", val, cty.StringVal("hello world"))
	}
}
