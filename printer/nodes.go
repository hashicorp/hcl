package printer

import (
	"bytes"
	"fmt"

	"github.com/fatih/hcl/ast"
)

const (
	blank   = byte(' ')
	newline = byte('\n')
	tab     = byte('\t')
)

// node
func (p *printer) output(n ast.Node) []byte {
	var buf bytes.Buffer

	switch t := n.(type) {
	case *ast.ObjectList:
		for i, item := range t.Items {
			buf.Write(p.objectItem(item))
			if i != len(t.Items)-1 {
				buf.Write([]byte{newline, newline})
			}
		}
	case *ast.ObjectKey:
		buf.WriteString(t.Token.Text)
	case *ast.ObjectItem:
		buf.Write(p.objectItem(t))
	case *ast.LiteralType:
		buf.WriteString(t.Token.Text)
	case *ast.ListType:
		buf.Write(p.list(t))
	case *ast.ObjectType:
		buf.Write(p.objectType(t))
	default:
		fmt.Printf(" unknown type: %T\n", n)
	}

	return buf.Bytes()
}

func (p *printer) objectItem(o *ast.ObjectItem) []byte {
	var buf bytes.Buffer

	for i, k := range o.Keys {
		buf.WriteString(k.Token.Text)
		buf.WriteByte(blank)

		// reach end of key
		if i == len(o.Keys)-1 && len(o.Keys) == 1 {
			buf.WriteString("=")
			buf.WriteByte(blank)
		}
	}

	buf.Write(p.output(o.Val))
	return buf.Bytes()
}

func (p *printer) literal(l *ast.LiteralType) []byte {
	return []byte(l.Token.Text)
}

func (p *printer) objectType(o *ast.ObjectType) []byte {
	var buf bytes.Buffer
	buf.WriteString("{")
	buf.WriteByte(newline)

	for _, item := range o.List.Items {
		buf.Write(p.indent(p.objectItem(item)))
		buf.WriteByte(newline)
	}

	buf.WriteString("}")
	return buf.Bytes()
}

// printList prints a HCL list
func (p *printer) list(l *ast.ListType) []byte {
	var buf bytes.Buffer
	buf.WriteString("[")

	for i, item := range l.List {
		if item.Pos().Line != l.Lbrack.Line {
			// multiline list, add newline before we add each item
			buf.WriteByte(newline)
			// also indent each line
			buf.Write(p.indent(p.output(item)))
		} else {
			buf.Write(p.output(item))
		}

		if i != len(l.List)-1 {
			buf.WriteString(",")
			buf.WriteByte(blank)
		} else if item.Pos().Line != l.Lbrack.Line {
			buf.WriteString(",")
			buf.WriteByte(newline)
		}
	}

	buf.WriteString("]")
	return buf.Bytes()
}

// indent indents the lines of the given buffer for each non-empty line
func (p *printer) indent(buf []byte) []byte {
	var prefix []byte
	if p.cfg.SpacesWidth != 0 {
		for i := 0; i < p.cfg.SpacesWidth; i++ {
			prefix = append(prefix, blank)
		}
	} else {
		prefix = []byte{tab}
	}

	var res []byte
	bol := true
	for _, c := range buf {
		if bol && c != '\n' {
			res = append(res, prefix...)
		}
		res = append(res, c)
		bol = c == '\n'
	}
	return res
}
