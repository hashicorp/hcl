// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package hclsyntax

import (
	"reflect"
	"testing"

	"github.com/hashicorp/hcl/v2"
)

func TestParseSimpleBody(t *testing.T) {
	const src = `
resource "aws_vpc" "foo" {
  vpc_id         = local.id
  literal_string = "hello"
  literal_list   = [1, true, "x"]
  literal_object = {
    plain = "value"
  }
  config = {
    primary = aws_vpc.foo.id
    names   = [local.id, "fallback"]
  }
  not_literal = upper(local.id)

  nested {
    from_resource = aws_vpc.foo.id
  }
  nested {
    from_resource = aws_vpc.foo.id
  }
}

locals {
  id = aws_vpc.foo.id
}
`

	file, diags := ParseConfig([]byte(src), "", hcl.Pos{})
	if diags.HasErrors() {
		t.Fatalf("unexpected parse diagnostics:\n%s", diags.Error())
	}

	got, diags := ParseSimpleBody(file.Body)
	if diags.HasErrors() {
		t.Fatalf("unexpected analysis diagnostics:\n%s", diags.Error())
	}

	if got == nil {
		t.Fatal("got nil result")
	}

	if len(got.Blocks) != 2 {
		t.Fatalf("wrong number of root blocks %d; want 2", len(got.Blocks))
	}

	resourceBlock := got.Blocks[0]
	if resourceBlock.Type != "resource" {
		t.Fatalf("wrong root block type %q; want resource", resourceBlock.Type)
	}
	if !reflect.DeepEqual(resourceBlock.Labels, []string{"aws_vpc", "foo"}) {
		t.Fatalf("wrong resource labels %#v", resourceBlock.Labels)
	}

	assertTraversal(t, resourceBlock.Body.Attributes["vpc_id"], []string{"local", "id"})
	assertLiteral(t, resourceBlock.Body.Attributes["literal_string"])
	assertLiteral(t, resourceBlock.Body.Attributes["literal_list"])
	assertLiteral(t, resourceBlock.Body.Attributes["literal_object"])
	assertNeither(t, resourceBlock.Body.Attributes["not_literal"])

	if len(resourceBlock.Body.Blocks) != 2 {
		t.Fatalf("wrong number of nested resource blocks %d; want 2", len(resourceBlock.Body.Blocks))
	}

	nestedBlock := resourceBlock.Body.Blocks[0]
	if nestedBlock.Type != "nested" {
		t.Fatalf("wrong nested block type %q; want nested", nestedBlock.Type)
	}
	assertTraversal(t, nestedBlock.Body.Attributes["from_resource"], []string{"aws_vpc", "foo", "id"})

	localsBlock := got.Blocks[1]
	if localsBlock.Type != "locals" {
		t.Fatalf("wrong second root block type %q; want locals", localsBlock.Type)
	}
	assertTraversal(t, localsBlock.Body.Attributes["id"], []string{"aws_vpc", "foo", "id"})
}

func assertLiteral(t *testing.T, got SimpleAttribute) {
	t.Helper()

	if !got.IsLiteral() {
		t.Fatalf("expression %#v is not marked as literal", got.Expr)
	}
	if len(got.Traversal) != 0 {
		t.Fatalf("literal expression unexpectedly has traversal %#v", got.Traversal)
	}
	if got.Expr == nil {
		t.Fatal("literal expression has nil raw expression")
	}
}

func assertTraversal(t *testing.T, got SimpleAttribute, want []string) {
	t.Helper()

	if got.IsLiteral() {
		t.Fatalf("traversal expression unexpectedly marked literal")
	}
	if got.Expr == nil {
		t.Fatal("traversal expression has nil raw expression")
	}

	if len(got.Traversal) != len(want) {
		t.Fatalf("wrong traversal length %d; want %d", len(got.Traversal), len(want))
	}

	for i, step := range got.Traversal {
		switch step := step.(type) {
		case hcl.TraverseRoot:
			if got, want := step.Name, want[i]; got != want {
				t.Fatalf("wrong traversal step %d name %q; want %q", i, got, want)
			}
		case hcl.TraverseAttr:
			if got, want := step.Name, want[i]; got != want {
				t.Fatalf("wrong traversal step %d name %q; want %q", i, got, want)
			}
		default:
			t.Fatalf("unexpected traversal step type %T", step)
		}
	}
}

func assertNeither(t *testing.T, got SimpleAttribute) {
	t.Helper()

	if got.IsLiteral() || got.IsTraversal() {
		t.Fatal("expression unexpectedly marked literal or traversal")
	}
	if len(got.Traversal) != 0 {
		t.Fatalf("expression unexpectedly has traversal %#v", got.Traversal)
	}
	if got.Expr == nil {
		t.Fatal("expression has nil raw expression")
	}
}
