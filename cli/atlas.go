package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/atlas"
	"github.com/spf13/cobra"
)

const (
	tileSize    = 16
	manifestURL = "https://launchermeta.mojang.com/mc/game/version_manifest_v2.json"

	// both viewer generators run from the viewer package, so their output lands
	// beside the go:embed directives that pick it up
	viewerAssets = "assets"

	// a texture reference may point at another slot in the same model, and a
	// malformed pack can make that chain loop
	maxTextureIndirection = 8
)

type versionManifest struct {
	Versions []manifestVersion `json:"versions"`
}

type manifestVersion struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type clientVersion struct {
	Downloads clientDownloads `json:"downloads"`
}

type clientDownloads struct {
	Client clientJar `json:"client"`
}

type clientJar struct {
	URL string `json:"url"`
}

type blockModel struct {
	Parent   string            `json:"parent"`
	Textures map[string]string `json:"textures"`
}

type faceNames struct {
	up, down, side string
}

// like every other generator this one runs through go:generate, here from the
// viewer package that embeds what it writes
func atlasCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "atlas <protocol>",
		Short: "Generate assets/{atlas.png,blocks.json} from a codec's Minecraft textures",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sources, err := openCodecAssets(args[0])
			if err != nil {
				return err
			}

			var models map[string]blockModel
			if err := json.Unmarshal(sources.models, &models); err != nil {
				return err
			}
			textures, err := readTextures(sources.jar)
			if err != nil {
				return err
			}

			faces := resolveFaces(models, textures)
			index := assignTiles(faces)

			if err := os.MkdirAll(viewerAssets, 0o755); err != nil {
				return err
			}
			if err := writeAtlas(viewerAssets+"/atlas.png", textures, index); err != nil {
				return err
			}
			if err := writeBlocks(viewerAssets+"/blocks.json", faces, index); err != nil {
				return err
			}

			cmd.Printf("atlas: %d blocks, %d tiles\n", len(faces), len(index))

			return nil
		},
	}
}

// codecAssets are the two downloads both viewer generators start from, cached
// next to the codec they describe
type codecAssets struct {
	jar    []byte
	models []byte
}

func openCodecAssets(protocol string) (codecAssets, error) {
	known, err := releaseOf(protocol)
	if err != nil {
		return codecAssets{}, err
	}

	cache := fmt.Sprintf("../codec/v%s/assets", protocol)

	downloadJar := func() ([]byte, error) {
		return fetchClientJar(known.game)
	}
	jar, err := cached(cache+"/client.jar", downloadJar)
	if err != nil {
		return codecAssets{}, err
	}

	downloadModels := func() ([]byte, error) {
		return fetch(known.modelsURL())
	}
	models, err := cached(cache+"/blocks_models.json", downloadModels)
	if err != nil {
		return codecAssets{}, err
	}

	return codecAssets{
		jar:    jar,
		models: models,
	}, nil
}

func cached(pathname string, download func() ([]byte, error)) ([]byte, error) {
	cachedData, err := os.ReadFile(pathname)
	if err == nil {
		return cachedData, nil
	}

	data, err := download()
	if err != nil {
		return nil, err
	}

	return data, os.WriteFile(pathname, data, 0o644)
}

func readTextures(jar []byte) (map[string]image.Image, error) {
	archive, err := zip.NewReader(bytes.NewReader(jar), int64(len(jar)))
	if err != nil {
		return nil, err
	}

	const prefix = "assets/minecraft/textures/block/"
	textures := map[string]image.Image{}
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

		textures[strings.TrimSuffix(strings.TrimPrefix(file.Name, prefix), ".png")] = img
	}

	return textures, nil
}

func fetchClientJar(game string) ([]byte, error) {
	raw, err := fetch(manifestURL)
	if err != nil {
		return nil, err
	}

	var manifest versionManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return nil, err
	}

	var versionURL string
	for _, version := range manifest.Versions {
		if version.ID == game {
			versionURL = version.URL
		}
	}
	if versionURL == "" {
		return nil, fmt.Errorf("gen: version %s not in manifest", game)
	}

	raw, err = fetch(versionURL)
	if err != nil {
		return nil, err
	}

	var detail clientVersion
	if err := json.Unmarshal(raw, &detail); err != nil {
		return nil, err
	}

	return fetch(detail.Downloads.Client.URL)
}

func fetch(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gen: GET %s: %s", url, response.Status)
	}

	return io.ReadAll(response.Body)
}

