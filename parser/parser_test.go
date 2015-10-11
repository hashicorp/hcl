package parser

import (
	"fmt"
	"testing"
)

func TestAssignStatment(t *testing.T) {
	src := `ami = "${var.foo}"`

	p := New([]byte(src))
	p.enableTrace = true
	n := p.Parse()

	fmt.Println(n)

}
