package v765

import (
	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/codec"
)

type ChunkData struct {
	X          codec.Int
	Z          codec.Int
	Heightmaps codec.NBT
	Sections   codec.Bytes
}

func (*ChunkData) ID() int32 {
	return 0x25
}

func (*ChunkData) Name() string {
	return "ChunkData"
}

func (*ChunkData) State() codec.State {
	return codec.StatePlay
}

func (*ChunkData) Direction() codec.Direction {
	return codec.Clientbound
}

func (p ChunkData) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.X, p.Z, p.Heightmaps, p.Sections)
}

func (p *ChunkData) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.X, &p.Z, &p.Heightmaps, &p.Sections)
}

func (p *ChunkData) Column(minY, height int) (*graft.Column, error) {
	column := graft.ChunkColumn(p.X.Int32(), p.Z.Int32(), minY, height)
	if err := column.Decode(codec.NewReader(p.Sections)); err != nil {
		return nil, err
	}

	return column, nil
}

// Z really does come first: 765 encodes this one packet backwards
type UnloadChunk struct {
	Z codec.Int
	X codec.Int
}

func (*UnloadChunk) ID() int32 {
	return 0x1F
}

func (*UnloadChunk) Name() string {
	return "UnloadChunk"
}

func (*UnloadChunk) State() codec.State {
	return codec.StatePlay
}

func (*UnloadChunk) Direction() codec.Direction {
	return codec.Clientbound
}

func (p UnloadChunk) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.Z, p.X)
}

func (p *UnloadChunk) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.Z, &p.X)
}

type BlockUpdate struct {
	Location graft.Position
	Block    codec.VarInt
}

func (*BlockUpdate) ID() int32 {
	return 0x09
}

func (*BlockUpdate) Name() string {
	return "BlockUpdate"
}

func (*BlockUpdate) State() codec.State {
	return codec.StatePlay
}

func (*BlockUpdate) Direction() codec.Direction {
	return codec.Clientbound
}

func (p BlockUpdate) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.Location, p.Block)
}

func (p *BlockUpdate) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.Location, &p.Block)
}

type BlockChange struct {
	X     int
	Y     int
	Z     int
	State graft.BlockState
}

func (p *BlockUpdate) Change() BlockChange {
	return BlockChange{
		X:     p.Location.X,
		Y:     p.Location.Y,
		Z:     p.Location.Z,
		State: graft.BlockState(p.Block),
	}
}

type SectionBlocksUpdate struct {
	Section codec.Long
	Packed  codec.Slice[codec.VarLong]
}

func (*SectionBlocksUpdate) ID() int32 {
	return 0x47
}

func (*SectionBlocksUpdate) Name() string {
	return "SectionBlocksUpdate"
}

func (*SectionBlocksUpdate) State() codec.State {
	return codec.StatePlay
}

func (*SectionBlocksUpdate) Direction() codec.Direction {
	return codec.Clientbound
}

func (p SectionBlocksUpdate) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.Section, p.Packed)
}

func (p *SectionBlocksUpdate) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.Section, &p.Packed)
}

func (p *SectionBlocksUpdate) Changes() []BlockChange {
	baseX := int(p.Section.Signed(42, 22)) * 16
	baseZ := int(p.Section.Signed(20, 22)) * 16
	baseY := int(p.Section.Signed(0, 20)) * 16

	changes := make([]BlockChange, len(p.Packed))
	for i, packed := range p.Packed {
		block := codec.Long(packed)
		changes[i] = BlockChange{
			X:     baseX + int(block.Unsigned(8, 4)),
			Y:     baseY + int(block.Unsigned(0, 4)),
			Z:     baseZ + int(block.Unsigned(4, 4)),
			State: graft.BlockState(block.Unsigned(12, 52)),
		}
	}

	return changes
}

type ChunkBatchStart struct{}

func (*ChunkBatchStart) ID() int32 {
	return 0x0D
}

func (*ChunkBatchStart) Name() string {
	return "ChunkBatchStart"
}

func (*ChunkBatchStart) State() codec.State {
	return codec.StatePlay
}

func (*ChunkBatchStart) Direction() codec.Direction {
	return codec.Clientbound
}

func (ChunkBatchStart) Append(dst []byte) []byte {
	return dst
}

func (*ChunkBatchStart) Decode(*codec.Reader) error {
	return nil
}

type ChunkBatchFinished struct {
	BatchSize codec.VarInt
}

func (*ChunkBatchFinished) ID() int32 {
	return 0x0C
}

func (*ChunkBatchFinished) Name() string {
	return "ChunkBatchFinished"
}

func (*ChunkBatchFinished) State() codec.State {
	return codec.StatePlay
}

func (*ChunkBatchFinished) Direction() codec.Direction {
	return codec.Clientbound
}

func (p ChunkBatchFinished) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.BatchSize)
}

func (p *ChunkBatchFinished) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.BatchSize)
}

type ChunkBatchReceived struct {
	ChunksPerTick codec.Float
}

func (*ChunkBatchReceived) ID() int32 {
	return 0x07
}

func (*ChunkBatchReceived) Name() string {
	return "ChunkBatchReceived"
}

func (*ChunkBatchReceived) State() codec.State {
	return codec.StatePlay
}

func (*ChunkBatchReceived) Direction() codec.Direction {
	return codec.Serverbound
}

func (p ChunkBatchReceived) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.ChunksPerTick)
}

func (p *ChunkBatchReceived) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.ChunksPerTick)
}
