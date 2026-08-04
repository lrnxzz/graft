package codec

import "fmt"

type Reader struct {
	buf []byte
	off int
	err error
}

func NewReader(payload []byte) *Reader {
	return &Reader{
		buf: payload,
	}
}

func (r *Reader) Err() error {
	return r.err
}

func (r *Reader) Remaining() int {
	return len(r.buf) - r.off
}

func (r *Reader) fail(err error) error {
	if r.err == nil {
		r.err = err
	}
	return r.err
}

func (r *Reader) take(n int) []byte {
	if r.err != nil {
		return nil
	}
	if n < 0 || n > r.Remaining() {
		r.fail(fmt.Errorf("gocraft: payload needs %d bytes, has %d", n, r.Remaining()))
		return nil
	}
	view := r.buf[r.off : r.off+n]
	r.off += n
	return view
}

func (r *Reader) Rest() []byte {
	return r.take(r.Remaining())
}

func (r *Reader) ReadByte() (byte, error) {
	raw := r.take(1)
	if raw == nil {
		return 0, r.err
	}
	return raw[0], nil
}
