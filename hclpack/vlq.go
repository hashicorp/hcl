package hclpack

import (
	"github.com/bsm/go-vlq"
)

type vlqBuf []byte

var vlqSpace [vlq.MaxLen64]byte

func newVLQBuf(byteCap int) vlqBuf {
	return make(vlqBuf, 0, byteCap)
}

func (b vlqBuf) AppendInt(i int) vlqBuf {
	spc := cap(b) - len(b)
	if spc < len(vlqSpace) {
		b = append(b, vlqSpace[:]...)
		b = b[:len(b)-len(vlqSpace)]
	}
	into := b[len(b):cap(b)]
	l := vlq.PutInt(into, int64(i))
	b = b[:len(b)+l]
	return b
}

func (b vlqBuf) AppendRawByte(by byte) vlqBuf {
	return append(b, by)
}

func (b vlqBuf) Bytes() []byte {
	return []byte(b)
}
