package gocraft_test

import (
	"testing"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/codec"
)

func TestIdentifierParts(t *testing.T) {
	tests := []struct {
		id        gocraft.Identifier
		namespace string
		path      string
	}{
		{
			id:        "minecraft:stone",
			namespace: "minecraft",
			path:      "stone",
		},
		{
			id:        "stone",
			namespace: "minecraft",
			path:      "stone",
		},
		{
			id:        "mymod:block/custom",
			namespace: "mymod",
			path:      "block/custom",
		},
	}

	for _, tt := range tests {
		if got := tt.id.Namespace(); got != tt.namespace {
			t.Errorf("%q.Namespace() = %q, want %q", tt.id, got, tt.namespace)
		}
		if got := tt.id.Path(); got != tt.path {
			t.Errorf("%q.Path() = %q, want %q", tt.id, got, tt.path)
		}
	}
}

func TestIdentifierValid(t *testing.T) {
	tests := []struct {
		id    gocraft.Identifier
		valid bool
	}{
		{
			id:    "minecraft:stone",
			valid: true,
		},
		{
			id:    "mymod:block/deep/path",
			valid: true,
		},
		{
			id:    "Minecraft:Stone",
			valid: false,
		},
		{
			id:    "minecraft:",
			valid: false,
		},
		{
			id:    "minecraft:sla/sh",
			valid: true,
		},
		{
			id:    "space mod:x",
			valid: false,
		},
	}

	for _, tt := range tests {
		if got := tt.id.Valid(); got != tt.valid {
			t.Errorf("%q.Valid() = %t, want %t", tt.id, got, tt.valid)
		}
	}
}

func TestIdentifierRecoversEncodedValue(t *testing.T) {
	want := gocraft.NewIdentifier("mymod", "block/custom")

	var got gocraft.Identifier
	if err := codec.Unmarshal(want.Append(nil), &got); err != nil {
		t.Fatal(err)
	}

	if got != want {
		t.Errorf("round trip of %q yielded %q", want, got)
	}
}
