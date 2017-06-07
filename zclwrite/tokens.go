package zclwrite

import (
	"github.com/zclconf/go-zcl/zcl/zclsyntax"
)

// TokenGen is an abstract type that can append tokens to a list. It is the
// low-level foundation underlying the zclwrite AST; the AST provides a
// convenient abstraction over raw token sequences to facilitate common tasks,
// but it's also possible to directly manipulate the tree of token generators
// to make changes that the AST API doesn't directly allow.
type TokenGen interface {
	EachToken(TokenCallback)
}

// TokenCallback is used with TokenGen implementations to specify the action
// that is to be taken for each token in the flattened token sequence.
type TokenCallback func(*Token)

// Token is a single sequence of bytes annotated with a type. It is similar
// in purpose to zclsyntax.Token, but discards the source position information
// since that is not useful in code generation.
type Token struct {
	Type  zclsyntax.TokenType
	Bytes []byte

	// We record the number of spaces before each token so that we can
	// reproduce the exact layout of the original file when we're making
	// surgical changes in-place. When _new_ code is created it will always
	// be in the canonical style, but we preserve layout of existing code.
	SpacesBefore int
}

// Tokens is a flat list of tokens.
type Tokens []*Token

// TokenSeq combines zero or more TokenGens together to produce a flat sequence
// of tokens from a tree of TokenGens.
type TokenSeq []TokenGen

func (t *Token) EachToken(cb TokenCallback) {
	cb(t)
}

func (ts Tokens) EachToken(cb TokenCallback) {
	for _, t := range ts {
		cb(t)
	}
}

func (ts *TokenSeq) EachToken(cb TokenCallback) {
	for _, gen := range *ts {
		gen.EachToken(cb)
	}
}

// Tokens returns the flat list of tokens represented by the receiving
// token sequence.
func (ts *TokenSeq) Tokens() Tokens {
	var tokens Tokens
	ts.EachToken(func(token *Token) {
		tokens = append(tokens, token)
	})
	return tokens
}

func (ts *TokenSeq) Append(other *TokenSeq) {
	*ts = append(*ts, other)
}

// TokenSeqEmpty is a TokenSeq that contains no tokens. It can be used anywhere,
// but its primary purpose is to be assigned as a replacement for a non-empty
// TokenSeq when eliminating a section of an input file.
var TokenSeqEmpty = TokenSeq([]TokenGen(nil))
