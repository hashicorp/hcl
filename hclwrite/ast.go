// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package hclwrite

import (
	"bytes"
	"io"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

type File struct {
	inTree

	srcBytes []byte
	body     *node
}

// NewEmptyFile constructs a new file with no content, ready to be mutated
// by other calls that append to its body.
func NewEmptyFile() *File {
	f := &File{
		inTree: newInTree(),
	}
	body := newBody()
	f.body = f.children.Append(body)
	return f
}

// Body returns the root body of the file, which contains the top-level
// attributes and blocks.
func (f *File) Body() *Body {
	return f.body.content.(*Body)
}

// WriteTo writes the tokens underlying the receiving file to the given writer.
//
// The tokens first have a simple formatting pass applied that adjusts only
// the spaces between them.
func (f *File) WriteTo(wr io.Writer) (int64, error) {
	tokens := f.children.BuildTokens(nil)
	format(tokens)
	return tokens.WriteTo(wr)
}

// Bytes returns a buffer containing the source code resulting from the
// tokens underlying the receiving file. If any updates have been made via
// the AST API, these will be reflected in the result.
func (f *File) Bytes() []byte {
	buf := &bytes.Buffer{}
	//nolint:errcheck // FIXME: Propogate errors upward.
	f.WriteTo(buf)
	return buf.Bytes()
}

type comments struct {
	leafNode

	tokens Tokens
}

func newComments(tokens Tokens) *comments {
	return &comments{
		tokens: tokens,
	}
}

func (c *comments) BuildTokens(to Tokens) Tokens {
	return c.tokens.BuildTokens(to)
}

type identifier struct {
	leafNode

	token *Token
}

func newIdentifier(token *Token) *identifier {
	return &identifier{
		token: token,
	}
}

func (i *identifier) BuildTokens(to Tokens) Tokens {
	return append(to, i.token)
}

func (i *identifier) hasName(name string) bool {
	return name == string(i.token.Bytes)
}

type number struct {
	leafNode

	token *Token
}

func newNumber(token *Token) *number {
	return &number{
		token: token,
	}
}

func (n *number) BuildTokens(to Tokens) Tokens {
	return append(to, n.token)
}

type quoted struct {
	leafNode

	tokens Tokens
}

func newQuoted(tokens Tokens) *quoted {
	return &quoted{
		tokens: tokens,
	}
}

func (q *quoted) BuildTokens(to Tokens) Tokens {
	return q.tokens.BuildTokens(to)
}

// hasValue returns true if the quoted string literal, with any escape
// sequences decoded, is equal to the given string.
func (q *quoted) hasValue(s string) bool {
	expr, diags := hclsyntax.ParseExpression(q.tokens.Bytes(), "", hcl.InitialPos)
	if diags.HasErrors() {
		return false
	}
	val, diags := expr.Value(nil)
	if diags.HasErrors() || !val.Type().Equals(cty.String) || !val.IsKnown() || val.IsNull() {
		return false
	}
	return val.AsString() == s
}
