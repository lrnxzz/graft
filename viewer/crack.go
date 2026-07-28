package viewer

import (
	"bytes"
	_ "embed"
	"image/png"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/agent"
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

// Crack is a WorldLayer with its own shader too: the break overlay is a textured
// cube, not an outline
type Crack struct {
	bot     *agent.Agent
	program *gpu.Program
	texture *gpu.Texture
	model   *gpu.Mesh
	block   gocraft.Position
	stage   int
}

func NewCrack(bot *agent.Agent) (*Crack, error) {
	program, err := gpu.NewProgram(vertexShader, fragmentShader)
	if err != nil {
		return nil, err
	}

	img, err := png.Decode(bytes.NewReader(breakingImage))
	if err != nil {
		return nil, err
	}

	return &Crack{
		bot:     bot,
		program: program,
		texture: gpu.NewTexture(img),
		stage:   noStage,
	}, nil
}

func (c *Crack) DrawWorld(painter *Painter) {
	block, progress, digging := c.bot.Excavation()
	c.update(block, progress, digging)

	if c.model == nil {
		return
	}

	c.program.Use()
	c.program.Mat4("viewProjection", painter.Camera().ViewProjection())
	c.texture.Bind(0)
	c.program.Int("atlas", 0)
	c.model.Draw()
}

func (c *Crack) update(block gocraft.Position, progress float64, digging bool) {
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

func (c *Crack) Close() {
	c.clear()
	c.texture.Delete()
	c.program.Delete()
}

func (c *Crack) clear() {
	if c.model != nil {
		c.model.Delete()
		c.model = nil
	}

	c.stage = noStage
}
