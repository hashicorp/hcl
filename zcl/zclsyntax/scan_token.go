// line 1 "scan_token.rl"
package zclsyntax

import (
	"github.com/zclconf/go-zcl/zcl"
)

// This file is generated from scan_token.rl. DO NOT EDIT.

// line 12 "scan_token.go"
var _zcltok_actions []byte = []byte{
	0, 1, 0, 1, 1, 1, 2, 1, 3,
	1, 4, 1, 5, 1, 6, 1, 7,
}

var _zcltok_key_offsets []byte = []byte{
	0, 0, 2, 4, 5, 15, 17, 19,
}

var _zcltok_trans_keys []byte = []byte{
	128, 191, 128, 191, 32, 128, 191, 192,
	223, 224, 239, 240, 247, 248, 255, 128,
	191, 128, 191, 128, 191,
}

var _zcltok_single_lengths []byte = []byte{
	0, 0, 0, 1, 0, 0, 0, 0,
}

var _zcltok_range_lengths []byte = []byte{
	0, 1, 1, 0, 5, 1, 1, 1,
}

var _zcltok_index_offsets []byte = []byte{
	0, 0, 2, 4, 6, 12, 14, 16,
}

var _zcltok_trans_targs []byte = []byte{
	4, 4, 1, 4, 3, 0, 4, 5,
	6, 7, 4, 4, 4, 4, 1, 4,
	2, 4, 4, 4, 4, 4, 4,
}

var _zcltok_trans_actions []byte = []byte{
	9, 15, 0, 15, 1, 0, 11, 0,
	7, 7, 11, 9, 9, 13, 0, 13,
	0, 13, 15, 15, 13, 13, 13,
}

var _zcltok_to_state_actions []byte = []byte{
	0, 0, 0, 3, 3, 0, 0, 0,
}

var _zcltok_from_state_actions []byte = []byte{
	0, 0, 0, 0, 5, 0, 0, 0,
}

var _zcltok_eof_trans []byte = []byte{
	0, 20, 20, 0, 0, 23, 23, 23,
}

const zcltok_start int = 3
const zcltok_first_final int = 3
const zcltok_error int = 0

const zcltok_en_token int = 4
const zcltok_en_main int = 3

// line 13 "scan_token.rl"

func nextToken(data []byte, filename string, start zcl.Pos) (Token, []byte) {
	offset := 0

	f := tokenFactory{
		Filename: filename,
		Bytes:    data,
		Start:    start,
	}

	// line 69 "scan_token.rl"

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

	// line 104 "scan_token.go"
	{
		cs = zcltok_start
		ts = 0
		te = 0
		act = 0
	}

	// line 112 "scan_token.go"
	{
		var _klen int
		var _trans int
		var _acts int
		var _nacts uint
		var _keys int
		if p == pe {
			goto _test_eof
		}
		if cs == 0 {
			goto _out
		}
	_resume:
		_acts = int(_zcltok_from_state_actions[cs])
		_nacts = uint(_zcltok_actions[_acts])
		_acts++
		for ; _nacts > 0; _nacts-- {
			_acts++
			switch _zcltok_actions[_acts-1] {
			case 2:
				// line 1 "NONE"

				ts = p

				// line 136 "scan_token.go"
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
			case 0:
				// line 25 "scan_token.rl"

				offset = p
				cs = 4
				goto _again

			case 3:
				// line 1 "NONE"

				te = p + 1

			case 4:
				// line 30 "scan_token.rl"

				te = p + 1
				{
					return f.makeToken(TokenInvalid, offset, p+1)
				}
			case 5:
				// line 34 "scan_token.rl"

				te = p + 1
				{
					return f.makeToken(TokenBadUTF8, offset, p+1)
				}
			case 6:
				// line 34 "scan_token.rl"

				te = p
				p--
				{
					return f.makeToken(TokenBadUTF8, offset, p+1)
				}
			case 7:
				// line 34 "scan_token.rl"

				p = (te) - 1
				{
					return f.makeToken(TokenBadUTF8, offset, p+1)
				}
				// line 248 "scan_token.go"
			}
		}

	_again:
		_acts = int(_zcltok_to_state_actions[cs])
		_nacts = uint(_zcltok_actions[_acts])
		_acts++
		for ; _nacts > 0; _nacts-- {
			_acts++
			switch _zcltok_actions[_acts-1] {
			case 1:
				// line 1 "NONE"

				ts = 0

				// line 263 "scan_token.go"
			}
		}

		if cs == 0 {
			goto _out
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

	_out:
		{
		}
	}

	// line 89 "scan_token.rl"

	// If we fall out here then we'll just classify the remainder of the
	// file as invalid.
	return f.makeToken(TokenInvalid, 0, len(data))
}
