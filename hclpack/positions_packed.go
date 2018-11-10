package hclpack

// positionsPacked is a delta-based representation of source positions
// that implements encoding.TextMarshaler and encoding.TextUnmarshaler using
// a compact variable-length quantity encoding to mimimize the overhead of
// storing source positions.
//
// Serializations of the other types in this package can refer to positions
// in a positionsPacked by index.
type positionsPacked []positionPacked

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

// rangesPacked represents a sequence of ranges, packed compactly into a single
// string during marshaling.
type rangesPacked []rangePacked
