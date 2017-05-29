package zclwrite

type Node interface {
	walkChildNodes(w internalWalkFunc)
	Tokens() *TokenSeq
}

type internalWalkFunc func(Node)

type File struct {
	Name  string
	Bytes []byte

	Body Body
}

type Body struct {
	// Items may contain Attribute, Block and Unstructured instances.
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

type Attribute struct {
	AllTokens *TokenSeq

	LeadCommentTokens *TokenSeq
	NameToken         *Token
	EqualsToken       *Token
	Value             *Expression
	LineCommentTokens *TokenSeq
	EOLToken          *Token
}

func (a *Attribute) walkChildNodes(w internalWalkFunc) {
	w(a.Value)
}

type Block struct {
	AllTokens *TokenSeq

	LeadCommentTokens *TokenSeq
	TypeToken         *Token
	LabelTokens       []*TokenSeq
	LabelTokensFlat   *TokenSeq
	OBraceToken       *Token
	Body              *Body
	CBraceToken       *Token
	EOLToken          *Token
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
