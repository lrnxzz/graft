package codec_test

import (
	"bytes"
	"math"
	"net"
	"testing"

	"github.com/lrnxzz/graft/codec"
)

func TestConnPreservesFramesAcrossThresholds(t *testing.T) {
	tests := []struct {
		threshold int
		packet    codec.Frame
	}{
		{
			threshold: -1,
			packet: codec.Frame{
				ID:      0x00,
				Payload: codec.Marshal(codec.VarInt(765), codec.String("mc.local"), codec.UShort(25565), codec.VarInt(1)),
			},
		},
		{
			threshold: -1,
			packet: codec.Frame{
				ID: 0x01,
			},
		},
		{
			threshold: 64,
			packet: codec.Frame{
				ID:      0x02,
				Payload: codec.Marshal(codec.String("below threshold")),
			},
		},
		{
			threshold: 16,
			packet: codec.Frame{
				ID:      0x03,
				Payload: bytes.Repeat(codec.Marshal(codec.String("chunk data")), 256),
			},
		},
	}

	for _, tt := range tests {
		client, server := net.Pipe()

		in := codec.NewConn(client)
		out := codec.NewConn(server)
		in.SetThreshold(tt.threshold)
		out.SetThreshold(tt.threshold)

		errs := make(chan error, 1)
		go func() {
			errs <- in.WriteFrame(tt.packet)
		}()

		got, err := out.ReadFrame()

		if err != nil {
			t.Errorf("ReadFrame (threshold %d): %v", tt.threshold, err)
		} else if got.ID != tt.packet.ID || !bytes.Equal(got.Payload, tt.packet.Payload) {
			t.Errorf("packet 0x%02x round trip yielded id 0x%02x with %d bytes, want %d bytes",
				tt.packet.ID, got.ID, len(got.Payload), len(tt.packet.Payload))
		}

		if err := <-errs; err != nil {
			t.Errorf("WriteFrame (threshold %d): %v", tt.threshold, err)
		}

		client.Close()
		server.Close()
	}
}

func TestConnRejectsMalformedFrame(t *testing.T) {
	tests := []struct {
		frame []byte
	}{
		{
			frame: codec.AppendVar(nil, int32(0)),
		},
		{
			frame: codec.AppendVar(nil, int32(-1)),
		},
		{
			frame: codec.AppendVar(nil, int32(math.MaxInt32)),
		},
	}

	for _, tt := range tests {
		client, server := net.Pipe()

		out := codec.NewConn(server)

		go func() {
			client.Write(tt.frame)
		}()

		if _, err := out.ReadFrame(); err == nil {
			t.Errorf("ReadFrame(frame %x): expected an error, got nil", tt.frame)
		}

		client.Close()
		server.Close()
	}
}
