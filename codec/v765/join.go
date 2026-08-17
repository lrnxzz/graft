package v765

import (
	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/codec"
)

type DeathLocation struct {
	DimensionName graft.Identifier
	Location      graft.Position
}

func (p DeathLocation) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.DimensionName, p.Location)
}

func (p *DeathLocation) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.DimensionName, &p.Location)
}

type JoinGame struct {
	EntityID            codec.Int
	Hardcore            codec.Bool
	Worlds              codec.Slice[graft.Identifier]
	MaxPlayers          codec.VarInt
	ViewDistance        codec.VarInt
	SimulationDistance  codec.VarInt
	ReducedDebugInfo    codec.Bool
	EnableRespawnScreen codec.Bool
	LimitedCrafting     codec.Bool
	DimensionType       graft.Identifier
	DimensionName       graft.Identifier
	HashedSeed          codec.Long
	GameMode            codec.UByte
	PreviousGameMode    codec.Byte
	Debug               codec.Bool
	Flat                codec.Bool
	Death               codec.Option[DeathLocation]
	PortalCooldown      codec.VarInt
}

func (*JoinGame) ID() int32 {
	return 0x29
}

func (*JoinGame) Name() string {
	return "JoinGame"
}

func (*JoinGame) State() codec.State {
	return codec.StatePlay
}

func (*JoinGame) Direction() codec.Direction {
	return codec.Clientbound
}

func (p JoinGame) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.EntityID, p.Hardcore, p.Worlds, p.MaxPlayers, p.ViewDistance,
		p.SimulationDistance, p.ReducedDebugInfo, p.EnableRespawnScreen, p.LimitedCrafting,
		p.DimensionType, p.DimensionName, p.HashedSeed, p.GameMode, p.PreviousGameMode,
		p.Debug, p.Flat, p.Death, p.PortalCooldown)
}

func (p *JoinGame) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.EntityID, &p.Hardcore, &p.Worlds, &p.MaxPlayers, &p.ViewDistance,
		&p.SimulationDistance, &p.ReducedDebugInfo, &p.EnableRespawnScreen, &p.LimitedCrafting,
		&p.DimensionType, &p.DimensionName, &p.HashedSeed, &p.GameMode, &p.PreviousGameMode,
		&p.Debug, &p.Flat, &p.Death, &p.PortalCooldown)
}

func (p *JoinGame) Apply(player *graft.Player) {
	player.EntityID = p.EntityID.Int32()
	player.GameMode = graft.GameMode(p.GameMode)
	player.Dimension = p.DimensionName
}
