package zclwrite

import (
	"bytes"
	"io"
)

type Node interface {
	walkChildNodes(w internalWalkFunc)
	Tokens() *TokenSeq
}

type internalWalkFunc func(Node)

type File struct {
	Name     string
	SrcBytes []byte

	Body      *Body
	AllTokens *TokenSeq
}

// WriteTo writes the tokens underlying the receiving file to the given writer.
func (f *File) WriteTo(wr io.Writer) (int, error) {
	return f.AllTokens.WriteTo(wr)
}

// Bytes returns a buffer containing the source code resulting from the
// tokens underlying the receiving file. If any updates have been made via
// the AST API, these will be reflected in the result.
func (f *File) Bytes() []byte {
	buf := &bytes.Buffer{}
	f.WriteTo(buf)
	return buf.Bytes()
}

// Format makes in-place modifications to the tokens underlying the receiving
// file in order to change the whitespace to be in canonical form.
func (f *File) Format() {
	format(f.Body.AllTokens.Tokens())
}

type Body struct {
	// Items may contain Attribute, Block and Unstructured instances.
	// Items and AllTokens should be updated only by methods of this type,
	// since they must be kept synchronized for correct operation.
	Items     []Node
	AllTokens *TokenSeq

	// IndentLevel is the number of spaces that should appear at the start
	// of lines added within this body.
	IndentLevel int
}

func (n *Body) walkChildNodes(w internalWalkFunc) {
	for _, item := range n.Items {
		w(item)
	}
}

func (n *Body) Tokens() *TokenSeq {
	return n.AllTokens
}

func (n *Body) AppendItem(node Node) {
	if n.AllTokens == nil {
		new := make(TokenSeq, 0, 1)
		n.AllTokens = &new
	}
	n.Items = append(n.Items, node)
	*(n.AllTokens) = append(*(n.AllTokens), node.Tokens())
}

type Attribute struct {
	AllTokens *TokenSeq

	LeadCommentTokens *TokenSeq
	NameTokens        *TokenSeq
	EqualsTokens      *TokenSeq
	Expr              *Expression
	LineCommentTokens *TokenSeq
	EOLTokens         *TokenSeq
}

func (a *Attribute) walkChildNodes(w internalWalkFunc) {
	w(a.Expr)
}

func (n *Attribute) Tokens() *TokenSeq {
	return n.AllTokens
}

type Block struct {
	AllTokens *TokenSeq

	LeadCommentTokens *TokenSeq
	TypeTokens        *TokenSeq
	LabelTokens       []*TokenSeq
	LabelTokensFlat   *TokenSeq
	OBraceTokens      *TokenSeq
	Body              *Body
	CBraceTokens      *TokenSeq
	EOLTokens         *TokenSeq
}

func (n *Block) walkChildNodes(w internalWalkFunc) {
	w(n.Body)
}

// Unstructured represents consecutive sets of tokens within a Body that
// aren't part of any particular construct. This includes blank lines
// and comments that aren't immediately before an attribute or nested block.
type Unstructured struct {
	AllTokens *TokenSeq
}

func (n *Unstructured) Tokens() *TokenSeq {
	return n.AllTokens
}

func (n *Unstructured) walkChildNodes(w internalWalkFunc) {
	// no child nodes
}

type Expression struct {
	AllTokens *TokenSeq
	VarRefs   []*VarRef
}

func (n *Expression) walkChildNodes(w internalWalkFunc) {
	for _, name := range n.VarRefs {
		w(name)
	}
}

func (n *Expression) Tokens() *TokenSeq {
	return n.AllTokens
}

type VarRef struct {
	// Tokens alternate between TokenIdent and TokenDot, with the first
	// and last elements always being TokenIdent.
	AllTokens *TokenSeq
}

func (n *VarRef) walkChildNodes(w internalWalkFunc) {
	// no child nodes of a variable name
}

func (n *VarRef) Tokens() *TokenSeq {
	return n.AllTokens
}
