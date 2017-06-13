package zclsyntax

import (
	"bytes"
)

type Keyword []byte

var forKeyword = Keyword([]byte{'f', 'o', 'r'})
var inKeyword = Keyword([]byte{'i', 'n'})
var ifKeyword = Keyword([]byte{'i', 'f'})

func (kw Keyword) TokenMatches(token Token) bool {
	if token.Type != TokenIdent {
		return false
	}
	return bytes.Equal([]byte(kw), token.Bytes)
}
