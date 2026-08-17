package codec_test

import (
	"bytes"
	"encoding/binary"
	"math"
	"reflect"
	"testing"

	"github.com/lrnxzz/graft/codec"
)

type profileEntry struct {
	Name codec.String
	ID   codec.UUID
}

func (p profileEntry) Append(dst []byte) []byte {
	dst = p.Name.Append(dst)

	return p.ID.Append(dst)
}

func (p *profileEntry) Decode(r *codec.Reader) error {
	if err := p.Name.Decode(r); err != nil {
		return err
	}

	return p.ID.Decode(r)
}

func TestFieldTypesSurviveEncoding(t *testing.T) {
	var id codec.UUID
	for i := range id {
		id[i] = byte(i)
	}

	fields := []codec.Field{
		codec.Bool(true),
		codec.Bool(false),
		codec.Byte(math.MinInt8),
		codec.UByte(math.MaxUint8),
		codec.Short(math.MinInt16),
		codec.UShort(math.MaxUint16),
		codec.Int(math.MinInt32),
		codec.Long(math.MinInt64),
		codec.Float(math.Pi),
		codec.Double(-math.MaxFloat64),
		codec.VarInt(-1),
		codec.VarLong(math.MinInt64),
		codec.String("graft ⛏"),
		id,
		codec.Slice[codec.VarInt]{0, -1, 25565, math.MaxInt32},
		codec.Slice[profileEntry]{
			{Name: "steve", ID: id},
			{Name: "alex"},
		},
		codec.Some(codec.String("skin-data")),
		codec.None[codec.String](),
	}

	for _, field := range fields {
		decoded := reflect.New(reflect.TypeOf(field))

		if err := codec.Unmarshal(field.Append(nil), decoded.Interface().(codec.FieldPtr)); err != nil {
			t.Errorf("decode %#v: %v", field, err)
			continue
		}

		if got := decoded.Elem().Interface(); !reflect.DeepEqual(got, field) {
			t.Errorf("round trip of %#v yielded %#v", field, got)
		}
	}
}

func TestFixedEncodingAgainstStdlib(t *testing.T) {
	var expected bytes.Buffer

	err := binary.Write(&expected, binary.BigEndian, struct {
		Flag   bool
		Kind   int8
		Level  uint8
		Delta  int16
		Port   uint16
		Block  int32
		Seed   int64
		Angle  float32
		Health float64
	}{true, math.MinInt8, math.MaxUint8, math.MinInt16, 25565, math.MinInt32, math.MinInt64, math.Pi, math.Pi})
	if err != nil {
		t.Fatal(err)
	}

	payload := codec.Marshal(
		codec.Bool(true),
		codec.Byte(math.MinInt8),
		codec.UByte(math.MaxUint8),
		codec.Short(math.MinInt16),
		codec.UShort(25565),
		codec.Int(math.MinInt32),
		codec.Long(math.MinInt64),
		codec.Float(math.Pi),
		codec.Double(math.Pi),
	)

	if !bytes.Equal(payload, expected.Bytes()) {
		t.Errorf("payload = %x, want %x", payload, expected.Bytes())
	}
}

func TestMarshalUnmarshalHandshake(t *testing.T) {
	payload := codec.Marshal(
		codec.VarInt(765),
		codec.String("mc.local"),
		codec.UShort(25565),
		codec.VarInt(1),
	)

	var (
		protocolVersion codec.VarInt
		serverAddress   codec.String
		serverPort      codec.UShort
		nextState       codec.VarInt
	)

	if err := codec.Unmarshal(payload, &protocolVersion, &serverAddress, &serverPort, &nextState); err != nil {
		t.Fatal(err)
	}

	if protocolVersion != 765 || serverAddress != "mc.local" || serverPort != 25565 || nextState != 1 {
		t.Errorf("decoded (%d, %q, %d, %d), want (765, \"mc.local\", 25565, 1)",
			protocolVersion, serverAddress, serverPort, nextState)
	}
}

func TestStringRejectsMalformedPayload(t *testing.T) {
	truncated := codec.String("graft").Append(nil)

	tests := []struct {
		input []byte
	}{
		{
			input: codec.AppendVar(nil, int32(-1)),
		},
		{
			input: truncated[:len(truncated)-1],
		},
	}

	for _, tt := range tests {
		var s codec.String

		if err := codec.Unmarshal(tt.input, &s); err == nil {
			t.Errorf("Unmarshal(%x): expected an error, got nil", tt.input)
		}
	}
}

func TestSliceRejectsOverclaimedCount(t *testing.T) {
	var s codec.Slice[codec.VarInt]

	if err := codec.Unmarshal(codec.AppendVar(nil, int32(math.MaxInt32)), &s); err == nil {
		t.Error("expected an error, got nil")
	}
}

func TestOptionGet(t *testing.T) {
	value, ok := codec.Some(codec.VarInt(7)).Get()

	if !ok || value != 7 {
		t.Errorf("Some(7).Get() = (%d, %t), want (7, true)", value, ok)
	}

	if _, ok := codec.None[codec.VarInt]().Get(); ok {
		t.Error("None().Get() reported presence")
	}
}

func TestNoneEncodesAsSingleByte(t *testing.T) {
	raw := codec.None[codec.VarInt]().Append(nil)

	if len(raw) != 1 {
		t.Errorf("None() encoded as %d bytes, want 1", len(raw))
	}
}
