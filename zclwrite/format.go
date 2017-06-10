package zclwrite

import (
	"github.com/zclconf/go-zcl/zcl/zclsyntax"
)

// format rewrites tokens within the given sequence, in-place, to adjust the
// whitespace around their content to achieve canonical formatting.
func format(tokens Tokens) {
	// currently does nothing
}

func linesForFormat(tokens Tokens) []formatLine {
	if len(tokens) == 0 {
		// should never happen, since we should always have EOF, but let's
		// not crash anyway.
		return make([]formatLine, 0)
	}

	// first we'll count our lines, so we can allocate the array for them in
	// a single block. (We want to minimize memory pressure in this codepath,
	// so it can be run somewhat-frequently by editor integrations.)
	lineCount := 1 // if there are zero newlines then there is one line
	for _, tok := range tokens {
		if tokenIsNewline(tok) {
			lineCount++
		}
	}

	// To start, we'll just put everything in the "lead" cell on each line,
	// and then do another pass over the lines afterwards to adjust.
	lines := make([]formatLine, lineCount)
	li := 0
	lineStart := 0
	for i, tok := range tokens {
		if tok.Type == zclsyntax.TokenEOF {
			// The EOF token doesn't belong to any line, and terminates the
			// token sequence.
			lines[li].lead = tokens[lineStart:i]
			break
		}

		if tokenIsNewline(tok) {
			lines[li].lead = tokens[lineStart : i+1]
			lineStart = i + 1
			li++
		}
	}

	// Now we'll pick off any trailing comments and attribute assignments
	// to shuffle off into the "comment" and "assign" cells.
	for i := range lines {
		line := &lines[i]
		if len(line.lead) == 0 {
			// if the line is empty then there's nothing for us to do
			// (this should happen only for the final line, because all other
			// lines would have a newline token of some kind)
			continue
		}

		if len(line.lead) > 1 && line.lead[len(line.lead)-1].Type == zclsyntax.TokenComment {
			line.comment = line.lead[len(line.lead)-1:]
			line.lead = line.lead[:len(line.lead)-1]
		}

		for i, tok := range line.lead {
			if tok.Type == zclsyntax.TokenEqual {
				line.assign = line.lead[i:]
				line.lead = line.lead[:i]
			}
		}
	}

	return lines
}

func tokenIsNewline(tok *Token) bool {
	if tok.Type == zclsyntax.TokenNewline {
		return true
	} else if tok.Type == zclsyntax.TokenComment {
		// Single line tokens (# and //) consume their terminating newline,
		// so we need to treat them as newline tokens as well.
		if len(tok.Bytes) > 0 && tok.Bytes[len(tok.Bytes)-1] == '\n' {
			return true
		}
	}
	return false
}

// formatLine represents a single line of source code for formatting purposes,
// splitting its tokens into up to three "cells":
//
// lead: always present, representing everything up to one of the others
// assign: if line contains an attribute assignment, represents the tokens
//    starting at (and including) the equals symbol
// comment: if line contains any non-comment tokens and ends with a
//    single-line comment token, represents the comment.
//
// When formatting, the leading spaces of the first tokens in each of these
// cells is adjusted to align vertically their occurences on consecutive
// rows.
type formatLine struct {
	lead    Tokens
	assign  Tokens
	comment Tokens
}
