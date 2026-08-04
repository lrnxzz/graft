package codec_test

import (
	"testing"

	"github.com/lrnxzz/go-craft/codec"
)

type keepAlivePacket struct {
	Nonce codec.Long
	Label codec.String
}

func (*keepAlivePacket) ID() int32 {
	return 0x2A
}

func (*keepAlivePacket) Name() string {
	return "KeepAlive"
}

func (*keepAlivePacket) State() codec.State {
	return codec.StatePlay
}

func (*keepAlivePacket) Direction() codec.Direction {
	return codec.Clientbound
}

func (p keepAlivePacket) Append(dst []byte) []byte {
	return codec.Marshal(p.Nonce, p.Label)
}

func (p *keepAlivePacket) Decode(r *codec.Reader) error {
	if err := p.Nonce.Decode(r); err != nil {
		return err
	}

	return p.Label.Decode(r)
}

func TestProtocolDecodesRegisteredPacket(t *testing.T) {
	proto := codec.NewProtocol()
	codec.Bind[keepAlivePacket](proto)

	original := &keepAlivePacket{
		Nonce: 99,
		Label: "alive",
	}
	frame := codec.EncodeFrame(original)

	packet, ok, err := proto.Decode(codec.StatePlay, codec.Clientbound, frame)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("packet not registered in protocol")
	}

	got, isKeepAlive := packet.(*keepAlivePacket)
	if !isKeepAlive {
		t.Fatalf("decoded %T, want *keepAlivePacket", packet)
	}
	if *got != *original {
		t.Errorf("round trip got %+v, want %+v", got, original)
	}
}

func TestProtocolUnknownPacket(t *testing.T) {
	proto := codec.NewProtocol()

	_, ok, err := proto.Decode(codec.StatePlay, codec.Clientbound, codec.Frame{
		ID: 0x99,
	})
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("unknown packet reported as registered")
	}
}

func TestProtocolIsolatesStateAndDirection(t *testing.T) {
	proto := codec.NewProtocol()
	codec.Bind[keepAlivePacket](proto)

	if _, ok := proto.NewPacket(codec.StateLogin, codec.Clientbound, 0x2A); ok {
		t.Error("packet leaked across states")
	}
	if _, ok := proto.NewPacket(codec.StatePlay, codec.Serverbound, 0x2A); ok {
		t.Error("packet leaked across directions")
	}
}

func TestBindPanicsOnDuplicateRegistration(t *testing.T) {
	proto := codec.NewProtocol()
	codec.Bind[keepAlivePacket](proto)

	defer func() {
		if recover() == nil {
			t.Error("expected a panic registering the same key twice, got none")
		}
	}()

	codec.Bind[keepAlivePacket](proto)
}

func TestStateAndDirectionString(t *testing.T) {
	if got := codec.StatePlay.String(); got != "play" {
		t.Errorf("StatePlay.String() = %q, want play", got)
	}
	if got := codec.Clientbound.String(); got != "clientbound" {
		t.Errorf("Clientbound.String() = %q, want clientbound", got)
	}
}
