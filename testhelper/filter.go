package testhelper

import "github.com/hashicorp/hcl/hcl/ast"

// TestFilter is a test filter for the printer.
type TestFilter struct{}

// Filter implements printer.Filter for TestFilter.
func (f *TestFilter) Filter(n *ast.File) error {
	n.Node.(*ast.ObjectList).Items[0].Val.(*ast.ObjectType).List.Items[0].Val.(*ast.LiteralType).Token.Text = "\"two\""
	return nil
}
