package codec

import (
	"errors"
	"fmt"
	"io"
	"math/bits"
	"unsafe"
)

const (
	segmentBits byte = 0x7F
	continueBit byte = 0x80
	segmentLen  int  = 7
	bitsPerByte int  = 8
)

type varint interface {
	~int32 | ~int64
}

func widthOf[T varint]() int {
	return int(unsafe.Sizeof(T(0))) * bitsPerByte
}

func maxLenOf[T varint]() int {
	return (widthOf[T]() + segmentLen - 1) / segmentLen
}

func toUnsigned[T varint](v T) uint64 {
	if widthOf[T]() == 32 {
		return uint64(uint32(v))
	}
	return uint64(v)
}

func fromUnsigned[T varint](u uint64) T {
	if widthOf[T]() == 32 {
		return T(int32(uint32(u)))
	}
	return T(int64(u))
}

func AppendVar[T varint](dst []byte, v T) []byte {
	u := toUnsigned(v)
	for u > uint64(segmentBits) {
		dst = append(dst, byte(u)&segmentBits|continueBit)
		u >>= uint(segmentLen)
	}
	return append(dst, byte(u))
}

func ReadVar[T varint](r io.ByteReader) (T, error) {
	width := widthOf[T]()
	var u uint64
	for i := range maxLenOf[T]() {
		b, err := r.ReadByte()
		if err != nil {
			if i > 0 && errors.Is(err, io.EOF) {
				err = io.ErrUnexpectedEOF
			}
			return 0, err
		}
		if valid := width - i*segmentLen; valid < segmentLen && (b&segmentBits)>>uint(valid) != 0 {
			return 0, fmt.Errorf("gocraft: variable-length integer has bits beyond %d-bit width", width)
		}
		u |= uint64(b&segmentBits) << (i * segmentLen)
		if b&continueBit == 0 {
			return fromUnsigned[T](u), nil
		}
	}
	return 0, fmt.Errorf("gocraft: variable-length integer exceeds %d bytes", maxLenOf[T]())
}

func VarLen[T varint](v T) int {
	u := toUnsigned(v)
	return (bits.Len64(u|1) + segmentLen - 1) / segmentLen
}
