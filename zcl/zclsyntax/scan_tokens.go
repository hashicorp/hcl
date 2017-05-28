// line 1 "scan_tokens.rl"
package zclsyntax

import (
	"github.com/zclconf/go-zcl/zcl"
)

// This file is generated from scan_tokens.rl. DO NOT EDIT.

// line 12 "scan_tokens.go"
var _zcltok_actions []byte = []byte{
	0, 1, 0, 1, 1, 1, 2, 1, 3,
	1, 4, 1, 5, 1, 6, 1, 7,
	1, 8,
}

var _zcltok_key_offsets []byte = []byte{
	0, 2, 4, 16, 17, 18, 20, 22,
}

var _zcltok_trans_keys []byte = []byte{
	128, 191, 128, 191, 9, 32, 128, 191,
	192, 223, 224, 239, 240, 247, 248, 255,
	9, 32, 128, 191, 128, 191, 128, 191,
}

var _zcltok_single_lengths []byte = []byte{
	0, 0, 2, 1, 1, 0, 0, 0,
}

var _zcltok_range_lengths []byte = []byte{
	1, 1, 5, 0, 0, 1, 1, 1,
}

var _zcltok_index_offsets []byte = []byte{
	0, 2, 4, 12, 14, 16, 18, 20,
}

var _zcltok_trans_targs []byte = []byte{
	2, 2, 0, 2, 3, 4, 2, 5,
	6, 7, 2, 2, 3, 2, 4, 2,
	2, 2, 0, 2, 1, 2, 2, 2,
	2, 2, 2, 2, 2,
}

var _zcltok_trans_actions []byte = []byte{
	7, 17, 0, 17, 0, 0, 9, 0,
	5, 5, 9, 7, 0, 13, 0, 11,
	7, 15, 0, 15, 0, 15, 17, 17,
	13, 11, 15, 15, 15,
}

var _zcltok_to_state_actions []byte = []byte{
	0, 0, 1, 0, 0, 0, 0, 0,
}

var _zcltok_from_state_actions []byte = []byte{
	0, 0, 3, 0, 0, 0, 0, 0,
}

var _zcltok_eof_trans []byte = []byte{
	24, 24, 0, 25, 26, 29, 29, 29,
}

const zcltok_start int = 2
const zcltok_first_final int = 2
const zcltok_error int = -1

const zcltok_en_main int = 2

// line 13 "scan_tokens.rl"

func scanTokens(data []byte, filename string, start zcl.Pos) []Token {
	f := &tokenAccum{
		Filename: filename,
		Bytes:    data,
		Pos:      start,
	}

	// line 47 "scan_tokens.rl"

	// Ragel state
	cs := 0         // Current State
	p := 0          // "Pointer" into data
	pe := len(data) // End-of-data "pointer"
	ts := 0
	te := 0
	act := 0
	eof := pe

	// Make Go compiler happy
	_ = ts
	_ = te
	_ = act
	_ = eof

	token := func(ty TokenType) {
		f.emitToken(ty, ts, te)
	}

	// line 109 "scan_tokens.go"
	{
		cs = zcltok_start
		ts = 0
		te = 0
		act = 0
	}

	// line 117 "scan_tokens.go"
	{
		var _klen int
		var _trans int
		var _acts int
		var _nacts uint
		var _keys int
		if p == pe {
			goto _test_eof
		}
	_resume:
		_acts = int(_zcltok_from_state_actions[cs])
		_nacts = uint(_zcltok_actions[_acts])
		_acts++
		for ; _nacts > 0; _nacts-- {
			_acts++
			switch _zcltok_actions[_acts-1] {
			case 1:
				// line 1 "NONE"

				ts = p

				// line 138 "scan_tokens.go"
			}
		}

		_keys = int(_zcltok_key_offsets[cs])
		_trans = int(_zcltok_index_offsets[cs])

		_klen = int(_zcltok_single_lengths[cs])
		if _klen > 0 {
			_lower := int(_keys)
			var _mid int
			_upper := int(_keys + _klen - 1)
			for {
				if _upper < _lower {
					break
				}

				_mid = _lower + ((_upper - _lower) >> 1)
				switch {
				case data[p] < _zcltok_trans_keys[_mid]:
					_upper = _mid - 1
				case data[p] > _zcltok_trans_keys[_mid]:
					_lower = _mid + 1
				default:
					_trans += int(_mid - int(_keys))
					goto _match
				}
			}
			_keys += _klen
			_trans += _klen
		}

		_klen = int(_zcltok_range_lengths[cs])
		if _klen > 0 {
			_lower := int(_keys)
			var _mid int
			_upper := int(_keys + (_klen << 1) - 2)
			for {
				if _upper < _lower {
					break
				}

				_mid = _lower + (((_upper - _lower) >> 1) & ^1)
				switch {
				case data[p] < _zcltok_trans_keys[_mid]:
					_upper = _mid - 2
				case data[p] > _zcltok_trans_keys[_mid+1]:
					_lower = _mid + 2
				default:
					_trans += int((_mid - int(_keys)) >> 1)
					goto _match
				}
			}
			_trans += _klen
		}

	_match:
	_eof_trans:
		cs = int(_zcltok_trans_targs[_trans])

		if _zcltok_trans_actions[_trans] == 0 {
			goto _again
		}

		_acts = int(_zcltok_trans_actions[_trans])
		_nacts = uint(_zcltok_actions[_acts])
		_acts++
		for ; _nacts > 0; _nacts-- {
			_acts++
			switch _zcltok_actions[_acts-1] {
			case 2:
				// line 1 "NONE"

				te = p + 1

			case 3:
				// line 43 "scan_tokens.rl"

				te = p + 1
				{
					token(TokenInvalid)
				}
			case 4:
				// line 44 "scan_tokens.rl"

				te = p + 1
				{
					token(TokenBadUTF8)
				}
			case 5:
				// line 41 "scan_tokens.rl"

				te = p
				p--

			case 6:
				// line 42 "scan_tokens.rl"

				te = p
				p--
				{
					token(TokenTabs)
				}
			case 7:
				// line 44 "scan_tokens.rl"

				te = p
				p--
				{
					token(TokenBadUTF8)
				}
			case 8:
				// line 44 "scan_tokens.rl"

				p = (te) - 1
				{
					token(TokenBadUTF8)
				}
				// line 245 "scan_tokens.go"
			}
		}

	_again:
		_acts = int(_zcltok_to_state_actions[cs])
		_nacts = uint(_zcltok_actions[_acts])
		_acts++
		for ; _nacts > 0; _nacts-- {
			_acts++
			switch _zcltok_actions[_acts-1] {
			case 0:
				// line 1 "NONE"

				ts = 0

				// line 260 "scan_tokens.go"
			}
		}

		p++
		if p != pe {
			goto _resume
		}
	_test_eof:
		{
		}
		if p == eof {
			if _zcltok_eof_trans[cs] > 0 {
				_trans = int(_zcltok_eof_trans[cs] - 1)
				goto _eof_trans
			}
		}

	}

	// line 71 "scan_tokens.rl"

	// If we fall out here without being in a final state then we've
	// encountered something that the scanner can't match, which we'll
	// deal with as an invalid.
	if cs < zcltok_first_final {
		f.emitToken(TokenInvalid, p, len(data))
	}

	// We always emit a synthetic EOF token at the end, since it gives the
	// parser position information for an "unexpected EOF" diagnostic.
	f.emitToken(TokenEOF, len(data), len(data))

	return f.Tokens
}
