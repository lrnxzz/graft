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

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/spf13/cobra"
)

const breakStages = 10

var guiTextures = map[string]string{
	"assets/minecraft/textures/gui/sprites/hud/hotbar.png":           "viewer/assets/hotbar.png",
	"assets/minecraft/textures/gui/sprites/hud/hotbar_selection.png": "viewer/assets/hotbar_selection.png",
	"assets/minecraft/textures/gui/container/inventory.png":          "viewer/assets/inventory.png",
	"assets/minecraft/textures/entity/player/wide/steve.png":         "viewer/assets/steve.png",
	"assets/minecraft/textures/font/ascii.png":                       "viewer/assets/ascii.png",
	"assets/minecraft/textures/gui/sprites/hud/crosshair.png":        "viewer/assets/crosshair.png",
}

type registryItem struct {
	Name gocraft.Identifier `json:"name"`
}

type iconsFile struct {
	Tile    int            `json:"tile"`
	Columns int            `json:"columns"`
	Rows    int            `json:"rows"`
	Items   map[string]int `json:"items"`
}

type iconSources struct {
	jar    []byte
	icons  map[string]image.Image
	blocks map[string]image.Image
}

func iconsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "icons <version>",
		Short: "Generate viewer/assets/{icons.png,icons.json} from a codec's Minecraft item sprites",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			assets := fmt.Sprintf("codec/v%s/assets", args[0])

			sources, err := loadIconSources(assets)
			if err != nil {
				return err
			}

			index := assignIcons(sources.icons)

			if err := os.MkdirAll("viewer/assets", 0o755); err != nil {
				return err
			}
			if err := writeIconAtlas("viewer/assets/icons.png", sources.icons, index); err != nil {
				return err
			}
			if err := writeIcons("viewer/assets/icons.json", index); err != nil {
				return err
			}
			if err := writeBreaking("viewer/assets/breaking.png", sources.blocks); err != nil {
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

func loadIconSources(assets string) (iconSources, error) {
	fetchModels := func() ([]byte, error) {
		return fetch(modelsURL)
	}

	jar, err := cached(assets+"/client.jar", fetchClientJar)
	if err != nil {
		return iconSources{}, err
	}
	rawItems, err := os.ReadFile(assets + "/items.json")
	if err != nil {
		return iconSources{}, err
	}
	rawModels, err := cached(assets+"/blocks_models.json", fetchModels)
	if err != nil {
		return iconSources{}, err
	}

	var items []registryItem
	if err := json.Unmarshal(rawItems, &items); err != nil {
		return iconSources{}, err
	}
	var models map[string]blockModel
	if err := json.Unmarshal(rawModels, &models); err != nil {
		return iconSources{}, err
	}

	sprites, err := readSprites(jar)
	if err != nil {
		return iconSources{}, err
	}
	blocks, err := readTextures(jar)
	if err != nil {
		return iconSources{}, err
	}

	return iconSources{
		jar:    jar,
		icons:  pickIcons(items, sprites, resolveFaces(models, blocks), blocks),
		blocks: blocks,
	}, nil
}

func pickIcons(items []registryItem, sprites map[string]image.Image, faces map[string]faceNames, blocks map[string]image.Image) map[string]image.Image {
	icons := map[string]image.Image{}
	for _, item := range items {
		name := item.Name.Path()

		sprite, drawn := sprites[name]
		if drawn {
			icons[name] = sprite

			continue
		}

		face, cubic := faces[name]
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
	columns := atlasColumns(len(index))
	data := iconsFile{
		Tile:    tileSize,
		Columns: columns,
		Rows:    atlasRows(len(index), columns),
		Items:   index,
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
