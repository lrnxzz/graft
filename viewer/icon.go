package viewer

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"image/png"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/atlas"
	"github.com/lrnxzz/go-craft/codec/v765/items"
	"github.com/lrnxzz/go-craft/viewer/gpu"
)

//go:generate go run github.com/lrnxzz/go-craft/cli gen icons 765

//go:embed assets/icons.png
var iconsImage []byte

//go:embed assets/icons.json
var iconsMapping []byte

type Iconset struct {
	atlas *gpu.Atlas
	items map[string]int
}

func LoadIconset() (*Iconset, error) {
	img, err := png.Decode(bytes.NewReader(iconsImage))
	if err != nil {
		return nil, err
	}

	var file atlas.ItemSheet
	if err := json.Unmarshal(iconsMapping, &file); err != nil {
		return nil, err
	}

	return &Iconset{
		atlas: gpu.NewAtlas(gpu.NewTexture(img), file.Columns, file.Rows),
		items: file.Items,
	}, nil
}

func (s *Iconset) Delete() {
	s.atlas.Delete()
}

func (s *Iconset) Atlas() *gpu.Atlas {
	return s.atlas
}

func (s *Iconset) Icon(item gocraft.ItemID) (gpu.UV, bool) {
	registered, known := items.Of(item)
	if !known {
		return gpu.UV{}, false
	}

	tile, drawable := s.items[registered.Name.Path()]
	if !drawable {
		return gpu.UV{}, false
	}

	return s.atlas.Tile(tile), true
}
