package printer

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/fatih/hcl/ast"
	"github.com/fatih/hcl/token"
)

const (
	blank   = byte(' ')
	newline = byte('\n')
	tab     = byte('\t')
)

type printer struct {
	cfg  Config
	prev ast.Node

	comments           []*ast.CommentGroup // may be nil, contains all comments
	standaloneComments []*ast.CommentGroup // contains all standalone comments (not assigned to any node)

	enableTrace bool
	indentTrace int
}

type ByPosition []*ast.CommentGroup

func (b ByPosition) Len() int           { return len(b) }
func (b ByPosition) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByPosition) Less(i, j int) bool { return b[i].Pos().Before(b[j].Pos()) }

func (p *printer) collectComments(node ast.Node) {
	// first collect all comments. This is already stored in
	// ast.File.(comments)
	ast.Walk(node, func(nn ast.Node) bool {
		switch t := nn.(type) {
		case *ast.File:
			p.comments = t.Comments
			return false
		}
		return true
	})

	standaloneComments := make(map[token.Pos]*ast.CommentGroup, 0)
	for _, c := range p.comments {
		standaloneComments[c.Pos()] = c
	}

	// next remove all lead and line comments from the overall comment map.
	// This will give us comments which are standalone, comments which are not
	// assigned to any kind of node.
	ast.Walk(node, func(nn ast.Node) bool {
		switch t := nn.(type) {
		case *ast.ObjectItem:
			if t.LeadComment != nil {
				for _, comment := range t.LeadComment.List {
					if _, ok := standaloneComments[comment.Pos()]; ok {
						delete(standaloneComments, comment.Pos())
					}
				}
			}

			if t.LineComment != nil {
				for _, comment := range t.LineComment.List {
					if _, ok := standaloneComments[comment.Pos()]; ok {
						delete(standaloneComments, comment.Pos())
					}
				}
			}
		}

		return true
	})

	for _, c := range standaloneComments {
		p.standaloneComments = append(p.standaloneComments, c)
	}
	sort.Sort(ByPosition(p.standaloneComments))

	fmt.Printf("standaloneComments = %+v\n", len(p.standaloneComments))
	for _, c := range p.standaloneComments {
		for _, comment := range c.List {
			fmt.Printf("comment = %+v\n", comment)
		}
	}

}

var count int

