package hclpack

import (
	"errors"

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

func (b vlqBuf) ReadInt() (int, vlqBuf, error) {
	v, adv := vlq.Int([]byte(b))
	if adv <= 0 {
		if adv == 0 {
			return 0, b, errors.New("missing expected VLQ value")
		} else {
			return 0, b, errors.New("invalid VLQ value")
		}
	}
	if int64(int(v)) != v {
		return 0, b, errors.New("VLQ value too big for integer on this platform")
	}
	return int(v), b[adv:], nil
}

func (b vlqBuf) AppendRawByte(by byte) vlqBuf {
	return append(b, by)
}

func (b vlqBuf) Bytes() []byte {
	return []byte(b)
}
