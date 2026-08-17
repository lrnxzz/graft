package v765

import (
	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/codec"
	"github.com/lrnxzz/graft/codec/v765/blocks"
	"github.com/lrnxzz/graft/codec/v765/items"
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

type blockPrediction struct {
	position graft.Position
	server   graft.BlockState
}

func (s *Session) onBlockAck(c *codec.Client, p *AcknowledgeBlockChange) error {
	for _, prediction := range s.pending.Ack(p.Sequence.Int32()) {
		s.world.SetBlock(prediction.position.X, prediction.position.Y, prediction.position.Z, prediction.server)
	}

	return nil
}

func (s *Session) StartDigging(hit graft.RayHit) error {
	return s.action(digStart, hit)
}

func (s *Session) CancelDigging(hit graft.RayHit) error {
	return s.action(digCancel, hit)
}

func (s *Session) FinishDigging(hit graft.RayHit) error {
	if err := s.action(digFinish, hit); err != nil {
		return err
	}

	s.world.SetBlock(hit.Block.X, hit.Block.Y, hit.Block.Z, 0)

	return nil
}

func (s *Session) action(status codec.VarInt, hit graft.RayHit) error {
	packet := &PlayerAction{
		Status:   status,
		Location: hit.Block,
		Face:     codec.Byte(hit.Face),
		Sequence: codec.VarInt(s.pending.Push(s.predict(hit.Block))),
	}

	return s.client.Send(packet)
}

func (s *Session) PlaceBlock(hit graft.RayHit) error {
	placed := hit.Block.Neighbor(hit.Face)
	cursor := hit.Point.Sub(hit.Block.Corner())

	use := &UseItemOn{
		Hand:     mainHand,
		Location: hit.Block,
		Face:     codec.VarInt(hit.Face),
		CursorX:  codec.Float(cursor.X),
		CursorY:  codec.Float(cursor.Y),
		CursorZ:  codec.Float(cursor.Z),
		Sequence: codec.VarInt(s.pending.Push(s.predict(placed))),
	}
	if err := s.client.Send(use); err != nil {
		return err
	}

	held, ok := items.Of(s.inventory.Held().Item)
	if !ok {
		return nil
	}

	block, ok := blocks.Named(held.Name)
	if !ok {
		return nil
	}

	s.world.SetBlock(placed.X, placed.Y, placed.Z, block.DefaultState)

	return nil
}

func (s *Session) predict(position graft.Position) blockPrediction {
	server, _ := s.world.BlockAt(position)

	return blockPrediction{
		position: position,
		server:   server,
	}
}

func (s *Session) applyServerBlock(change BlockChange) {
	position := graft.Position{
		X: change.X,
		Y: change.Y,
		Z: change.Z,
	}

	predicted := false
	s.pending.Each(func(prediction *blockPrediction) {
		if prediction.position == position {
			prediction.server = change.State
			predicted = true
		}
	})
	if predicted {
		return
	}

	s.world.SetBlock(change.X, change.Y, change.Z, change.State)
}
