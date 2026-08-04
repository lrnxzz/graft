package viewer

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/png"
	"regexp"
	"strings"
	"sync"

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

func iconSheet() (atlas.ItemSheet, error) {
	var file atlas.ItemSheet
	err := json.Unmarshal(iconsMapping, &file)

	return file, err
}

func LoadIconset() (*Iconset, error) {
	img, err := png.Decode(bytes.NewReader(iconsImage))
	if err != nil {
		return nil, err
	}

	file, err := iconSheet()
	if err != nil {
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

// A page names a sprite by class, the same way the hud names it by item
var mentioned = regexp.MustCompile(`icon-([a-z0-9_]+)`)

const iconRules = `.icon{display:inline-block;flex:none;width:var(--icon,32px);height:var(--icon,32px);` +
	`background-size:100% 100%;background-repeat:no-repeat;` +
	`image-rendering:pixelated;image-rendering:-webkit-optimize-contrast}
`

// IconStylesheet writes css for exactly the sprites a page mentions, so that
// <i class="icon icon-diamond_ore"></i> shows the same art the hud draws.
//
// Only what is named gets inlined, because the whole atlas cannot be: it is
// 370 KB as one data uri and the engine silently refuses anything past about
// 128 KB, which measured as a sprite that simply never appears. Cut one tile at
// a time, a panel listing twenty ores costs a few kilobytes.
//
// It needs no gpu, so it is a plain function rather than a method on Iconset: a
// page can be built and checked without a window.
func IconStylesheet(markup string) (string, error) {
	names := mentioned.FindAllStringSubmatch(markup, -1)
	if len(names) == 0 {
		return "", nil
	}

	file, err := iconSheet()
	if err != nil {
		return "", err
	}

	sheet := strings.Builder{}
	sheet.WriteString(iconRules)

	written := make(map[string]bool, len(names))
	for _, found := range names {
		name := found[1]
		if written[name] {
			continue
		}
		written[name] = true

		tile, drawable := file.Items[name]
		if !drawable {
			continue
		}

		sprite, err := cutIcon(file.Sheet, tile)
		if err != nil {
			return "", err
		}

		sheet.WriteString(".icon-")
		sheet.WriteString(name)
		sheet.WriteString("{background-image:url(data:image/png;base64,")
		sheet.WriteString(base64.StdEncoding.EncodeToString(sprite))
		sheet.WriteString(")}\n")
	}

	return sheet.String(), nil
}

// cutting keeps what has already been cut out of the atlas. Pages reload on
// every change, so the same tile is asked for over and over; the lock is there
// because nothing in the type says which thread a page is built on.
type cutting struct {
	sync.Mutex
	atlas  image.Image
	read   sync.Once
	failed error
	kept   map[int][]byte
}

var sprites = cutting{kept: map[int][]byte{}}

// The engine ignores image-rendering entirely — pixelated, crisp-edges and
// -webkit-optimize-contrast all measured identical to no rule at all, and a 16px
// sprite blown up to 30px came out smeared. So the sprite is enlarged here, by
// repeating pixels, and the engine only ever scales it down. Flat colour costs
// almost nothing to encode at this size.
const iconScale = 8

// cutIcon crops one tile and keeps it, since a page reloads on every change and
// re-encoding the same sprite each time would be paid for in frames
func cutIcon(sheet atlas.Sheet, tile int) ([]byte, error) {
	sprites.Lock()
	defer sprites.Unlock()

	kept, have := sprites.kept[tile]
	if have {
		return kept, nil
	}

	sprites.read.Do(func() {
		sprites.atlas, sprites.failed = png.Decode(bytes.NewReader(iconsImage))
	})
	if sprites.failed != nil {
		return nil, sprites.failed
	}

	corner := image.Pt((tile%sheet.Columns)*sheet.Tile, (tile/sheet.Columns)*sheet.Tile)

	side := sheet.Tile * iconScale
	grown := image.NewNRGBA(image.Rect(0, 0, side, side))

	for y := range side {
		for x := range side {
			grown.Set(x, y, sprites.atlas.At(corner.X+x/iconScale, corner.Y+y/iconScale))
		}
	}

	written := bytes.Buffer{}
	if err := png.Encode(&written, grown); err != nil {
		return nil, err
	}

	sprites.kept[tile] = written.Bytes()

	return sprites.kept[tile], nil
}
