package nbt

import (
	"encoding/binary"
	"fmt"
	"maps"
	"math"
	"slices"
)

func Encode(root Compound) []byte {
	buf := []byte{byte(TagCompound)}

	return encodePayload(buf, root)
}

func EncodeNamed(name string, root Compound) []byte {
	buf := []byte{byte(TagCompound)}
	buf = encodeString(buf, name)

	return encodePayload(buf, root)
}

func encodePayload(buf []byte, tag Tag) []byte {
	switch value := tag.(type) {
	case Byte:
		return append(buf, byte(value))
	case Short:
		return binary.BigEndian.AppendUint16(buf, uint16(value))
	case Int:
		return binary.BigEndian.AppendUint32(buf, uint32(value))
	case Long:
		return binary.BigEndian.AppendUint64(buf, uint64(value))
	case Float:
		return binary.BigEndian.AppendUint32(buf, math.Float32bits(float32(value)))
	case Double:
		return binary.BigEndian.AppendUint64(buf, math.Float64bits(float64(value)))
	case ByteArray:
		buf = binary.BigEndian.AppendUint32(buf, uint32(len(value)))
		return append(buf, value...)
	case String:
		return encodeString(buf, string(value))
	case List:
		return encodeList(buf, value)
	case Compound:
		return encodeCompound(buf, value)
	case IntArray:
		buf = binary.BigEndian.AppendUint32(buf, uint32(len(value)))
		for _, element := range value {
			buf = binary.BigEndian.AppendUint32(buf, uint32(element))
		}
		return buf
	case LongArray:
		buf = binary.BigEndian.AppendUint32(buf, uint32(len(value)))
		for _, element := range value {
			buf = binary.BigEndian.AppendUint64(buf, uint64(element))
		}
		return buf
	}

	return buf
}

func encodeList(buf []byte, list List) []byte {
	elem := list.Elem
	if elem == TagEnd && len(list.Items) > 0 {
		elem = list.Items[0].Type()
	}

	buf = append(buf, byte(elem))
	buf = binary.BigEndian.AppendUint32(buf, uint32(len(list.Items)))

	for _, item := range list.Items {
		buf = encodePayload(buf, item)
	}

	return buf
}

func encodeCompound(buf []byte, compound Compound) []byte {
	for _, name := range slices.Sorted(maps.Keys(compound)) {
		value := compound[name]
		buf = append(buf, byte(value.Type()))
		buf = encodeString(buf, name)
		buf = encodePayload(buf, value)
	}

	return append(buf, byte(TagEnd))
}

func encodeString(buf []byte, s string) []byte {
	start := len(buf)
	buf = append(buf, 0, 0)
	buf = encodeMUTF8(buf, s)

	written := len(buf) - start - 2
	if written > math.MaxUint16 {
		// the length prefix cannot carry it, and a wrapped value would silently
		// corrupt every byte that follows
		panic(fmt.Sprintf("nbt: string of %d bytes exceeds the %d-byte limit", written, math.MaxUint16))
	}

	binary.BigEndian.PutUint16(buf[start:], uint16(written))

	return buf
}
