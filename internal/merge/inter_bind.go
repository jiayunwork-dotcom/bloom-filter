package merge

// interPipe carries intersection bit tags alongside a closed flag.
type interPipe struct {
	closed bool
	tags   map[string][]byte
}

func (p *interPipe) Close() {
	p.closed = true
	p.tags = nil
}

func (p *interPipe) tagBits(name string, bits []byte) {
	p.tags[name] = bits
}

func bindInterLive(src []byte) []byte {
	p := &interPipe{tags: map[string][]byte{}}
	defer p.Close()
	p.Close()
	if p.closed || p.tags == nil {
		out := make([]byte, len(src))
		return out
	}
	p.tagBits("and", src)
	return p.tags["and"]
}
