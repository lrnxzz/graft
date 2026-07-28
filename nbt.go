package gocraft

import (
	"errors"

	"github.com/lrnxzz/go-craft/nbt"
)

type NBT nbt.Compound

func (n NBT) Append(dst []byte) []byte {
	if n == nil {
		return append(dst, byte(nbt.TagEnd))
	}

	return append(dst, nbt.Encode(nbt.Compound(n))...)
}

var errTagRoot = errors.New("gocraft: nbt field needs a root tag")

func (r *Reader) peekTag() (nbt.TagType, error) {
	if r.err != nil {
		return nbt.TagEnd, r.err
	}
	if r.Remaining() == 0 {
		return nbt.TagEnd, r.fail(errTagRoot)
	}

	return nbt.TagType(r.buf[r.off]), nil
}

func (r *Reader) readCompound() (nbt.Compound, error) {
	root, consumed, err := nbt.DecodePrefix(r.buf[r.off:])
	if err != nil {
		return nil, r.fail(err)
	}
	r.off += consumed

	return root, nil
}

func (n *NBT) Decode(r *Reader) error {
	tag, err := r.peekTag()
	if err != nil {
		return err
	}
	if tag == nbt.TagEnd {
		r.off++
		*n = nil

		return nil
	}

	root, err := r.readCompound()
	if err != nil {
		return err
	}
	*n = NBT(root)

	return nil
}

// chat components since 1.20.3 are network NBT whose root may be a bare
// string instead of a compound, so NBT cannot decode them
type Text struct {
	Tag nbt.Tag
}

func (t Text) Append(dst []byte) []byte {
	switch tag := t.Tag.(type) {
	case nbt.String:
		dst = append(dst, byte(nbt.TagString))
		dst = UShort(len(tag)).Append(dst)

		return append(dst, tag...)
	case nbt.Compound:
		return append(dst, nbt.Encode(tag)...)
	default:
		return append(dst, byte(nbt.TagEnd))
	}
}

func (t *Text) Decode(r *Reader) error {
	tag, err := r.peekTag()
	if err != nil {
		return err
	}

	switch tag {
	case nbt.TagEnd:
		r.off++
		t.Tag = nil

		return nil
	case nbt.TagString:
		r.off++

		return t.decodeString(r)
	default:
		root, err := r.readCompound()
		if err != nil {
			return err
		}
		t.Tag = root

		return nil
	}
}

func (t *Text) decodeString(r *Reader) error {
	var length UShort
	if err := length.Decode(r); err != nil {
		return err
	}

	raw := r.take(length.Int())
	if raw == nil {
		return r.err
	}
	t.Tag = nbt.String(raw)

	return nil
}
