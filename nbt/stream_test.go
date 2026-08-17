package nbt_test

import (
	"bytes"
	"compress/gzip"
	"reflect"
	"testing"

	"github.com/lrnxzz/graft/nbt"
)

func TestReadRecoversWrittenTree(t *testing.T) {
	original := richTree()

	var buf bytes.Buffer
	if err := nbt.Write(&buf, original); err != nil {
		t.Fatal(err)
	}

	decoded, err := nbt.Read(&buf)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(decoded, original) {
		t.Errorf("stream round trip mismatch:\n got %#v\nwant %#v", decoded, original)
	}
}

func TestReadNamedThroughGzip(t *testing.T) {
	original := nbt.Compound{
		"level": nbt.String("overworld"),
		"seed":  nbt.Long(-4096),
	}

	var buf bytes.Buffer

	writer := gzip.NewWriter(&buf)
	if err := nbt.WriteNamed(writer, "Data", original); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	reader, err := gzip.NewReader(&buf)
	if err != nil {
		t.Fatal(err)
	}

	name, decoded, err := nbt.ReadNamed(reader)
	if err != nil {
		t.Fatal(err)
	}

	if name != "Data" {
		t.Errorf("root name = %q, want Data", name)
	}
	if !reflect.DeepEqual(decoded, original) {
		t.Errorf("gzip round trip mismatch:\n got %#v\nwant %#v", decoded, original)
	}
}

func TestDecodePrefixReportsConsumed(t *testing.T) {
	original := nbt.Compound{
		"flag": nbt.Byte(1),
	}

	encoded := nbt.Encode(original)
	trailer := []byte{0xDE, 0xAD, 0xBE, 0xEF}

	decoded, consumed, err := nbt.DecodePrefix(append(encoded, trailer...))
	if err != nil {
		t.Fatal(err)
	}

	if consumed != len(encoded) {
		t.Errorf("consumed = %d, want %d", consumed, len(encoded))
	}
	if !reflect.DeepEqual(decoded, original) {
		t.Errorf("decoded %#v, want %#v", decoded, original)
	}
}
