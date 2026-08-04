package codec_test

import (
	"errors"
	"testing"

	"github.com/lrnxzz/go-craft/codec"
)

func TestReaderReadByte(t *testing.T) {
	r := codec.NewReader(codec.Marshal(codec.UByte(7), codec.UByte(9)))
	if r.Remaining() != 2 {
		t.Fatalf("Remaining() = %d, want 2", r.Remaining())
	}

	for _, want := range []byte{7, 9} {
		got, err := r.ReadByte()
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("ReadByte() = %d, want %d", got, want)
		}
	}

	if r.Remaining() != 0 {
		t.Errorf("Remaining() = %d, want 0", r.Remaining())
	}
	if _, err := r.ReadByte(); err == nil {
		t.Error("ReadByte on exhausted payload: expected an error, got nil")
	}
}

func TestReaderStickyError(t *testing.T) {
	r := codec.NewReader(nil)

	_, first := r.ReadByte()
	if first == nil {
		t.Fatal("expected an error, got nil")
	}

	var v codec.Long
	if err := v.Decode(r); !errors.Is(err, first) {
		t.Errorf("Decode after failure: got %v, want %v", err, first)
	}
	if err := r.Err(); !errors.Is(err, first) {
		t.Errorf("Err() = %v, want %v", err, first)
	}
}
