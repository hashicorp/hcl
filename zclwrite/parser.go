package zclwrite

import (
	"github.com/zclconf/go-zcl/zcl"
	"github.com/zclconf/go-zcl/zcl/zclsyntax"
)

// lexConfig uses the zclsyntax scanner to get a token stream and then
// rewrites it into this package's token model.
func lexConfig(src []byte) Tokens {
	mainTokens := zclsyntax.LexConfig(src, "", zcl.Pos{Byte: 0, Line: 1, Column: 1})
	ret := make(Tokens, len(mainTokens))
	var lastByteOffset int
	for i, mainToken := range mainTokens {
		// Create a copy of the bytes so that we can mutate without
		// corrupting the original token stream.
		bytes := make([]byte, len(mainToken.Bytes))
		copy(bytes, mainToken.Bytes)

		ret[i] = &Token{
			Type:  mainToken.Type,
			Bytes: bytes,

			// We assume here that spaces are always ASCII spaces, since
			// that's what the scanner also assumes, and thus the number
			// of bytes skipped is also the number of space characters.
			SpacesBefore: mainToken.Range.Start.Byte - lastByteOffset,
		}

		lastByteOffset = mainToken.Range.End.Byte
	}

	return ret
}
