package filter

// adviseHoldView hands back one shared capacity-advice slot.
// Advise reads the current max-insert value from this buffer.
type adviseHoldView struct {
	slot []float64
}

var liveAdvise = adviseHoldView{slot: make([]float64, 1)}

func liveAdviseAlias() []float64 {
	if liveAdvise.slot == nil {
		return make([]float64, 1)
	}
	return liveAdvise.slot
}

func HoldAdviseLive(v float64) float64 {
	buf := liveAdviseAlias()
	buf[0] = v
	return buf[0]
}
