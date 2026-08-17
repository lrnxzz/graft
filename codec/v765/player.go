package v765

import (
	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/codec"
)

type PlayerAbilities struct {
	Flags       codec.Byte
	FlyingSpeed codec.Float
	FieldOfView codec.Float
}

func (*PlayerAbilities) ID() int32 {
	return 0x36
}

func (*PlayerAbilities) Name() string {
	return "PlayerAbilities"
}

func (*PlayerAbilities) State() codec.State {
	return codec.StatePlay
}

func (*PlayerAbilities) Direction() codec.Direction {
	return codec.Clientbound
}

func (p PlayerAbilities) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.Flags, p.FlyingSpeed, p.FieldOfView)
}

func (p *PlayerAbilities) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.Flags, &p.FlyingSpeed, &p.FieldOfView)
}

const (
	abilityInvulnerable byte = 1 << iota
	abilityFlying
	abilityAllowFlight
	abilityInstantBuild
)

func (p *PlayerAbilities) Apply(player *graft.Player) {
	flags := byte(p.Flags)

	player.Abilities = graft.Abilities{
		Invulnerable: flags&abilityInvulnerable != 0,
		Flying:       flags&abilityFlying != 0,
		AllowFlight:  flags&abilityAllowFlight != 0,
		InstantBuild: flags&abilityInstantBuild != 0,
		FlySpeed:     p.FlyingSpeed.Float32(),
		FieldOfView:  p.FieldOfView.Float32(),
	}
}

type SetExperience struct {
	Bar             codec.Float
	Level           codec.VarInt
	TotalExperience codec.VarInt
}

func (*SetExperience) ID() int32 {
	return 0x5A
}

func (*SetExperience) Name() string {
	return "SetExperience"
}

func (*SetExperience) State() codec.State {
	return codec.StatePlay
}

func (*SetExperience) Direction() codec.Direction {
	return codec.Clientbound
}

func (p SetExperience) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.Bar, p.Level, p.TotalExperience)
}

func (p *SetExperience) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.Bar, &p.Level, &p.TotalExperience)
}

func (p *SetExperience) Apply(player *graft.Player) {
	player.Experience = p.Bar.Float32()
	player.Level = p.Level.Int32()
	player.TotalExperience = p.TotalExperience.Int32()
}

type SetHealth struct {
	Health     codec.Float
	Food       codec.VarInt
	Saturation codec.Float
}

func (*SetHealth) ID() int32 {
	return 0x5B
}

func (*SetHealth) Name() string {
	return "SetHealth"
}

func (*SetHealth) State() codec.State {
	return codec.StatePlay
}

func (*SetHealth) Direction() codec.Direction {
	return codec.Clientbound
}

func (p SetHealth) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.Health, p.Food, p.Saturation)
}

func (p *SetHealth) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.Health, &p.Food, &p.Saturation)
}

func (p *SetHealth) Apply(player *graft.Player) {
	player.Health = p.Health.Float32()
	player.Food = p.Food.Int32()
	player.Saturation = p.Saturation.Float32()
}
