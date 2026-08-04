package gocraft

import (
	"github.com/lrnxzz/go-craft/codec"

	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Position struct {
	X int
	Y int
	Z int
}

func At(x, y, z int) Position {
	return Position{
		X: x,
		Y: y,
		Z: z,
	}
}

var errPositionFormat = errors.New("gocraft: position must be written as x,y,z")

func ParsePosition(text string) (Position, error) {
	parts := strings.Split(text, ",")
	if len(parts) != 3 {
		return Position{}, errPositionFormat
	}

	var coordinates [3]int
	for index, part := range parts {
		value, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return Position{}, fmt.Errorf("gocraft: invalid position coordinate %q", part)
		}

		coordinates[index] = value
	}

	return At(coordinates[0], coordinates[1], coordinates[2]), nil
}

func (p Position) Append(dst []byte) []byte {
	packed := int64(p.X&0x3FFFFFF)<<38 | int64(p.Z&0x3FFFFFF)<<12 | int64(p.Y&0xFFF)

	return codec.Long(packed).Append(dst)
}

func (p *Position) Decode(r *codec.Reader) error {
	var packed codec.Long
	if err := packed.Decode(r); err != nil {
		return err
	}

	p.X = int(packed.Signed(38, 26))
	p.Y = int(packed.Signed(0, 12))
	p.Z = int(packed.Signed(12, 26))

	return nil
}

func (p Position) Add(dx, dy, dz int) Position {
	return At(p.X+dx, p.Y+dy, p.Z+dz)
}

func (p Position) Neighbor(face BlockFace) Position {
	offset := face.Offset()

	return p.Add(offset.X, offset.Y, offset.Z)
}

func (p Position) Chunk() ChunkPos {
	return Chunk(int32(p.X>>4), int32(p.Z>>4))
}

func (p Position) Horizontal() Vec2d {
	return Vec2(float64(p.X), float64(p.Z))
}

func (p Position) Corner() Vec3d {
	return Vec3(float64(p.X), float64(p.Y), float64(p.Z))
}

func (p Position) Center() Vec3d {
	return Vec3(float64(p.X)+0.5, float64(p.Y)+0.5, float64(p.Z)+0.5)
}

func (p Position) String() string {
	return fmt.Sprintf("(%d, %d, %d)", p.X, p.Y, p.Z)
}

type BlockFace uint8

const (
	FaceDown BlockFace = iota
	FaceUp
	FaceNorth
	FaceSouth
	FaceWest
	FaceEast
)

var faceOffsets = [...]Position{
	FaceDown:  {Y: -1},
	FaceUp:    {Y: 1},
	FaceNorth: {Z: -1},
	FaceSouth: {Z: 1},
	FaceWest:  {X: -1},
	FaceEast:  {X: 1},
}

var faceNames = [...]string{
	FaceDown:  "down",
	FaceUp:    "up",
	FaceNorth: "north",
	FaceSouth: "south",
	FaceWest:  "west",
	FaceEast:  "east",
}

func (f BlockFace) Offset() Position {
	if int(f) >= len(faceOffsets) {
		return Position{}
	}

	return faceOffsets[f]
}

func (f BlockFace) Normal() Vec3d {
	return f.Offset().Corner()
}

func (f BlockFace) Opposite() BlockFace {
	return f ^ 1
}

func (f BlockFace) String() string {
	if int(f) >= len(faceNames) {
		return fmt.Sprintf("face(%d)", uint8(f))
	}

	return faceNames[f]
}

type Angle uint8

func (a Angle) Append(dst []byte) []byte {
	return codec.UByte(a).Append(dst)
}

func (a *Angle) Decode(r *codec.Reader) error {
	var raw codec.UByte
	if err := raw.Decode(r); err != nil {
		return err
	}

	*a = Angle(raw)

	return nil
}

func (a Angle) Degrees() float64 {
	return float64(a) * 360 / 256
}

func (a Angle) Radians() float64 {
	return float64(a) * 2 * math.Pi / 256
}

func AngleFromDegrees(degrees float64) Angle {
	return Angle(int64(math.Round(degrees / 360 * 256)))
}

func AngleFromRadians(radians float64) Angle {
	return Angle(int64(math.Round(radians / (2 * math.Pi) * 256)))
}
