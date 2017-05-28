package zclsyntax

import (
	"github.com/apparentlymart/go-textseg/textseg"
	"github.com/zclconf/go-zcl/zcl"
)

// Token represents a sequence of bytes from some zcl code that has been
// tagged with a type and its range within the source file.
type Token struct {
	Type  TokenType
	Bytes []byte
	Range zcl.Range
}

// TokenType is an enumeration used for the Type field on Token.
type TokenType rune

const (
	// Single-character tokens are represented by their own character, for
	// convenience in producing these within the scanner. However, the values
	// are otherwise arbitrary and just intended to be mnemonic for humans
	// who might see them in debug output.

	TokenOBrace TokenType = '{'
	TokenCBrace TokenType = '}'
	TokenOBrack TokenType = '['
	TokenCBrack TokenType = ']'
	TokenOParen TokenType = '('
	TokenCParen TokenType = ')'
	TokenOQuote TokenType = '«'
	TokenCQuote TokenType = '»'

	TokenDot   TokenType = '.'
	TokenStar  TokenType = '*'
	TokenSlash TokenType = '/'
	TokenPlus  TokenType = '+'
	TokenMinus TokenType = '-'

	TokenEqual         TokenType = '='
	TokenNotEqual      TokenType = '≠'
	TokenLessThan      TokenType = '<'
	TokenLessThanEq    TokenType = '≤'
	TokenGreaterThan   TokenType = '>'
	TokenGreaterThanEq TokenType = '≥'

	TokenAnd  TokenType = '∧'
	TokenOr   TokenType = '∨'
	TokenBang TokenType = '!'

	TokenQuestion TokenType = '?'
	TokenColon    TokenType = ':'

	TokenTemplateInterp  TokenType = '∫'
	TokenTemplateControl TokenType = 'λ'
	TokenTemplateSeqEnd  TokenType = '∎'

	TokenStringLit TokenType = 'S'
	TokenHeredoc   TokenType = 'H'
	TokenNumberLit TokenType = 'N'
	TokenIdent     TokenType = 'I'

	TokenNewline TokenType = '\n'
	TokenEOF     TokenType = '␄'

	// The rest are not used in the language but recognized by the scanner so
	// we can generate good diagnostics in the parser when users try to write
	// things that might work in other languages they are familiar with, or
	// simply make incorrect assumptions about the zcl language.

	TokenBitwiseAnd TokenType = '&'
	TokenBitwiseOr  TokenType = '|'
	TokenBitwiseNot TokenType = '~'
	TokenBitwiseXor TokenType = '^'
	TokenStarStar   TokenType = '➚'
	TokenBacktick   TokenType = '`'
	TokenSemicolon  TokenType = ';'
	TokenTab        TokenType = '␉'
	TokenInvalid    TokenType = '�'
	TokenBadUTF8    TokenType = '💩'
)

type tokenAccum struct {
	Filename string
	Bytes    []byte
	Start    zcl.Pos
	Tokens   []Token
}

func (f *tokenAccum) emitToken(ty TokenType, startOfs int, endOfs int) {
	// Walk through our buffer to figure out how much we need to adjust
	// the start pos to get our end pos.

	start := f.Start
	start.Byte += startOfs
	start.Column += startOfs // Safe because only ASCII spaces can be in the offset

	end := start
	end.Byte = f.Start.Byte + endOfs
	b := f.Bytes
	for len(b) > 0 {
		advance, seq, _ := textseg.ScanGraphemeClusters(b, true)
		if len(seq) == 1 && seq[0] == '\n' {
			end.Line++
			end.Column = 1
		} else {
			end.Column++
		}
		b = b[advance:]
	}

	f.Tokens = append(f.Tokens, Token{
		Type:  ty,
		Bytes: f.Bytes[startOfs:endOfs],
		Range: zcl.Range{
			Filename: f.Filename,
			Start:    start,
			End:      end,
		},
	})
}
