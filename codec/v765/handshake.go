package v765

import (
	"github.com/lrnxzz/graft/codec"
)

type Handshake struct {
	ProtocolVersion codec.VarInt
	ServerAddress   codec.String
	ServerPort      codec.UShort
	NextState       codec.VarInt
}

func (*Handshake) ID() int32 {
	return 0x00
}

func (*Handshake) Name() string {
	return "Handshake"
}

func (*Handshake) State() codec.State {
	return codec.StateHandshaking
}

func (*Handshake) Direction() codec.Direction {
	return codec.Serverbound
}

func (p Handshake) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.ProtocolVersion, p.ServerAddress, p.ServerPort, p.NextState)
}

func (p *Handshake) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.ProtocolVersion, &p.ServerAddress, &p.ServerPort, &p.NextState)
}
