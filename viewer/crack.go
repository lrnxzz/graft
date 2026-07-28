package viewer

import (
	"bytes"
	_ "embed"
	"image/png"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/viewer/gpu"
	"github.com/lrnxzz/go-craft/viewer/mesh"
)

//go:embed assets/breaking.png
var breakingImage []byte

const (
	breakStages  = 10
	crackInflate = 0.004
	noStage      = -1
)

type Crack struct {
	program *gpu.Program
	texture *gpu.Texture
	model   *gpu.Mesh
	block   gocraft.Position
	stage   int
}

func NewCrack() (*Crack, error) {
	program, err := gpu.NewProgram(vertexShader, fragmentShader)
	if err != nil {
		return nil, err
	}

	img, err := png.Decode(bytes.NewReader(breakingImage))
	if err != nil {
		return nil, err
	}

	return &Crack{
		program: program,
		texture: gpu.NewTexture(img),
		stage:   noStage,
	}, nil
}

func (c *Crack) Update(block gocraft.Position, progress float64, digging bool) {
	if !digging {
		c.clear()

		return
	}

	stage := min(int(progress*breakStages), breakStages-1)
	if c.model != nil && stage == c.stage && block == c.block {
		return
	}

	c.clear()
	c.block = block
	c.stage = stage
	c.model = mesh.Overlay(block, crackInflate, stageTile(stage)).Upload()
}

func stageTile(stage int) gpu.UV {
	return gpu.UV{
		U0: 0,
		V0: float32(stage) / breakStages,
		U1: 1,
		V1: float32(stage+1) / breakStages,
	}
}

func (c *Crack) Draw(camera Camera) {
	if c.model == nil {
		return
	}

	c.program.Use()
	c.program.Mat4("viewProjection", camera.ViewProjection())
	c.texture.Bind(0)
	c.program.Int("atlas", 0)
	c.model.Draw()
}

func (c *Crack) Close() {
	c.clear()
}

func (c *Crack) clear() {
	if c.model != nil {
		c.model.Delete()
		c.model = nil
	}

	c.stage = noStage
}
