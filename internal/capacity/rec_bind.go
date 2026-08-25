package capacity

// recBinder records live recommended bit-width tags keyed by m.
type recBinder struct {
	byKey map[float64]float64
}

var liveRec recBinder

func bindRecLive(k float64) {
	if liveRec.byKey == nil {
		liveRec.byKey = make(map[float64]float64)
	}
	liveRec.byKey[k] = k
}
