package rcon

import (
	"bytes"
	"errors"
	"testing"
)

func TestFrameRoundTrips(t *testing.T) {
	sent := frame{
		id:      7,
		typed:   frameCommand,
		payload: "setblock 1 2 3 stone",
	}

	got, err := readFrame(bytes.NewReader(sent.encode()))
	if err != nil {
		t.Fatalf("readFrame: %v", err)
	}

	if got.id != sent.id {
		t.Errorf("id = %d, want %d", got.id, sent.id)
	}
	if got.typed != sent.typed {
		t.Errorf("type = %d, want %d", got.typed, sent.typed)
	}
	if got.payload != sent.payload {
		t.Errorf("payload = %q, want %q", got.payload, sent.payload)
	}
}

func TestFrameLengthCountsTheHeaderAndTerminator(t *testing.T) {
	sent := frame{
		id:      1,
		typed:   frameLogin,
		payload: "secret",
	}

	encoded := sent.encode()

	got := len(encoded)
	want := fieldSize + headerSize + len("secret") + terminator
	if got != want {
		t.Errorf("encoded %d bytes, want %d", got, want)
	}
}

func TestReadFrameRejectsTruncatedLengths(t *testing.T) {
	short := []byte{4, 0, 0, 0, 0, 0, 0, 0}

	_, err := readFrame(bytes.NewReader(short))
	if !errors.Is(err, errTruncated) {
		t.Errorf("err = %v, want %v", err, errTruncated)
	}
}

func TestReadFrameKeepsEmptyPayloads(t *testing.T) {
	sent := frame{
		id:    3,
		typed: frameResponse,
	}

	got, err := readFrame(bytes.NewReader(sent.encode()))
	if err != nil {
		t.Fatalf("readFrame: %v", err)
	}
	if got.payload != "" {
		t.Errorf("payload = %q, want empty", got.payload)
	}
}