// the model name is where the wire data becomes a domain identity, so the
// conversion happens once here and never again downstream
func resolveFaces(models map[string]blockModel, textures map[string]image.Image) map[gocraft.Identifier]faceNames {
	faces := map[gocraft.Identifier]faceNames{}
	for name := range models {
		merged := mergeTextures(models, name)

		up := lookup(merged, textures, "up", "top", "end", "all", "side", "particle")
		down := lookup(merged, textures, "down", "bottom", "end", "all", "side", "particle")
		side := lookup(merged, textures, "north", "side", "all", "end", "particle")
		if up == "" || down == "" || side == "" {
			continue
		}

		faces[gocraft.Identifier(name)] = faceNames{
			up:   up,
			down: down,
			side: side,
		}
	}

	return faces
}

func mergeTextures(models map[string]blockModel, name string) map[string]string {
	merged := map[string]string{}
	for name != "" {
		current, ok := models[name]
		if !ok {
			break
		}
		for key, value := range current.Textures {
			_, inherited := merged[key]
			if !inherited {
				merged[key] = value
			}
		}
		name = trimNamespace(current.Parent)
	}

	return merged
}

func lookup(merged map[string]string, textures map[string]image.Image, keys ...string) string {
	for _, key := range keys {
		ref, ok := merged[key]
		if !ok {
			continue
		}
		name := dereference(merged, ref, 0)

		_, drawn := textures[name]
		if drawn {
			return name
		}
	}

	return ""
}

func dereference(merged map[string]string, ref string, depth int) string {
	if depth > maxTextureIndirection {
		return ""
	}
	if strings.HasPrefix(ref, "#") {
		next, ok := merged[ref[1:]]
		if !ok {
			return ""
		}

		return dereference(merged, next, depth+1)
	}

	return trimNamespace(ref)
}

func trimNamespace(ref string) string {
	path := gocraft.Identifier(ref).Path()

	return strings.TrimPrefix(path, "block/")
}

func assignTiles(faces map[gocraft.Identifier]faceNames) map[string]int {
	used := map[string]bool{}
	for _, face := range faces {
		used[face.up], used[face.down], used[face.side] = true, true, true
	}

	names := make([]string, 0, len(used))
	for name := range used {
		names = append(names, name)
	}
	sort.Strings(names)

	index := make(map[string]int, len(names))
	for i, name := range names {
		index[name] = i
	}

	return index
}

func writeAtlas(pathname string, textures map[string]image.Image, index map[string]int) error {
	columns := atlasColumns(len(index))
	rows := atlasRows(len(index), columns)
	canvas := image.NewRGBA(image.Rect(0, 0, columns*tileSize, rows*tileSize))

	for name, tile := range index {
		originX, originY := (tile%columns)*tileSize, (tile/columns)*tileSize
		src := textures[name]
		tinted := foliage(name)
		for y := range tileSize {
			for x := range tileSize {
				pixel := src.At(src.Bounds().Min.X+x, src.Bounds().Min.Y+y)
				canvas.Set(originX+x, originY+y, tint(pixel, tinted))
			}
		}
	}

	return writePNG(pathname, canvas)
}

func writeBlocks(pathname string, faces map[gocraft.Identifier]faceNames, index map[string]int) error {
	data := atlas.BlockSheet{
		Sheet:  sheetOf(len(index)),
		Blocks: make(map[gocraft.Identifier]atlas.Faces, len(faces)),
	}
	for name, face := range faces {
		data.Blocks[name] = atlas.Faces{
			Up:   index[face.up],
			Down: index[face.down],
			Side: index[face.side],
		}
	}

	encoded, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	return os.WriteFile(pathname, encoded, 0o644)
}

// the sheet is squared off so both generators lay their tiles out the same way
func sheetOf(tiles int) atlas.Sheet {
	columns := atlasColumns(tiles)

	return atlas.Sheet{
		Tile:    tileSize,
		Columns: columns,
		Rows:    atlasRows(tiles, columns),
	}
}

func atlasColumns(tiles int) int {
	return int(math.Ceil(math.Sqrt(float64(tiles))))
}

func atlasRows(tiles, columns int) int {
	return (tiles + columns - 1) / columns
}

func writePNG(pathname string, canvas image.Image) error {
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, canvas); err != nil {
		return err
	}

	return os.WriteFile(pathname, encoded.Bytes(), 0o644)
}

func foliage(name string) bool {
	return strings.Contains(name, "grass_block_top") ||
		strings.Contains(name, "leaves") ||
		strings.Contains(name, "grass") ||
		strings.Contains(name, "fern") ||
		strings.Contains(name, "vine")
}

func tint(pixel color.Color, tinted bool) color.Color {
	if !tinted {
		return pixel
	}

	r, g, b, a := pixel.RGBA()

	return color.RGBA{
		R: uint8((r >> 8) * 124 / 255),
		G: uint8((g >> 8) * 189 / 255),
		B: uint8((b >> 8) * 84 / 255),
		A: uint8(a >> 8),
	}
}