// output prints creates b printable HCL output and returns it.
func (p *printer) output(n interface{}) []byte {
	var buf bytes.Buffer
	count++

	switch t := n.(type) {
	case *ast.File:
		return p.output(t.Node)
	case *ast.ObjectList:
		for i, item := range t.Items {
			fmt.Printf("[%d] item: %s\n", i, item.Keys[0].Token.Text)
			buf.Write(p.output(item))
			if i != len(t.Items)-1 {
				buf.Write([]byte{newline, newline})
			}
		}
	case *ast.ObjectKey:
		buf.WriteString(t.Token.Text)
	case *ast.ObjectItem:
		for _, c := range p.standaloneComments {
			for _, comment := range c.List {
				fmt.Printf("[%d] OBJECTITEM p.prev = %+v\n", count, p.prev.Pos())
				fmt.Printf("[%d] OBJECTITEM comment.Pos() = %+v\n", count, comment.Pos())
				fmt.Printf("[%d] OBJECTTYPE t.Pos() = %+v\n", count, t.Pos())
				if comment.Pos().After(p.prev.Pos()) && comment.Pos().Before(t.Pos()) {
					buf.WriteString(comment.Text)
					buf.WriteByte(newline)
					buf.WriteByte(newline)
				}
			}
		}

		p.prev = t
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

	// if item, ok := n.(ast.Node); ok {
	// 	p.prev = item
	// }

	return buf.Bytes()
}

func (p *printer) objectItem(o *ast.ObjectItem) []byte {
	defer un(trace(p, fmt.Sprintf("ObjectItem: %s", o.Keys[0].Token.Text)))
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

func (p *printer) objectType(o *ast.ObjectType) []byte {
	defer un(trace(p, "ObjectType"))
	var buf bytes.Buffer
	buf.WriteString("{")
	buf.WriteByte(newline)

	for _, c := range p.standaloneComments {
		for _, comment := range c.List {
			fmt.Printf("[%d] OBJECTTYPE p.prev = %+v\n", count, p.prev.Pos())
			fmt.Printf("[%d] OBJECTTYPE comment.Pos() = %+v\n", count, comment.Pos())
			fmt.Printf("[%d] OBJECTTYPE t.Pos() = %+v\n", count, o.Pos())
			firstItem := o.List.Pos()
			if comment.Pos().After(p.prev.Pos()) && comment.Pos().Before(firstItem) {
				buf.Write(p.indent([]byte(comment.Text))) // TODO(arslan): indent
				buf.WriteByte(newline)
				buf.WriteByte(newline)
			}
		}
	}

	var index int
	for {
		// check if we have adjacent one liner items. If yes we'll going to align
		// the comments.
		var aligned []*ast.ObjectItem
		for i, item := range o.List.Items[index:] {
			// we don't group one line lists
			if len(o.List.Items) == 1 {
				break
			}

			// one means a oneliner with out any lead comment
			// two means a oneliner with lead comment
			// anything else might be something else
			cur := lines(string(p.objectItem(item)))
			if cur > 2 {
				break
			}

			next := 0
			if index != len(o.List.Items)-1 {
				next = lines(string(p.objectItem(o.List.Items[index+1])))
			}

			prev := 0
			if index != 0 {
				prev = lines(string(p.objectItem(o.List.Items[index-1])))
			}

			if (cur == next && next == 1) || (next == 1 && cur == 2 && i == 0) {
				aligned = append(aligned, item)
				index++
			} else if (cur == prev && prev == 1) || (prev == 2 && cur == 1) {
				aligned = append(aligned, item)
				index++
			} else {
				break
			}
		}

		// fmt.Printf("==================> len(aligned) = %+v\n", len(aligned))
		// for _, b := range aligned {
		// 	fmt.Printf("b = %+v\n", b)
		// }

		// put newlines if the items are between other non aligned items
		if index != len(aligned) {
			buf.WriteByte(newline)
		}

		if len(aligned) >= 1 {
			p.prev = aligned[len(aligned)-1]

			items := p.alignedItems(aligned)
			buf.Write(p.indent(items))
		} else {
			p.prev = o.List.Items[index]

			buf.Write(p.indent(p.objectItem(o.List.Items[index])))
			index++
		}

		buf.WriteByte(newline)

		if index == len(o.List.Items) {
			break
		}

	}

	buf.WriteString("}")
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

	for i, item := range items {
		if item.LeadComment != nil {
			for _, comment := range item.LeadComment.List {
				buf.WriteString(comment.Text)
				buf.WriteByte(newline)
			}
		}

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

		// do not print for the last item
		if i != len(items)-1 {
			buf.WriteByte(newline)
		}
	}

	return buf.Bytes()
}

func (p *printer) literal(l *ast.LiteralType) []byte {
	return []byte(l.Token.Text)
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
	endline := 1
	for i := 0; i < len(txt); i++ {
		if txt[i] == '\n' {
			endline++
		}
	}
	return endline
}

// ----------------------------------------------------------------------------
// Tracing support

func (p *printer) printTrace(a ...interface{}) {
	if !p.enableTrace {
		return
	}

	const dots = ". . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . "
	const n = len(dots)
	i := 2 * p.indentTrace
	for i > n {
		fmt.Print(dots)
		i -= n
	}
	// i <= n
	fmt.Print(dots[0:i])
	fmt.Println(a...)
}

func trace(p *printer, msg string) *printer {
	p.printTrace(msg, "(")
	p.indentTrace++
	return p
}

// Usage pattern: defer un(trace(p, "..."))
func un(p *printer) {
	p.indentTrace--
	p.printTrace(")")
}
