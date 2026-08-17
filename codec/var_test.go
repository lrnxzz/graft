package codec_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"testing"
	"unsafe"

	"github.com/lrnxzz/graft/codec"
)

type varintDomain interface {
	~int32 | ~int64
}

func boundaries[T varintDomain]() []T {
	width := int(unsafe.Sizeof(T(0))) * 8
	values := []T{0, 1, -1, 2, -2}

	for shift := 7; shift < width-1; shift += 7 {
		base := T(1) << shift
		values = append(values, base-1, base, base+1, -base-1, -base, -base+1)
	}

	maxValue := T(1)<<(width-1) - 1

	return append(values, maxValue, maxValue-1, -maxValue-1, -maxValue)
}

func stdlibLeb128[T varintDomain](v T) []byte {
	if unsafe.Sizeof(v) == 4 {
		return binary.AppendUvarint(nil, uint64(uint32(v)))
	}

	return binary.AppendUvarint(nil, uint64(v))
}

func TestAppendVarAgainstStdlib(t *testing.T) {
	for _, value := range boundaries[int32]() {
		got := codec.AppendVar(nil, value)
		want := stdlibLeb128(value)

		if !bytes.Equal(got, want) {
			t.Errorf("AppendVar(int32 %d) = %x, want %x", value, got, want)
		}
	}

	for _, value := range boundaries[int64]() {
		got := codec.AppendVar(nil, value)
		want := stdlibLeb128(value)

		if !bytes.Equal(got, want) {
			t.Errorf("AppendVar(int64 %d) = %x, want %x", value, got, want)
		}
	}
}

func TestReadVarRecoversEncodedValue(t *testing.T) {
	for _, value := range boundaries[int32]() {
		got, err := codec.ReadVar[int32](bytes.NewReader(codec.AppendVar(nil, value)))

		if err != nil {
			t.Errorf("ReadVar(codec.AppendVar(int32 %d)): %v", value, err)
			continue
		}
		if got != value {
			t.Errorf("round trip of int32 %d yielded %d", value, got)
		}
	}

	for _, value := range boundaries[int64]() {
		got, err := codec.ReadVar[int64](bytes.NewReader(codec.AppendVar(nil, value)))

		if err != nil {
			t.Errorf("ReadVar(codec.AppendVar(int64 %d)): %v", value, err)
			continue
		}
		if got != value {
			t.Errorf("round trip of int64 %d yielded %d", value, got)
		}
	}
}

func TestVarLenAgainstEncoding(t *testing.T) {
	for _, value := range boundaries[int32]() {
		got := codec.VarLen(value)
		want := len(codec.AppendVar(nil, value))

		if got != want {
			t.Errorf("VarLen(int32 %d) = %d, want %d", value, got, want)
		}
	}

	for _, value := range boundaries[int64]() {
		got := codec.VarLen(value)
		want := len(codec.AppendVar(nil, value))

		if got != want {
			t.Errorf("VarLen(int64 %d) = %d, want %d", value, got, want)
		}
	}
}

func TestReadVarOnMalformedInput(t *testing.T) {
	unterminated := bytes.Repeat(codec.AppendVar(nil, int32(math.MinInt32))[:1], 16)

	tests := []struct {
		input   []byte
		wantErr error
	}{
		{
			input:   nil,
			wantErr: io.EOF,
		},
		{
			input:   codec.AppendVar(nil, int32(math.MaxInt32))[:2],
			wantErr: io.ErrUnexpectedEOF,
		},
		{
			input: unterminated,
		},
		{
			input: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0x7F},
		},
	}

	for _, tt := range tests {
		_, err := codec.ReadVar[int32](bytes.NewReader(tt.input))

		if err == nil {
			t.Errorf("ReadVar(%x): expected an error, got nil", tt.input)
			continue
		}
		if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
			t.Errorf("ReadVar(%x) = %v, want %v", tt.input, err, tt.wantErr)
		}
	}

	if _, err := codec.ReadVar[int64](bytes.NewReader(unterminated)); err == nil {
		t.Error("ReadVar[int64](unterminated): expected an error, got nil")
	}

	overlong := append(bytes.Repeat([]byte{0xFF}, 9), 0x02)
	if _, err := codec.ReadVar[int64](bytes.NewReader(overlong)); err == nil {
		t.Error("ReadVar[int64](overlong): expected an error, got nil")
	}
}
