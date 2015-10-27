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

// output prints creates a printable HCL output and returns it.
func (p *printer) output(n interface{}) []byte {
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

	if o.LeadComment != nil {
		for _, comment := range o.LeadComment.List {
			buf.WriteString(comment.Text)
			buf.WriteByte(newline)
		}
	}

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

	if o.Val.Pos().Line == o.Keys[0].Pos().Line && o.LineComment != nil {
		buf.WriteByte(blank)
		for _, comment := range o.LineComment.List {
			buf.WriteString(comment.Text)
		}
	}

	return buf.Bytes()
}

func (p *printer) alignedItems(items []*ast.ObjectItem) []byte {
	var buf bytes.Buffer

	var longestLine int
	for _, item := range items {
		lineLen := len(item.Keys[0].Token.Text) + len(p.output(item.Val))
		if lineLen > longestLine {
			longestLine = lineLen
		}
	}

	for _, item := range items {
		curLen := 0
		for i, k := range item.Keys {
			buf.WriteString(k.Token.Text)
			buf.WriteByte(blank)

			// reach end of key
			if i == len(item.Keys)-1 && len(item.Keys) == 1 {
				buf.WriteString("=")
				buf.WriteByte(blank)
			}

			curLen = len(k.Token.Text) // two blanks and one assign
		}
		val := p.output(item.Val)
		buf.Write(val)
		curLen += len(val)

		if item.Val.Pos().Line == item.Keys[0].Pos().Line && item.LineComment != nil {
			for i := 0; i < longestLine-curLen+1; i++ {
				buf.WriteByte(blank)
			}

			for _, comment := range item.LineComment.List {
				buf.WriteString(comment.Text)
			}
		}

		buf.WriteByte(newline)
	}

	return buf.Bytes()
}

func (p *printer) literal(l *ast.LiteralType) []byte {
	return []byte(l.Token.Text)
}

func (p *printer) objectType(o *ast.ObjectType) []byte {
	var buf bytes.Buffer
	buf.WriteString("{")
	buf.WriteByte(newline)

	// check if we have adjacent one liner items. If yes we'll going to align
	// the comments.
	var index int
	for {
		var oneLines []*ast.ObjectItem
		for _, item := range o.List.Items[index:] {
			// protect agains slice bounds
			if index == len(o.List.Items)-1 {
				// check for the latest item of a series of one liners in the
				// end of a list.
				if index != 0 && // do not check if the list length is one
					lines(string(p.objectItem(item))) < 1 && // be sure it's really a one line
					o.List.Items[index-1].Pos().Line == item.Pos().Line-1 {

					oneLines = append(oneLines, item)
					index++
				}
				break
			}

			if o.List.Items[1+index].Pos().Line == item.Pos().Line+1 {
				oneLines = append(oneLines, item)
				index++
			} else {
				// break in any nonadjacent items
				break
			}
		}

		// fmt.Printf("len(oneLines) = %+v\n", len(oneLines))
		// for _, i := range oneLines {
		// 	a := i.Keys[0]
		// 	fmt.Printf("a = %+v\n", a)
		// }

		if len(oneLines) != 0 {
			items := p.alignedItems(oneLines)
			buf.Write(p.indent(items))

			if index != len(o.List.Items) {
				buf.WriteByte(newline)
			}
		}

		if index == len(o.List.Items) {
			break
		}

		buf.Write(p.indent(p.objectItem(o.List.Items[index])))
		buf.WriteByte(newline)
		index++
	}

	buf.WriteString("}")
	buf.WriteByte(newline)
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

func lines(txt string) int {
	endline := 0
	for i := 0; i < len(txt); i++ {
		if txt[i] == '\n' {
			endline++
		}
	}
	return endline
}
