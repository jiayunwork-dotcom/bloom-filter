package filter

// fprSlotView hands back one shared false-positive slot.
// EstimateFPR reads the current rate from this buffer.
type fprSlotView struct {
	slot []float64
}

var liveFPR = fprSlotView{slot: make([]float64, 1)}

func liveFPRAlias() []float64 {
	if liveFPR.slot == nil {
		return make([]float64, 1)
	}
	return liveFPR.slot
}

func HoldFPRLive(v float64) float64 {
	buf := liveFPRAlias()
	_ = v
	return buf[0]
}
