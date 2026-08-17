package v765

import (
	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/codec"
)

const (
	digStart  codec.VarInt = 0
	digCancel codec.VarInt = 1
	digFinish codec.VarInt = 2

	mainHand codec.VarInt = 0
)

type PlayerAction struct {
	Status   codec.VarInt
	Location graft.Position
	Face     codec.Byte
	Sequence codec.VarInt
}

func (*PlayerAction) ID() int32 {
	return 0x21
}

func (*PlayerAction) Name() string {
	return "PlayerAction"
}

func (*PlayerAction) State() codec.State {
	return codec.StatePlay
}

func (*PlayerAction) Direction() codec.Direction {
	return codec.Serverbound
}

func (p PlayerAction) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.Status, p.Location, p.Face, p.Sequence)
}

func (p *PlayerAction) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.Status, &p.Location, &p.Face, &p.Sequence)
}

type UseItemOn struct {
	Hand        codec.VarInt
	Location    graft.Position
	Face        codec.VarInt
	CursorX     codec.Float
	CursorY     codec.Float
	CursorZ     codec.Float
	InsideBlock codec.Bool
	Sequence    codec.VarInt
}

func (*UseItemOn) ID() int32 {
	return 0x35
}

func (*UseItemOn) Name() string {
	return "UseItemOn"
}

func (*UseItemOn) State() codec.State {
	return codec.StatePlay
}

func (*UseItemOn) Direction() codec.Direction {
	return codec.Serverbound
}

func (p UseItemOn) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.Hand, p.Location, p.Face,
		p.CursorX, p.CursorY, p.CursorZ, p.InsideBlock, p.Sequence)
}

func (p *UseItemOn) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.Hand, &p.Location, &p.Face,
		&p.CursorX, &p.CursorY, &p.CursorZ, &p.InsideBlock, &p.Sequence)
}

type AcknowledgeBlockChange struct {
	Sequence codec.VarInt
}

func (*AcknowledgeBlockChange) ID() int32 {
	return 0x05
}

func (*AcknowledgeBlockChange) Name() string {
	return "AcknowledgeBlockChange"
}

func (*AcknowledgeBlockChange) State() codec.State {
	return codec.StatePlay
}

func (*AcknowledgeBlockChange) Direction() codec.Direction {
	return codec.Clientbound
}

func (p AcknowledgeBlockChange) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.Sequence)
}

func (p *AcknowledgeBlockChange) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.Sequence)
}
