package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"sort"
	"strings"

	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/atlas"
	"github.com/spf13/cobra"
)

const breakStages = 10

// vanilla GUI art copied straight out of the jar, keyed by its path inside it
var guiTextures = map[string]string{
	"assets/minecraft/textures/gui/sprites/hud/hotbar.png":           viewerAssets + "/hotbar.png",
	"assets/minecraft/textures/gui/sprites/hud/hotbar_selection.png": viewerAssets + "/hotbar_selection.png",
	"assets/minecraft/textures/gui/container/inventory.png":          viewerAssets + "/inventory.png",
	"assets/minecraft/textures/entity/player/wide/steve.png":         viewerAssets + "/steve.png",
	"assets/minecraft/textures/font/ascii.png":                       viewerAssets + "/ascii.png",
	"assets/minecraft/textures/gui/sprites/hud/crosshair.png":        viewerAssets + "/crosshair.png",
}

type registryItem struct {
	Name graft.Identifier `json:"name"`
}

type iconSources struct {
	jar    []byte
	icons  map[string]image.Image
	blocks map[string]image.Image
}

func iconsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "icons <protocol>",
		Short: "Generate assets/{icons.png,icons.json} from a codec's Minecraft item sprites",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sources, err := loadIconSources(args[0])
			if err != nil {
				return err
			}

			index := assignIcons(sources.icons)

			if err := os.MkdirAll(viewerAssets, 0o755); err != nil {
				return err
			}
			if err := writeIconAtlas(viewerAssets+"/icons.png", sources.icons, index); err != nil {
				return err
			}
			if err := writeIcons(viewerAssets+"/icons.json", index); err != nil {
				return err
			}
			if err := writeBreaking(viewerAssets+"/breaking.png", sources.blocks); err != nil {
				return err
			}
			if err := extractTextures(sources.jar); err != nil {
				return err
			}

			cmd.Printf("icons: %d items\n", len(index))

			return nil
		},
	}
}

func loadIconSources(protocol string) (iconSources, error) {
	codec, err := openCodecAssets(protocol)
	if err != nil {
		return iconSources{}, err
	}

	rawItems, err := os.ReadFile(fmt.Sprintf("../codec/v%s/assets/items.json", protocol))
	if err != nil {
		return iconSources{}, err
	}

	var items []registryItem
	if err := json.Unmarshal(rawItems, &items); err != nil {
		return iconSources{}, err
	}
	var models map[string]blockModel
	if err := json.Unmarshal(codec.models, &models); err != nil {
		return iconSources{}, err
	}

	sprites, err := readSprites(codec.jar)
	if err != nil {
		return iconSources{}, err
	}
	blocks, err := readTextures(codec.jar)
	if err != nil {
		return iconSources{}, err
	}

	return iconSources{
		jar:    codec.jar,
		icons:  pickIcons(items, sprites, resolveFaces(models, blocks), blocks),
		blocks: blocks,
	}, nil
}

func blit(canvas *image.RGBA, src image.Image, originX, originY int) {
	bounds := src.Bounds()
	for y := range tileSize {
		for x := range tileSize {
			canvas.Set(originX+x, originY+y, src.At(bounds.Min.X+x, bounds.Min.Y+y))
		}
	}
}

// an item without a sprite of its own is drawn from the side texture of the
// block it places, which is why the block faces are consulted here too
func pickIcons(items []registryItem, sprites map[string]image.Image, faces map[graft.Identifier]faceNames, blocks map[string]image.Image) map[string]image.Image {
	icons := map[string]image.Image{}
	for _, item := range items {
		name := item.Name.Path()

		sprite, drawn := sprites[name]
		if drawn {
			icons[name] = sprite

			continue
		}

		face, cubic := faces[item.Name]
		if cubic {
			icons[name] = blocks[face.side]
		}
	}

	return icons
}

func assignIcons(icons map[string]image.Image) map[string]int {
	names := make([]string, 0, len(icons))
	for name := range icons {
		names = append(names, name)
	}
	sort.Strings(names)

	index := make(map[string]int, len(names))
	for tile, name := range names {
		index[name] = tile
	}

	return index
}

func readSprites(jar []byte) (map[string]image.Image, error) {
	archive, err := zip.NewReader(bytes.NewReader(jar), int64(len(jar)))
	if err != nil {
		return nil, err
	}

	const prefix = "assets/minecraft/textures/item/"
	sprites := map[string]image.Image{}
	for _, file := range archive.File {
		if !strings.HasPrefix(file.Name, prefix) || !strings.HasSuffix(file.Name, ".png") {
			continue
		}

		reader, err := file.Open()
		if err != nil {
			return nil, err
		}
		img, err := png.Decode(reader)
		reader.Close()
		if err != nil || img.Bounds().Dx() != tileSize {
			continue
		}

		sprites[strings.TrimSuffix(strings.TrimPrefix(file.Name, prefix), ".png")] = img
	}

	return sprites, nil
}

func extractTextures(jar []byte) error {
	archive, err := zip.NewReader(bytes.NewReader(jar), int64(len(jar)))
	if err != nil {
		return err
	}

	for _, file := range archive.File {
		target, wanted := guiTextures[file.Name]
		if !wanted {
			continue
		}

		reader, err := file.Open()
		if err != nil {
			return err
		}
		data, err := io.ReadAll(reader)
		reader.Close()
		if err != nil {
			return err
		}

		if err := os.WriteFile(target, data, 0o644); err != nil {
			return err
		}
	}

	return nil
}

func writeIconAtlas(pathname string, icons map[string]image.Image, index map[string]int) error {
	columns := atlasColumns(len(index))
	rows := atlasRows(len(index), columns)
	canvas := image.NewRGBA(image.Rect(0, 0, columns*tileSize, rows*tileSize))

	for name, tile := range index {
		blit(canvas, icons[name], (tile%columns)*tileSize, (tile/columns)*tileSize)
	}

	return writePNG(pathname, canvas)
}

func writeIcons(pathname string, index map[string]int) error {
	data := atlas.ItemSheet{
		Sheet: sheetOf(len(index)),
		Items: index,
	}

	encoded, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	return os.WriteFile(pathname, encoded, 0o644)
}

func writeBreaking(pathname string, blocks map[string]image.Image) error {
	canvas := image.NewRGBA(image.Rect(0, 0, tileSize, breakStages*tileSize))
	for stage := range breakStages {
		name := fmt.Sprintf("destroy_stage_%d", stage)

		src, ok := blocks[name]
		if !ok {
			return fmt.Errorf("gen: %s texture missing", name)
		}

		blit(canvas, src, 0, stage*tileSize)
	}

	return writePNG(pathname, canvas)
}
