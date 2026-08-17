package codec

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync/atomic"
	"time"
)

const (
	maxFrameLen    = 1<<21 - 1
	maxInflatedLen = 1 << 23
	noCompression  = -1
)

type Frame struct {
	ID      VarInt
	Payload []byte
}

type Conn struct {
	transport net.Conn
	reader    *bufio.Reader
	threshold atomic.Int64
}

func Dial(ctx context.Context, address string) (*Conn, error) {
	var dialer net.Dialer

	transport, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, err
	}

	slog.Debug("connected", "address", address)

	return NewConn(transport), nil
}

func NewConn(transport net.Conn) *Conn {
	c := &Conn{
		transport: transport,
		reader:    bufio.NewReader(transport),
	}
	c.threshold.Store(noCompression)

	return c
}

func (c *Conn) SetThreshold(threshold int) {
	c.threshold.Store(int64(threshold))
}

func (c *Conn) SetDeadline(deadline time.Time) error {
	return c.transport.SetDeadline(deadline)
}

func (c *Conn) Close() error {
	return c.transport.Close()
}

func (c *Conn) WriteFrame(frame Frame) error {
	body := frame.ID.Append(nil)
	body = append(body, frame.Payload...)

	encoded, err := c.frame(body)
	if err != nil {
		return err
	}

	_, err = c.transport.Write(encoded)

	return err
}

func (c *Conn) ReadFrame() (Frame, error) {
	frameLen, err := ReadVar[VarInt](c.reader)
	if err != nil {
		return Frame{}, err
	}
	if frameLen <= 0 || frameLen > maxFrameLen {
		return Frame{}, fmt.Errorf("graft: frame of %d bytes is out of range", frameLen)
	}

	frame := make([]byte, frameLen)
	if _, err := io.ReadFull(c.reader, frame); err != nil {
		return Frame{}, err
	}

	body := frame
	if c.threshold.Load() >= 0 {
		if body, err = c.inflate(frame); err != nil {
			return Frame{}, err
		}
	}

	r := NewReader(body)

	var id VarInt
	if err := id.Decode(r); err != nil {
		return Frame{}, err
	}

	return Frame{
		ID:      id,
		Payload: r.Rest(),
	}, nil
}

func (c *Conn) frame(body []byte) ([]byte, error) {
	threshold := int(c.threshold.Load())
	if threshold < 0 {
		frame := AppendVar(nil, VarInt(len(body)))

		return append(frame, body...), nil
	}

	if len(body) < threshold {
		marker := AppendVar(nil, VarInt(0))
		frame := AppendVar(nil, VarInt(len(marker)+len(body)))
		frame = append(frame, marker...)

		return append(frame, body...), nil
	}

	var compressed bytes.Buffer

	zw := zlib.NewWriter(&compressed)
	if _, err := zw.Write(body); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}

	inner := AppendVar(nil, VarInt(len(body)))
	inner = append(inner, compressed.Bytes()...)

	frame := AppendVar(nil, VarInt(len(inner)))

	return append(frame, inner...), nil
}

func (c *Conn) inflate(frame []byte) ([]byte, error) {
	r := NewReader(frame)

	var inflatedLen VarInt
	if err := inflatedLen.Decode(r); err != nil {
		return nil, err
	}

	compressed := r.Rest()
	if inflatedLen == 0 {
		return compressed, nil
	}
	if inflatedLen < 0 || inflatedLen > maxInflatedLen {
		return nil, fmt.Errorf("graft: inflated frame of %d bytes is out of range", inflatedLen)
	}

	zr, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	body := make([]byte, inflatedLen)
	if _, err := io.ReadFull(zr, body); err != nil {
		return nil, err
	}

	return body, nil
}
