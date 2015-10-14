package parser

import "testing"

func TestAssignStatement(t *testing.T) {
	src := `ami = "${var.foo}"`
	p := New([]byte(src))
	p.enableTrace = true
	n, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	if n.Pos().Line != 1 {
		t.Errorf("AssignStatement position is wrong\n\twant: '%d'\n\tgot : '%d'", 1, n.Pos().Line)
	}
}
