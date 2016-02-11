package printer

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/token"
)

const (
	blank    = byte(' ')
	newline  = byte('\n')
	tab      = byte('\t')
	infinity = 1 << 30 // offset or line
)

var (
	unindent = []byte("\uE123") // in the private use space
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
	ast.Walk(node, func(nn ast.Node) (ast.Node, bool) {
		switch t := nn.(type) {
		case *ast.File:
			p.comments = t.Comments
			return nn, false
		}
		return nn, true
	})

	standaloneComments := make(map[token.Pos]*ast.CommentGroup, 0)
	for _, c := range p.comments {
		standaloneComments[c.Pos()] = c
	}

	// next remove all lead and line comments from the overall comment map.
	// This will give us comments which are standalone, comments which are not
	// assigned to any kind of node.
	ast.Walk(node, func(nn ast.Node) (ast.Node, bool) {
		switch t := nn.(type) {
		case *ast.LiteralType:
			if t.LineComment != nil {
				for _, comment := range t.LineComment.List {
					if _, ok := standaloneComments[comment.Pos()]; ok {
						delete(standaloneComments, comment.Pos())
					}
				}
			}
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

		return nn, true
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
		end := token.Pos{Offset: infinity, Line: infinity}
		buf.Write(p.objectList(t, end))
	case *ast.ObjectKey:
		buf.WriteString(t.Token.Text)
	case *ast.ObjectItem:
		p.prev = t.Pos()
		buf.Write(p.objectItem(t))
	case *ast.LiteralType:
		buf.Write(p.literalType(t))
	case *ast.ListType:
		buf.Write(p.list(t))
	case *ast.ObjectType:
		buf.Write(p.objectType(t))
	default:
		fmt.Printf(" unknown type: %T\n", n)
	}

	return buf.Bytes()
}

func (p *printer) literalType(lit *ast.LiteralType) []byte {
	result := []byte(lit.Token.Text)
	if lit.Token.Type == token.HEREDOC {
		// Clear the trailing newline from heredocs
		if result[len(result)-1] == '\n' {
			result = result[:len(result)-1]
		}

		// Poison lines 2+ so that we don't indent them
		result = p.heredocIndent(result)
	}

	return result
}

// objectItem returns the printable HCL form of an object item. An object type
// starts with one/multiple keys and has a value. The value might be of any
// type.
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
		if o.Assign.IsValid() && i == len(o.Keys)-1 && len(o.Keys) == 1 {
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

// objectList returns the printable HCL form of a list of objects.
func (p *printer) objectList(o *ast.ObjectList, end token.Pos) []byte {
	var buf bytes.Buffer
	var index int
	var nextItem token.Pos
	for {
		commented := false

		// Print stand alone comments
		for _, c := range p.standaloneComments {
			for _, comment := range c.List {
				// if we hit the end, last item should be the brace
				if index != len(o.Items) {
					nextItem = o.Items[index].Pos()
				} else {
					nextItem = end
				}

				if comment.Pos().After(p.prev) && comment.Pos().Before(nextItem) {
					// add newline if it's between other printed nodes
					if index > 0 {
						commented = true
						buf.WriteByte(newline)
					}

					buf.Write([]byte(comment.Text))
					buf.WriteByte(newline)
					if index != len(o.Items) {
						buf.WriteByte(newline) // do not print on the end
					}
				}
			}
		}

		if index == len(o.Items) {
			p.prev = end
			break
		}

		// check if we have adjacent one liner items. If yes we'll going to align
		// the comments.
		var aligned []*ast.ObjectItem
		for _, item := range o.Items[index:] {
			// we don't group one line lists
			if len(o.Items) == 1 {
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
			if index != len(o.Items)-1 {
				nextPos = o.Items[index+1].Pos()
			}

			prevPos := token.Pos{}
			if index != 0 {
				prevPos = o.Items[index-1].Pos()
			}

			// fmt.Println("DEBUG ----------------")
			// fmt.Printf("prev = %#v prevPos: %s\n", "", prevPos)
			// fmt.Printf("cur = %+v curPos: %s\n", cur, curPos)
			// fmt.Printf("next = %#v nextPos: %s\n", "", nextPos)

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
			buf.Write(items)
		} else {
			p.prev = o.Items[index].Pos()

			buf.Write(p.objectItem(o.Items[index]))
			index++
		}

		buf.WriteByte(newline)
	}

	return buf.Bytes()
}

// objectType returns the printable HCL form of an object type. An object type
// begins with a brace and ends with a brace.
func (p *printer) objectType(o *ast.ObjectType) []byte {
	defer un(trace(p, "ObjectType"))
	var buf bytes.Buffer
	buf.WriteString("{")
	buf.WriteByte(newline)
	buf.Write(p.indent(p.objectList(o.List, o.Rbrace)))
	buf.WriteString("}")
	return buf.Bytes()
}

func (p *printer) alignedItems(items []*ast.ObjectItem) []byte {
	var buf bytes.Buffer

	// find the longest key and value length, needed for alignment
	var longestKeyLen int // longest key length
	var longestValLen int // longest value length
	for _, item := range items {
		key := len(item.Keys[0].Token.Text)
		val := len(p.output(item.Val))

		if key > longestKeyLen {
			longestKeyLen = key
		}

		if val > longestValLen {
			longestValLen = val
		}
	}

	for i, item := range items {
		if item.LeadComment != nil {
			for _, comment := range item.LeadComment.List {
				buf.WriteString(comment.Text)
				buf.WriteByte(newline)
			}
		}

		for i, k := range item.Keys {
			keyLen := len(k.Token.Text)
			buf.WriteString(k.Token.Text)
			for i := 0; i < longestKeyLen-keyLen+1; i++ {
				buf.WriteByte(blank)
			}

			// reach end of key
			if i == len(item.Keys)-1 && len(item.Keys) == 1 {
				buf.WriteString("=")
				buf.WriteByte(blank)
			}
		}

		val := p.output(item.Val)
		valLen := len(val)
		buf.Write(val)

		if item.Val.Pos().Line == item.Keys[0].Pos().Line && item.LineComment != nil {
			for i := 0; i < longestValLen-valLen+1; i++ {
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

// list returns the printable HCL form of an list type.
func (p *printer) list(l *ast.ListType) []byte {
	var buf bytes.Buffer
	buf.WriteString("[")

	var longestLine int
	for _, item := range l.List {
		// for now we assume that the list only contains literal types
		if lit, ok := item.(*ast.LiteralType); ok {
			lineLen := len(lit.Token.Text)
			if lineLen > longestLine {
				longestLine = lineLen
			}
		}
	}

	insertSpaceBeforeItem := false
	for i, item := range l.List {
		if item.Pos().Line != l.Lbrack.Line {
			// multiline list, add newline before we add each item
			buf.WriteByte(newline)
			insertSpaceBeforeItem = false
			// also indent each line
			val := p.output(item)
			curLen := len(val)
			buf.Write(p.indent(val))
			buf.WriteString(",")

			if lit, ok := item.(*ast.LiteralType); ok && lit.LineComment != nil {
				// if the next item doesn't have any comments, do not align
				buf.WriteByte(blank) // align one space
				for i := 0; i < longestLine-curLen; i++ {
					buf.WriteByte(blank)
				}

				for _, comment := range lit.LineComment.List {
					buf.WriteString(comment.Text)
				}
			}

			if i == len(l.List)-1 {
				buf.WriteByte(newline)
			}
		} else {
			if insertSpaceBeforeItem {
				buf.WriteByte(blank)
				insertSpaceBeforeItem = false
			}
			buf.Write(p.output(item))
			if i != len(l.List)-1 {
				buf.WriteString(",")
				insertSpaceBeforeItem = true
			}
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

// unindent removes all the indentation from the tombstoned lines
func (p *printer) unindent(buf []byte) []byte {
	var res []byte
	for i := 0; i < len(buf); i++ {
		skip := len(buf)-i <= len(unindent)
		if !skip {
			skip = !bytes.Equal(unindent, buf[i:i+len(unindent)])
		}
		if skip {
			res = append(res, buf[i])
			continue
		}

		// We have a marker. we have to backtrace here and clean out
		// any whitespace ahead of our tombstone up to a \n
		for j := len(res) - 1; j >= 0; j-- {
			if res[j] == '\n' {
				break
			}

			res = res[:j]
		}

		// Skip the entire unindent marker
		i += len(unindent) - 1
	}

	return res
}

// heredocIndent marks all the 2nd and further lines as unindentable
func (p *printer) heredocIndent(buf []byte) []byte {
	var res []byte
	bol := false
	for _, c := range buf {
		if bol && c != '\n' {
			res = append(res, unindent...)
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
