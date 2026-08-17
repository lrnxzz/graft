package v765_test

import (
	"testing"

	"github.com/lrnxzz/graft/codec"
	v765 "github.com/lrnxzz/graft/codec/v765"
)

func TestUnknownPacketIsSkipped(t *testing.T) {
	proto := v765.Protocol()

	_, ok, err := proto.Decode(codec.StateLogin, codec.Clientbound, codec.Frame{
		ID: 0x7F,
	})
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("unregistered packet reported as known")
	}
}
