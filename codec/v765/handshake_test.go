package v765_test

import (
	"testing"

	"github.com/lrnxzz/graft/codec"
	v765 "github.com/lrnxzz/graft/codec/v765"
)

func TestHandshakeCarriesConnectionParameters(t *testing.T) {
	original := &v765.Handshake{
		ProtocolVersion: v765.ProtocolVersion,
		ServerAddress:   "mc.local",
		ServerPort:      25565,
		NextState:       codec.VarInt(codec.StateLogin),
	}

	proto := v765.Protocol()
	decoded, ok, err := proto.Decode(codec.StateHandshaking, codec.Serverbound, codec.EncodeFrame(original))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("handshake not registered")
	}

	if got := decoded.(*v765.Handshake); *got != *original {
		t.Errorf("got %+v, want %+v", got, original)
	}
}
