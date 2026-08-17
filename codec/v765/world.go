package v765

import (
	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/codec"
)

// the server paces chunk delivery from this number; a bot has no frame budget
// to protect, so it asks for as much as the vanilla client ever would
const chunkBatchRate codec.Float = 64

func (s *Session) onChunkBatchFinished(c *codec.Client, p *ChunkBatchFinished) error {
	return c.Send(&ChunkBatchReceived{ChunksPerTick: chunkBatchRate})
}

func (s *Session) onChunkData(c *codec.Client, p *ChunkData) error {
	column, err := p.Column(s.bounds.minY, s.bounds.height)
	if err != nil {
		return err
	}

	s.world.LoadColumn(column)

	return nil
}

func (s *Session) onUnloadChunk(c *codec.Client, p *UnloadChunk) error {
	s.world.UnloadColumn(graft.Chunk(p.X.Int32(), p.Z.Int32()))

	return nil
}

func (s *Session) onBlockUpdate(c *codec.Client, p *BlockUpdate) error {
	s.applyServerBlock(p.Change())

	return nil
}

func (s *Session) onSectionBlocks(c *codec.Client, p *SectionBlocksUpdate) error {
	for _, change := range p.Changes() {
		s.applyServerBlock(change)
	}

	return nil
}
