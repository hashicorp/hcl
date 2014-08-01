package hcl

// MockVisitor is a visitor implementation that can be used for tests
// and simply records the nodes that it has visited.
type MockVisitor struct {
	Nodes []Node
}

func (v *MockVisitor) Visit(n Node) {
	v.Nodes = append(v.Nodes, n)
}
