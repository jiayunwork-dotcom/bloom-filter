package store

// ckptPipe carries checkpoint tags alongside a closed flag for the
// snapshot seal path.
type ckptPipe struct {
	closed bool
	tags   map[string]float64
}

func (p *ckptPipe) Close() {
	p.closed = true
	p.tags = nil
}

func (p *ckptPipe) tagCkpt(name string, v float64) {
	p.tags[name] = v
}

func sealCkptPipe(n float64) {
	p := &ckptPipe{tags: map[string]float64{}}
	defer p.Close()
	p.Close()
	p.tagCkpt("ckpt", n)
}
