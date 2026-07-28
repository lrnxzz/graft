package viewer

import (
	"bytes"
	_ "embed"
	"image/png"
	"math"

	"github.com/go-gl/mathgl/mgl32"
	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/codec/v765/blocks"
	"github.com/lrnxzz/go-craft/codec/v765/items"
	"github.com/lrnxzz/go-craft/viewer/gpu"
	"github.com/lrnxzz/go-craft/viewer/mesh"
)

//go:embed assets/steve.png
var skinImage []byte

type Hand struct {
	program *gpu.Program
	blocks  *Tileset
	icons   *Iconset
	skin    *gpu.Texture
	arm     *gpu.Mesh
	model   *gpu.Mesh
	item    gocraft.ItemID
	cube    bool
	empty   bool
}

func NewHand(tiles *Tileset, icons *Iconset) (*Hand, error) {
	program, err := gpu.NewProgram(vertexShader, fragmentShader)
	if err != nil {
		return nil, err
	}

	img, err := png.Decode(bytes.NewReader(skinImage))
	if err != nil {
		return nil, err
	}

	return &Hand{
		program: program,
		blocks:  tiles,
		icons:   icons,
		skin:    gpu.NewTexture(img),
		arm:     mesh.Arm().Upload(),
		empty:   true,
	}, nil
}

func (h *Hand) Update(held gocraft.ItemStack) {
	if held.Empty() {
		h.clear()
		h.item = 0
		h.empty = true

		return
	}

	if held.Item == h.item && h.model != nil {
		return
	}
	h.clear()
	h.item = held.Item
	h.empty = false

	item, known := items.Of(held.Item)
	if !known {
		return
	}

	block, placeable := blocks.Named(item.Name)
	if placeable {
		h.model = mesh.Block(block.DefaultState, h.blocks).Upload()
		h.cube = true

		return
	}

	uv, drawable := h.icons.Icon(held.Item)
	if !drawable {
		return
	}

	h.model = mesh.Sprite(uv).Upload()
	h.cube = false
}

func (h *Hand) Draw(window *gpu.Window, camera Camera, time, swing float64) {
	if h.model == nil && !h.empty {
		return
	}

	window.ClearDepth()

	h.program.Use()
	h.program.Mat4("viewProjection", camera.Projection().Mul4(h.pose(time, swing)))
	h.program.Int("atlas", 0)

	if h.empty {
		h.skin.Bind(0)
		h.arm.Draw()

		return
	}

	if h.cube {
		h.blocks.Atlas().Bind(0)
	} else {
		h.icons.Atlas().Bind(0)
	}
	h.model.Draw()
}

func (h *Hand) pose(time, swing float64) mgl32.Mat4 {
	bob := float32(math.Sin(time*1.8)) * 0.015
	sway := float32(math.Sin(time*0.9)) * 0.01

	root := math.Sqrt(swing)
	punch := float32(math.Sin(root * math.Pi))
	whirl := float32(math.Sin(swing * swing * math.Pi))

	if h.empty {
		// the exact vanilla renderPlayerArm chain, swing terms included; the
		// last two transforms map our shoulder-up box onto the y-down biped
		// right arm cube
		reach := mgl32.Translate3D(
			-0.3*punch,
			0.4*float32(math.Sin(root*2*math.Pi)),
			-0.4*float32(math.Sin(swing*math.Pi)))

		return reach.
			Mul4(mgl32.Translate3D(0.64+sway, -0.6+bob, -0.72)).
			Mul4(mgl32.HomogRotate3DY(mgl32.DegToRad(45))).
			Mul4(mgl32.HomogRotate3DY(mgl32.DegToRad(punch * 70))).
			Mul4(mgl32.HomogRotate3DZ(mgl32.DegToRad(whirl * -20))).
			Mul4(mgl32.Translate3D(-1, 3.6, 3.5)).
			Mul4(mgl32.HomogRotate3DZ(mgl32.DegToRad(120))).
			Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(200))).
			Mul4(mgl32.HomogRotate3DY(mgl32.DegToRad(-135))).
			Mul4(mgl32.Translate3D(5.6, 0, 0)).
			Mul4(mgl32.Translate3D(-0.5, 0.75, -0.125)).
			Mul4(mgl32.Scale3D(1, -1, 1))
	}

	// vanilla item sway plus a downward chop for the swing
	chop := mgl32.Translate3D(
		-0.4*punch,
		0.2*float32(math.Sin(root*2*math.Pi)),
		-0.2*float32(math.Sin(swing*math.Pi)))

	center := mgl32.Translate3D(-0.5, -0.5, -0.5)
	if h.cube {
		return chop.
			Mul4(mgl32.Translate3D(0.45+sway, -0.42+bob, -0.85)).
			Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(punch * -40))).
			Mul4(mgl32.HomogRotate3DY(mgl32.DegToRad(-40))).
			Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(12))).
			Mul4(mgl32.Scale3D(0.3, 0.3, 0.3)).
			Mul4(center)
	}

	return chop.
		Mul4(mgl32.Translate3D(0.42+sway, -0.36+bob, -0.85)).
		Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(punch * -40))).
		Mul4(mgl32.HomogRotate3DY(mgl32.DegToRad(-30))).
		Mul4(mgl32.HomogRotate3DZ(mgl32.DegToRad(18))).
		Mul4(mgl32.Scale3D(0.42, 0.42, 0.42)).
		Mul4(center)
}

func (h *Hand) Close() {
	h.clear()

	if h.arm != nil {
		h.arm.Delete()
		h.arm = nil
	}
}

func (h *Hand) clear() {
	if h.model != nil {
		h.model.Delete()
		h.model = nil
	}
}
