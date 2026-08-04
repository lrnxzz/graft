package gocraft

import (
	"github.com/lrnxzz/go-craft/codec"

	"fmt"
	"math"
)

type GameMode uint8

const (
	Survival GameMode = iota
	Creative
	Adventure
	Spectator
)

func (m GameMode) String() string {
	switch m {
	case Survival:
		return "survival"
	case Creative:
		return "creative"
	case Adventure:
		return "adventure"
	case Spectator:
		return "spectator"
	}

	return fmt.Sprintf("gamemode(%d)", uint8(m))
}

const (
	playerWidth  = 0.6
	playerHeight = 1.8

	EyeHeight  = 1.62
	BlockReach = 4.5
)

type Abilities struct {
	Invulnerable bool
	Flying       bool
	AllowFlight  bool
	InstantBuild bool
	FlySpeed     float32
	FieldOfView  float32
}

type Player struct {
	EntityID        int32
	UUID            codec.UUID
	Username        string
	Position        Vec3d
	Yaw             float32
	Pitch           float32
	OnGround        bool
	Health          float32
	Food            int32
	Saturation      float32
	Experience      float32
	Level           int32
	TotalExperience int32
	GameMode        GameMode
	Dimension       Identifier
	Abilities       Abilities
}

func (p *Player) Eye() Vec3d {
	return p.Position.Offset(0, EyeHeight, 0)
}

func (p *Player) LookDirection() Vec3d {
	yaw := float64(p.Yaw) * math.Pi / 180
	pitch := float64(p.Pitch) * math.Pi / 180

	return Vec3(
		-math.Sin(yaw)*math.Cos(pitch),
		-math.Sin(pitch),
		math.Cos(yaw)*math.Cos(pitch),
	)
}

func (p *Player) Box() AABB {
	return BoxAround(p.Position, playerWidth, playerHeight)
}

func LookAngles(from, to Vec3d) (float32, float32) {
	delta := to.Sub(from)

	yaw := math.Atan2(-delta.X, delta.Z) * 180 / math.Pi
	pitch := -math.Atan2(delta.Y, delta.HorizontalLength()) * 180 / math.Pi

	return float32(yaw), float32(pitch)
}

func (p *Player) Alive() bool {
	return p.Health > 0
}
