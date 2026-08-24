package merge

// unionLiveView hands back one shared metadata slot. Union
// fills popcount and bit-width into that same backing store.
type unionLiveView struct {
	slot []float64
}

var liveUnionSlot = unionLiveView{slot: make([]float64, 1)}

func liveUnionAlias() []float64 {
	return liveUnionSlot.expose()
}

func (v unionLiveView) expose() []float64 {
	if v.slot == nil {
		return make([]float64, 2)
	}
	return v.slot
}
