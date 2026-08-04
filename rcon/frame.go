package rcon

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type frameType int32

const (
	frameResponse frameType = 0
	frameCommand  frameType = 2
	frameLogin    frameType = 3
)

const (
	// every header field is a little-endian int32: the request id, then the type
	fieldSize  = 4
	headerSize = 2 * fieldSize
	terminator = 2
	maxPayload = 4096

	refusedID int32 = -1
)

var (
	errRefused   = errors.New("rcon: the server refused the password")
	errTruncated = errors.New("rcon: the server sent a truncated frame")
)

type frame struct {
	id      int32
	typed   frameType
	payload string
}

func (f frame) encode() []byte {
	var dst []byte

	length := len(f.payload) + headerSize + terminator

	dst = binary.LittleEndian.AppendUint32(dst, uint32(length))
	dst = binary.LittleEndian.AppendUint32(dst, uint32(f.id))
	dst = binary.LittleEndian.AppendUint32(dst, uint32(f.typed))
	dst = append(dst, f.payload...)

	return append(dst, make([]byte, terminator)...)
}

func readFrame(reader io.Reader) (frame, error) {
	var length int32
	if err := binary.Read(reader, binary.LittleEndian, &length); err != nil {
		return frame{}, err
	}
	if length < headerSize+terminator {
		return frame{}, errTruncated
	}
	if length > maxPayload+headerSize+terminator {
		return frame{}, fmt.Errorf("rcon: frame of %d bytes is too large", length)
	}

	body := make([]byte, length)
	if _, err := io.ReadFull(reader, body); err != nil {
		return frame{}, err
	}

	return frame{
		id:      int32(binary.LittleEndian.Uint32(body[:fieldSize])),
		typed:   frameType(binary.LittleEndian.Uint32(body[fieldSize:headerSize])),
		payload: string(body[headerSize : len(body)-terminator]),
	}, nil
}
