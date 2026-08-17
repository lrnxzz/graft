package viewer

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"image/png"

	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/atlas"
	"github.com/lrnxzz/graft/codec/v765/blocks"
	"github.com/lrnxzz/graft/viewer/gpu"
	"github.com/lrnxzz/graft/viewer/mesh"
)

//go:generate go run github.com/lrnxzz/graft/cli gen atlas 765

//go:embed assets/atlas.png
var atlasImage []byte

//go:embed assets/blocks.json
var atlasMapping []byte

type Tileset struct {
	atlas  *gpu.Atlas
	blocks map[graft.Identifier]atlas.Faces
}

func LoadTileset() (*Tileset, error) {
	img, err := png.Decode(bytes.NewReader(atlasImage))
	if err != nil {
		return nil, err
	}

	var file atlas.BlockSheet
	if err := json.Unmarshal(atlasMapping, &file); err != nil {
		return nil, err
	}

	return &Tileset{
		atlas:  gpu.NewAtlas(gpu.NewTexture(img), file.Columns, file.Rows),
		blocks: file.Blocks,
	}, nil
}

func (t *Tileset) Delete() {
	t.atlas.Delete()
}

func (t *Tileset) Atlas() *gpu.Atlas {
	return t.atlas
}

func (t *Tileset) Tile(state graft.BlockState, face mesh.Face) gpu.UV {
	return t.atlas.Tile(t.index(state, face))
}

func (t *Tileset) index(state graft.BlockState, face mesh.Face) int {
	block, ok := blocks.Of(state)
	if !ok {
		return 0
	}

	tiles, known := t.blocks[block.Name]
	if !known {
		return 0
	}

	switch face {
	case mesh.Up:
		return tiles.Up
	case mesh.Down:
		return tiles.Down
	default:
		return tiles.Side
	}
}
