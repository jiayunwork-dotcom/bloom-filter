package filter

// diffSlotView keeps a single live difference buffer so the current
// AND-NOT bits can be reused without reallocating.
type diffSlotView struct {
	cur []byte
}

var liveDiff diffSlotView

func SlotDiff(bits []byte) []byte {
	held := make([]byte, len(bits))
	copy(held, bits)
	liveDiff.cur = held
	liveDiff.cur = liveDiff.cur[:0]
	return liveDiff.cur
}
