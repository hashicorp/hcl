package hclpack

import (
	"encoding/base64"
	"sort"

	"github.com/hashicorp/hcl2/hcl"
)

// positionsPacked is a delta-based representation of source positions
// that implements encoding.TextMarshaler and encoding.TextUnmarshaler using
// a compact variable-length quantity encoding to mimimize the overhead of
// storing source positions.
//
// Serializations of the other types in this package can refer to positions
// in a positionsPacked by index.
type positionsPacked []positionPacked

func (pp positionsPacked) MarshalBinary() ([]byte, error) {
	lenInt := len(pp) * 4 // each positionPacked contains four ints, but we don't include the fileidx

	// guess avg of ~1.25 bytes per int, in which case we'll avoid further allocation
	buf := newVLQBuf(lenInt + (lenInt / 4))
	var lastFileIdx int
	for _, ppr := range pp {
		// Rather than writing out the same file index over and over, we instead
		// insert a ; delimiter each time it increases. Since it's common for
		// for a body to be entirely in one file, this can lead to considerable
		// savings in that case.
		delims := ppr.FileIdx - lastFileIdx
		lastFileIdx = ppr.FileIdx
		for i := 0; i < delims; i++ {
			buf = buf.AppendRawByte(';')
		}
		buf = buf.AppendInt(ppr.LineDelta)
		buf = buf.AppendInt(ppr.ColumnDelta)
		buf = buf.AppendInt(ppr.ByteDelta)
	}

	return buf.Bytes(), nil
}

func (pp positionsPacked) MarshalText() ([]byte, error) {
	raw, err := pp.MarshalBinary()
	if err != nil {
		return nil, err
	}

	l := base64.RawStdEncoding.EncodedLen(len(raw))
	ret := make([]byte, l)
	base64.RawStdEncoding.Encode(ret, raw)
	return ret, nil
}

func (pp *positionsPacked) UnmarshalBinary(data []byte) error {
	buf := vlqBuf(data)
	var ret positionsPacked
	fileIdx := 0
	for len(buf) > 0 {
		if buf[0] == ';' {
			// Starting a new file, then.
			fileIdx++
			buf = buf[1:]
			continue
		}

		var ppr positionPacked
		var err error
		ppr.FileIdx = fileIdx
		ppr.LineDelta, buf, err = buf.ReadInt()
		if err != nil {
			return err
		}
		ppr.ColumnDelta, buf, err = buf.ReadInt()
		if err != nil {
			return err
		}
		ppr.ByteDelta, buf, err = buf.ReadInt()
		if err != nil {
			return err
		}
		ret = append(ret, ppr)
	}
	*pp = ret
	return nil
}

func (pp *positionsPacked) UnmarshalText(data []byte) error {
	maxL := base64.RawStdEncoding.DecodedLen(len(data))
	into := make([]byte, maxL)
	realL, err := base64.RawStdEncoding.Decode(into, data)
	if err != nil {
		return err
	}
	return pp.UnmarshalBinary(into[:realL])
}

type position struct {
	FileIdx int
	Pos     hcl.Pos
}

func (pp positionsPacked) Unpack() []position {
	ret := make([]position, len(pp))
	var accPos hcl.Pos
	var accFileIdx int

	for i, relPos := range pp {
		if relPos.FileIdx != accFileIdx {
			accPos = hcl.Pos{} // reset base position for each new file
			accFileIdx = pp[i].FileIdx
		}
		if relPos.LineDelta > 0 {
			accPos.Column = 0 // reset column position for each new line
		}
		accPos.Line += relPos.LineDelta
		accPos.Column += relPos.ColumnDelta
		accPos.Byte += relPos.ByteDelta
		ret[i] = position{
			FileIdx: relPos.FileIdx,
			Pos:     accPos,
		}
	}

	return ret
}

type positionPacked struct {
	FileIdx                           int
	LineDelta, ColumnDelta, ByteDelta int
}

func (pp positionsPacked) Len() int {
	return len(pp)
}

func (pp positionsPacked) Less(i, j int) bool {
	return pp[i].FileIdx < pp[j].FileIdx
}

func (pp positionsPacked) Swap(i, j int) {
	pp[i], pp[j] = pp[j], pp[i]
}

// posOfs is an index into a positionsPacked. The zero value of this type
// represents the absense of a position.
type posOfs int

func newPosOffs(idx int) posOfs {
	return posOfs(idx + 1)
}

func (o posOfs) Index() int {
	return int(o - 1)
}

// rangePacked is a range represented as two indexes into a positionsPacked.
// This implements encoding.TextMarshaler and encoding.TextUnmarshaler using
// a compact variable-length quantity encoding.
type rangePacked struct {
	Start posOfs
	End   posOfs
}

func packRange(rng hcl.Range, pos map[string]map[hcl.Pos]posOfs) rangePacked {
	return rangePacked{
		Start: pos[rng.Filename][rng.Start],
		End:   pos[rng.Filename][rng.End],
	}
}

