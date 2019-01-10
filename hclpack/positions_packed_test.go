package hclpack

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestBinaryRoundTrip(t *testing.T) {
	startPacked := positionsPacked{
		{FileIdx: 0, LineDelta: 1, ColumnDelta: 2, ByteDelta: 3},
		{FileIdx: 1, LineDelta: 2, ColumnDelta: 3, ByteDelta: 4},
		{FileIdx: 2, LineDelta: 3, ColumnDelta: 4, ByteDelta: 5},
	}

	b, err := startPacked.MarshalBinary()
	if err != nil {
		t.Fatalf("Failed to marshal: %s", err)
	}

	var endPacked positionsPacked
	err = endPacked.UnmarshalBinary(b)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %s", err)
	}

	if !cmp.Equal(startPacked, endPacked) {
		t.Errorf("Incorrect result\n%s", cmp.Diff(startPacked, endPacked))
	}
}
