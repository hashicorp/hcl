package printer

import (
	"bytes"
	"fmt"

	"github.com/fatih/hcl/ast"
)

func (p *printer) printNode(n ast.Node) []byte {
	var buf bytes.Buffer

	switch t := n.(type) {
	case *ast.ObjectList:
		fmt.Println("printing objectList", t)
		for _, item := range t.Items {
			buf.Write(p.printObjectItem(item))
		}
	case *ast.ObjectKey:
		fmt.Println("printing objectKey", t)
	case *ast.ObjectItem:
		fmt.Println("printing objectItem", t)
		buf.Write(p.printObjectItem(t))
	case *ast.LiteralType:
		buf.Write(p.printLiteral(t))
	case *ast.ListType:
		buf.Write(p.printList(t))
	case *ast.ObjectType:
		fmt.Println("printing ObjectType", t)
	default:
		fmt.Printf(" unknown type: %T\n", n)
	}

	return buf.Bytes()
}

func (p *printer) printObjectItem(o *ast.ObjectItem) []byte {
	var buf bytes.Buffer

	for i, k := range o.Keys {
		buf.WriteString(k.Token.Text)
		if i != len(o.Keys)-1 || len(o.Keys) == 1 {
			buf.WriteString(" ")
		}

		// reach end of key
		if i == len(o.Keys)-1 {
			buf.WriteString("=")
			buf.WriteString(" ")
		}
	}

	buf.Write(p.printNode(o.Val))
	return buf.Bytes()
}

func (p *printer) printLiteral(l *ast.LiteralType) []byte {
	return []byte(l.Token.Text)
}

func (p *printer) printList(l *ast.ListType) []byte {
	var buf bytes.Buffer
	buf.WriteString("[")

	for i, item := range l.List {
		if item.Pos().Line != l.Lbrack.Line {
			// not same line
			buf.WriteString("\n")
		}

		buf.Write(p.printNode(item))

		if i != len(l.List)-1 {
			buf.WriteString(",")
			buf.WriteString(" ")
		} else if item.Pos().Line != l.Lbrack.Line {
			buf.WriteString(",")
			buf.WriteString("\n")
		}
	}

	buf.WriteString("]")
	return buf.Bytes()
}
