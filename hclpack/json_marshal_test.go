package hclpack

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/hcl2/hcl"
)

func TestJSONRoundTrip(t *testing.T) {
	src := `
	service "example" {
	  priority = 2
	  platform {
	    os   = "linux"
	    arch = "amd64"
	  }
	  process "web" {
	    exec = ["./webapp"]
	  }
	  process "worker" {
	    exec = ["./worker"]
	  }
	}
	`

	startBody, diags := PackNativeFile([]byte(src), "example.svc", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatalf("Failed to parse: %s", diags.Error())
	}

	jb, err := startBody.MarshalJSON()
	if err != nil {
		t.Fatalf("Failed to marshal: %s", err)
	}

	endBody := &Body{}
	err = endBody.UnmarshalJSON(jb)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %s", err)
	}

	if !cmp.Equal(startBody, endBody) {
		t.Errorf("incorrect result\n%s", cmp.Diff(startBody, endBody))
	}
}
