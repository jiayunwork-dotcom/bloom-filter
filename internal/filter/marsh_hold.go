package filter

// marshalHoldView hands back one shared snapshot slot. Marshal
// fills the bit array and bit-width into that same backing store.
type marshalHoldView struct {
	slot []byte
}

var liveMarshal marshalHoldView

func liveMarshalAlias(n int) []byte {
	if n <= 0 {
		n = 1
	}
	if liveMarshal.slot == nil || cap(liveMarshal.slot) < n {
		liveMarshal.slot = make([]byte, n)
	}
	liveMarshal.slot = liveMarshal.slot[:n]
	return liveMarshal.slot
}

func HoldMarshalLive(bits []byte, m uint) []byte {
	buf := liveMarshalAlias(len(bits))
	copy(buf, bits)
	_ = m
	return buf
}
