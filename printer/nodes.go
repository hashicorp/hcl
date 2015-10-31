package printer

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/fatih/hcl/ast"
	"github.com/fatih/hcl/token"
)

const (
	blank    = byte(' ')
	newline  = byte('\n')
	tab      = byte('\t')
	infinity = 1 << 30 // offset or line
)

type printer struct {
	cfg  Config
	prev token.Pos

	comments           []*ast.CommentGroup // may be nil, contains all comments
	standaloneComments []*ast.CommentGroup // contains all standalone comments (not assigned to any node)

	enableTrace bool
	indentTrace int
}

type ByPosition []*ast.CommentGroup

func (b ByPosition) Len() int           { return len(b) }
func (b ByPosition) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByPosition) Less(i, j int) bool { return b[i].Pos().Before(b[j].Pos()) }

// collectComments comments all standalone comments which are not lead or line
// comment
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
}

// output prints creates b printable HCL output and returns it.
func (p *printer) output(n interface{}) []byte {
	var buf bytes.Buffer

	switch t := n.(type) {
	case *ast.File:
		return p.output(t.Node)
	case *ast.ObjectList:
		var index int
		var nextItem token.Pos
		var commented bool
		for {
			// TODO(arslan): refactor below comment printing, we have the same in objectType

			// print upper leve stand alone comments
			for _, c := range p.standaloneComments {
				for _, comment := range c.List {
					if index != len(t.Items) {
						nextItem = t.Items[index].Pos()
					} else {
						nextItem = token.Pos{Offset: infinity, Line: infinity}
					}

					if comment.Pos().After(p.prev) && comment.Pos().Before(nextItem) {
						// if we hit the end add newlines so we can print the comment
						if index == len(t.Items) {
							buf.Write([]byte{newline, newline})
						}

						buf.WriteString(comment.Text)
						// TODO(arslan): do not print new lines if the comments are one liner

						buf.WriteByte(newline)
						if index != len(t.Items) {
							buf.WriteByte(newline)
						}
					}
				}
			}

			if index == len(t.Items) {
				break
			}

			buf.Write(p.output(t.Items[index]))
			if !commented && index != len(t.Items)-1 {
				buf.Write([]byte{newline, newline})
			}
			index++
		}
	case *ast.ObjectKey:
		buf.WriteString(t.Token.Text)
	case *ast.ObjectItem:
		p.prev = t.Pos()
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

	var index int
	var nextItem token.Pos
	var commented bool
	for {
		// Print stand alone comments
		for _, c := range p.standaloneComments {
			for _, comment := range c.List {
				// if we hit the end, last item should be the brace
				if index != len(o.List.Items) {
					nextItem = o.List.Items[index].Pos()
				} else {
					nextItem = o.Rbrace
				}

				if comment.Pos().After(p.prev) && comment.Pos().Before(nextItem) {
					// add newline if it's between other printed nodes
					if index > 0 {
						commented = true
						buf.WriteByte(newline)
					}

					buf.Write(p.indent([]byte(comment.Text)))
					buf.WriteByte(newline)
					if index != len(o.List.Items) {
						buf.WriteByte(newline) // do not print on the end
					}
				}
			}
		}

		if index == len(o.List.Items) {
			p.prev = o.Rbrace
			break
		}

		// check if we have adjacent one liner items. If yes we'll going to align
		// the comments.
		var aligned []*ast.ObjectItem
		for _, item := range o.List.Items[index:] {
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

			curPos := item.Pos()

			nextPos := token.Pos{}
			if index != len(o.List.Items)-1 {
				nextPos = o.List.Items[index+1].Pos()
			}

			prevPos := token.Pos{}
			if index != 0 {
				prevPos = o.List.Items[index-1].Pos()
			}

			// fmt.Println("DEBUG ----------------")
			// fmt.Printf("prev = %+v prevPos: %s\n", prev, prevPos)
			// fmt.Printf("cur = %+v curPos: %s\n", cur, curPos)
			// fmt.Printf("next = %+v nextPos: %s\n", next, nextPos)

			if curPos.Line+1 == nextPos.Line {
				aligned = append(aligned, item)
				index++
				continue
			}

			if curPos.Line-1 == prevPos.Line {
				aligned = append(aligned, item)
				index++

				// finish if we have a new line or comment next. This happens
				// if the next item is not adjacent
				if curPos.Line+1 != nextPos.Line {
					break
				}
				continue
			}

			break
		}

		// put newlines if the items are between other non aligned items.
		// newlines are also added if there is a standalone comment already, so
		// check it too
		if !commented && index != len(aligned) {
			buf.WriteByte(newline)
		}

		if len(aligned) >= 1 {
			p.prev = aligned[len(aligned)-1].Pos()

			items := p.alignedItems(aligned)
			buf.Write(p.indent(items))
		} else {
			p.prev = o.List.Items[index].Pos()

			buf.Write(p.indent(p.objectItem(o.List.Items[index])))
			index++
		}

		buf.WriteByte(newline)
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