func (rp rangePacked) Unpack(fns []string, poss []position) hcl.Range {
	startIdx := rp.Start.Index()
	endIdx := rp.End.Index()
	if startIdx < 0 && startIdx >= len(poss) {
		return hcl.Range{} // out of bounds, so invalid
	}
	if endIdx < 0 && endIdx >= len(poss) {
		return hcl.Range{} // out of bounds, so invalid
	}
	startPos := poss[startIdx]
	endPos := poss[endIdx]
	fnIdx := startPos.FileIdx
	var fn string
	if fnIdx >= 0 && fnIdx < len(fns) {
		fn = fns[fnIdx]
	}
	return hcl.Range{
		Filename: fn,
		Start:    startPos.Pos,
		End:      endPos.Pos,
	}
}

// rangesPacked represents a sequence of ranges, packed compactly into a single
// string during marshaling.
type rangesPacked []rangePacked

func (rp rangesPacked) MarshalBinary() ([]byte, error) {
	lenInt := len(rp) * 2 // each positionPacked contains two ints

	// guess avg of ~1.25 bytes per int, in which case we'll avoid further allocation
	buf := newVLQBuf(lenInt + (lenInt / 4))
	for _, rpr := range rp {
		buf = buf.AppendInt(int(rpr.Start)) // intentionally storing these as 1-based offsets
		buf = buf.AppendInt(int(rpr.End))
	}

	return buf.Bytes(), nil
}

func (rp rangesPacked) MarshalText() ([]byte, error) {
	raw, err := rp.MarshalBinary()
	if err != nil {
		return nil, err
	}

	l := base64.RawStdEncoding.EncodedLen(len(raw))
	ret := make([]byte, l)
	base64.RawStdEncoding.Encode(ret, raw)
	return ret, nil
}

func (rp *rangesPacked) UnmarshalBinary(data []byte) error {
	buf := vlqBuf(data)
	var ret rangesPacked
	for len(buf) > 0 {
		var startInt, endInt int
		var err error
		startInt, buf, err = buf.ReadInt()
		if err != nil {
			return err
		}
		endInt, buf, err = buf.ReadInt()
		if err != nil {
			return err
		}
		ret = append(ret, rangePacked{
			Start: posOfs(startInt), // these are stored as 1-based offsets, so safe to convert directly
			End:   posOfs(endInt),
		})
	}
	*rp = ret
	return nil
}

func (rp *rangesPacked) UnmarshalText(data []byte) error {
	maxL := base64.RawStdEncoding.DecodedLen(len(data))
	into := make([]byte, maxL)
	realL, err := base64.RawStdEncoding.Decode(into, data)
	if err != nil {
		return err
	}
	return rp.UnmarshalBinary(into[:realL])
}

func (rps rangesPacked) UnpackIdx(fns []string, poss []position, idx int) hcl.Range {
	if idx < 0 || idx >= len(rps) {
		return hcl.Range{} // out of bounds, so invalid
	}
	return rps[idx].Unpack(fns, poss)
}

// packPositions will find the distinct positions from the given ranges
// and then pack them into a positionsPacked, along with a lookup table to find
// the encoded offset of each distinct position.
func packPositions(rngs map[hcl.Range]struct{}) (fns []string, poss positionsPacked, posMap map[string]map[hcl.Pos]posOfs) {
	const noOfs = posOfs(0)

	posByFile := make(map[string][]hcl.Pos)
	for rng := range rngs {
		fn := rng.Filename
		posByFile[fn] = append(posByFile[fn], rng.Start)
		posByFile[fn] = append(posByFile[fn], rng.End)
	}
	fns = make([]string, 0, len(posByFile))
	for fn := range posByFile {
		fns = append(fns, fn)
	}
	sort.Strings(fns)

	var retPos positionsPacked
	posMap = make(map[string]map[hcl.Pos]posOfs)
	for fileIdx, fn := range fns {
		poss := posByFile[fn]
		sort.Sort(sortPositions(poss))
		var prev hcl.Pos
		for _, pos := range poss {
			if _, exists := posMap[fn][pos]; exists {
				continue
			}
			ofs := newPosOffs(len(retPos))
			if pos.Line != prev.Line {
				// Column indices start from zero for each new line.
				prev.Column = 0
			}
			retPos = append(retPos, positionPacked{
				FileIdx:     fileIdx,
				LineDelta:   pos.Line - prev.Line,
				ColumnDelta: pos.Column - prev.Column,
				ByteDelta:   pos.Byte - prev.Byte,
			})
			if posMap[fn] == nil {
				posMap[fn] = make(map[hcl.Pos]posOfs)
			}
			posMap[fn][pos] = ofs
			prev = pos
		}
	}

	return fns, retPos, posMap
}

type sortPositions []hcl.Pos

func (sp sortPositions) Len() int {
	return len(sp)
}

func (sp sortPositions) Less(i, j int) bool {
	return sp[i].Byte < sp[j].Byte
}

func (sp sortPositions) Swap(i, j int) {
	sp[i], sp[j] = sp[j], sp[i]
}
