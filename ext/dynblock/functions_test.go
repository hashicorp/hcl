// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynblock

import (
	"reflect"
	"testing"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/zclconf/go-cty/cty"

	"github.com/davecgh/go-spew/spew"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// This is heavily based on ext/dynblock/variables_test.go

func TestFunctions(t *testing.T) {
	const src = `

# We have some references to things inside the "val" attribute inside each
# of our "b" blocks, which should be included in the result of WalkFunctions
# but not WalkExpandFunctions.

a {
  dynamic "b" {
    for_each = [for i, v in some_list_0: "${i}=${v},${baz}"]
    labels = [b_label_func_0("${b.value}")]
    content {
      val = "${b_val_func_0(b.value)}"
    }
  }
}

dynamic "a" {
  for_each = b_fe_func_1(some_list_1)

  content {
    b "foo" {
      val = b_val_func_1("${a.value}")
    }

    dynamic "b" {
      for_each = b_fe_func_2(some_list_2)
      iterator = dyn_b
      labels = [b_label_func_2("${a.value} ${dyn_b.value}")]
      content {
        val = b_val_func_2("${a.value} ${dyn_b.value}")
      }
    }
  }
}

dynamic "a" {
  for_each = b_fe_func_3(some_list_3)
  iterator = dyn_a

  content {
    b "foo" {
      val = b_val_func_3("${dyn_a.value}")
    }

    dynamic "b" {
      for_each = b_fe_func_4(some_list_4)
      labels = [b_label_func_4("${dyn_a.value} ${b.value}")]
      content {
        val = b_val_func_4("${dyn_a.value} ${b.value}")
      }
    }
  }
}
`

	f, diags := hclsyntax.ParseConfig([]byte(src), "", hcl.Pos{})
	if len(diags) != 0 {
		t.Errorf("unexpected diagnostics during parse")
		for _, diag := range diags {
			t.Logf("- %s", diag)
		}
		return
	}

	spec := &hcldec.BlockListSpec{
		TypeName: "a",
		Nested: &hcldec.BlockMapSpec{
			TypeName:   "b",
			LabelNames: []string{"key"},
			Nested: &hcldec.AttrSpec{
				Name: "val",
				Type: cty.String,
			},
		},
	}

	t.Run("WalkFunctions", func(t *testing.T) {
		traversals := FunctionsHCLDec(f.Body, spec)
		got := make([]string, len(traversals))
		for i, traversal := range traversals {
			got[i] = traversal.RootName()
		}

		// The block structure is traversed one level at a time, so the ordering
		// here is reflecting first a pass of the root, then the first child
		// under the root, then the first child under that, etc.
		want := []string{
			"b_fe_func_1",
			"b_fe_func_3",
			"b_label_func_0",
			"b_val_func_0",
			"b_fe_func_2",
			"b_label_func_2",
			"b_val_func_1",
			"b_val_func_2",
			"b_fe_func_4",
			"b_label_func_4",
			"b_val_func_3",
			"b_val_func_4",
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("wrong result\ngot: %swant: %s", spew.Sdump(got), spew.Sdump(want))
		}
	})

	t.Run("WalkExpandFunctions", func(t *testing.T) {
		traversals := ExpandFunctionsHCLDec(f.Body, spec)
		got := make([]string, len(traversals))
		for i, traversal := range traversals {
			got[i] = traversal.RootName()
		}

		// The block structure is traversed one level at a time, so the ordering
		// here is reflecting first a pass of the root, then the first child
		// under the root, then the first child under that, etc.
		want := []string{
			"b_fe_func_1",
			"b_fe_func_3",
			"b_label_func_0",
			"b_fe_func_2",
			"b_label_func_2",
			"b_fe_func_4",
			"b_label_func_4",
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("wrong result\ngot: %swant: %s", spew.Sdump(got), spew.Sdump(want))
		}
	})
}
