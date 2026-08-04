package v765

import (
	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/codec"
)

type DeathLocation struct {
	DimensionName gocraft.Identifier
	Location      gocraft.Position
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
	Worlds              codec.Slice[gocraft.Identifier]
	MaxPlayers          codec.VarInt
	ViewDistance        codec.VarInt
	SimulationDistance  codec.VarInt
	ReducedDebugInfo    codec.Bool
	EnableRespawnScreen codec.Bool
	LimitedCrafting     codec.Bool
	DimensionType       gocraft.Identifier
	DimensionName       gocraft.Identifier
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

func (p *JoinGame) Apply(player *gocraft.Player) {
	player.EntityID = p.EntityID.Int32()
	player.GameMode = gocraft.GameMode(p.GameMode)
	player.Dimension = p.DimensionName
}

type PlayKeepAlive struct {
	KeepAliveID codec.Long
}

func (*PlayKeepAlive) ID() int32 {
	return 0x24
}

func (*PlayKeepAlive) Name() string {
	return "PlayKeepAlive"
}

func (*PlayKeepAlive) State() codec.State {
	return codec.StatePlay
}

func (*PlayKeepAlive) Direction() codec.Direction {
	return codec.Clientbound
}

func (p PlayKeepAlive) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.KeepAliveID)
}

func (p *PlayKeepAlive) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.KeepAliveID)
}

type PlayKeepAliveResponse struct {
	KeepAliveID codec.Long
}

func (*PlayKeepAliveResponse) ID() int32 {
	return 0x15
}

func (*PlayKeepAliveResponse) Name() string {
	return "PlayKeepAliveResponse"
}

func (*PlayKeepAliveResponse) State() codec.State {
	return codec.StatePlay
}

func (*PlayKeepAliveResponse) Direction() codec.Direction {
	return codec.Serverbound
}

func (p PlayKeepAliveResponse) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.KeepAliveID)
}

func (p *PlayKeepAliveResponse) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.KeepAliveID)
}

type SyncPlayerPosition struct {
	X          codec.Double
	Y          codec.Double
	Z          codec.Double
	Yaw        codec.Float
	Pitch      codec.Float
	Flags      codec.Byte
	TeleportID codec.VarInt
}

func (*SyncPlayerPosition) ID() int32 {
	return 0x3E
}

func (*SyncPlayerPosition) Name() string {
	return "SyncPlayerPosition"
}

func (*SyncPlayerPosition) State() codec.State {
	return codec.StatePlay
}

func (*SyncPlayerPosition) Direction() codec.Direction {
	return codec.Clientbound
}

func (p SyncPlayerPosition) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.X, p.Y, p.Z, p.Yaw, p.Pitch, p.Flags, p.TeleportID)
}

func (p *SyncPlayerPosition) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.X, &p.Y, &p.Z, &p.Yaw, &p.Pitch, &p.Flags, &p.TeleportID)
}

const (
	relativeX byte = 1 << iota
	relativeY
	relativeZ
	relativeYaw
	relativePitch
)

func (p *SyncPlayerPosition) Apply(player *gocraft.Player) {
	flags := byte(p.Flags)

	target := gocraft.Vec3(p.X.Float64(), p.Y.Float64(), p.Z.Float64())
	if flags&relativeX != 0 {
		target.X += player.Position.X
	}
	if flags&relativeY != 0 {
		target.Y += player.Position.Y
	}
	if flags&relativeZ != 0 {
		target.Z += player.Position.Z
	}

	yaw, pitch := p.Yaw.Float32(), p.Pitch.Float32()
	if flags&relativeYaw != 0 {
		yaw += player.Yaw
	}
	if flags&relativePitch != 0 {
		pitch += player.Pitch
	}

	player.Position = target
	player.Yaw = yaw
	player.Pitch = pitch
}

type ConfirmTeleport struct {
	TeleportID codec.VarInt
}

func (*ConfirmTeleport) ID() int32 {
	return 0x00
}

func (*ConfirmTeleport) Name() string {
	return "ConfirmTeleport"
}

func (*ConfirmTeleport) State() codec.State {
	return codec.StatePlay
}

func (*ConfirmTeleport) Direction() codec.Direction {
	return codec.Serverbound
}

func (p ConfirmTeleport) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.TeleportID)
}

func (p *ConfirmTeleport) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.TeleportID)
}

type PlayDisconnect struct {
	Reason codec.NBT
}

func (*PlayDisconnect) ID() int32 {
	return 0x1B
}

func (*PlayDisconnect) Name() string {
	return "PlayDisconnect"
}

func (*PlayDisconnect) State() codec.State {
	return codec.StatePlay
}

func (*PlayDisconnect) Direction() codec.Direction {
	return codec.Clientbound
}

func (p PlayDisconnect) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.Reason)
}

func (p *PlayDisconnect) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.Reason)
}

type ChunkData struct {
	X          codec.Int
	Z          codec.Int
	Heightmaps codec.NBT
	Sections   codec.Bytes
}

func (*ChunkData) ID() int32 {
	return 0x25
}

func (*ChunkData) Name() string {
	return "ChunkData"
}

func (*ChunkData) State() codec.State {
	return codec.StatePlay
}

func (*ChunkData) Direction() codec.Direction {
	return codec.Clientbound
}

func (p ChunkData) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.X, p.Z, p.Heightmaps, p.Sections)
}

func (p *ChunkData) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.X, &p.Z, &p.Heightmaps, &p.Sections)
}

func (p *ChunkData) Column(minY, height int) (*gocraft.Column, error) {
	column := gocraft.ChunkColumn(p.X.Int32(), p.Z.Int32(), minY, height)
	if err := column.Decode(codec.NewReader(p.Sections)); err != nil {
		return nil, err
	}

	return column, nil
}

type UnloadChunk struct {
	Z codec.Int
	X codec.Int
}

func (*UnloadChunk) ID() int32 {
	return 0x1F
}

func (*UnloadChunk) Name() string {
	return "UnloadChunk"
}

func (*UnloadChunk) State() codec.State {
	return codec.StatePlay
}

func (*UnloadChunk) Direction() codec.Direction {
	return codec.Clientbound
}

func (p UnloadChunk) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.Z, p.X)
}

func (p *UnloadChunk) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.Z, &p.X)
}

type BlockUpdate struct {
	Location gocraft.Position
	Block    codec.VarInt
}

func (*BlockUpdate) ID() int32 {
	return 0x09
}

func (*BlockUpdate) Name() string {
	return "BlockUpdate"
}

func (*BlockUpdate) State() codec.State {
	return codec.StatePlay
}

func (*BlockUpdate) Direction() codec.Direction {
	return codec.Clientbound
}

func (p BlockUpdate) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.Location, p.Block)
}

func (p *BlockUpdate) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.Location, &p.Block)
}

type BlockChange struct {
	X     int
	Y     int
	Z     int
	State gocraft.BlockState
}

func (p *BlockUpdate) Change() BlockChange {
	return BlockChange{
		X:     p.Location.X,
		Y:     p.Location.Y,
		Z:     p.Location.Z,
		State: gocraft.BlockState(p.Block),
	}
}

type SectionBlocksUpdate struct {
	Section codec.Long
	Packed  codec.Slice[codec.VarLong]
}

func (*SectionBlocksUpdate) ID() int32 {
	return 0x47
}

func (*SectionBlocksUpdate) Name() string {
	return "SectionBlocksUpdate"
}

func (*SectionBlocksUpdate) State() codec.State {
	return codec.StatePlay
}

func (*SectionBlocksUpdate) Direction() codec.Direction {
	return codec.Clientbound
}

func (p SectionBlocksUpdate) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.Section, p.Packed)
}

func (p *SectionBlocksUpdate) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.Section, &p.Packed)
}

func (p *SectionBlocksUpdate) Changes() []BlockChange {
	baseX := int(p.Section.Signed(42, 22)) * 16
	baseZ := int(p.Section.Signed(20, 22)) * 16
	baseY := int(p.Section.Signed(0, 20)) * 16

	changes := make([]BlockChange, len(p.Packed))
	for i, packed := range p.Packed {
		block := codec.Long(packed)
		changes[i] = BlockChange{
			X:     baseX + int(block.Unsigned(8, 4)),
			Y:     baseY + int(block.Unsigned(0, 4)),
			Z:     baseZ + int(block.Unsigned(4, 4)),
			State: gocraft.BlockState(block.Unsigned(12, 52)),
		}
	}

	return changes
}

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

func (p *PlayerAbilities) Apply(player *gocraft.Player) {
	flags := byte(p.Flags)

	player.Abilities = gocraft.Abilities{
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

func (p *SetExperience) Apply(player *gocraft.Player) {
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

func (p *SetHealth) Apply(player *gocraft.Player) {
	player.Health = p.Health.Float32()
	player.Food = p.Food.Int32()
	player.Saturation = p.Saturation.Float32()
}

type SetPlayerPosition struct {
	X        codec.Double
	Y        codec.Double
	Z        codec.Double
	OnGround codec.Bool
}

func (*SetPlayerPosition) ID() int32 {
	return 0x17
}

func (*SetPlayerPosition) Name() string {
	return "SetPlayerPosition"
}

func (*SetPlayerPosition) State() codec.State {
	return codec.StatePlay
}

func (*SetPlayerPosition) Direction() codec.Direction {
	return codec.Serverbound
}

func (p SetPlayerPosition) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.X, p.Y, p.Z, p.OnGround)
}

func (p *SetPlayerPosition) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.X, &p.Y, &p.Z, &p.OnGround)
}

type SetPlayerPositionRotation struct {
	X        codec.Double
	Y        codec.Double
	Z        codec.Double
	Yaw      codec.Float
	Pitch    codec.Float
	OnGround codec.Bool
}

func (*SetPlayerPositionRotation) ID() int32 {
	return 0x18
}

func (*SetPlayerPositionRotation) Name() string {
	return "SetPlayerPositionRotation"
}

func (*SetPlayerPositionRotation) State() codec.State {
	return codec.StatePlay
}

func (*SetPlayerPositionRotation) Direction() codec.Direction {
	return codec.Serverbound
}

func (p SetPlayerPositionRotation) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.X, p.Y, p.Z, p.Yaw, p.Pitch, p.OnGround)
}

func (p *SetPlayerPositionRotation) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.X, &p.Y, &p.Z, &p.Yaw, &p.Pitch, &p.OnGround)
}

type ChunkBatchStart struct{}

func (*ChunkBatchStart) ID() int32 {
	return 0x0D
}

func (*ChunkBatchStart) Name() string {
	return "ChunkBatchStart"
}

func (*ChunkBatchStart) State() codec.State {
	return codec.StatePlay
}

func (*ChunkBatchStart) Direction() codec.Direction {
	return codec.Clientbound
}

func (ChunkBatchStart) Append(dst []byte) []byte {
	return dst
}

func (*ChunkBatchStart) Decode(*codec.Reader) error {
	return nil
}

type ChunkBatchFinished struct {
	BatchSize codec.VarInt
}

func (*ChunkBatchFinished) ID() int32 {
	return 0x0C
}

func (*ChunkBatchFinished) Name() string {
	return "ChunkBatchFinished"
}

func (*ChunkBatchFinished) State() codec.State {
	return codec.StatePlay
}

func (*ChunkBatchFinished) Direction() codec.Direction {
	return codec.Clientbound
}

func (p ChunkBatchFinished) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.BatchSize)
}

func (p *ChunkBatchFinished) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.BatchSize)
}

type ChunkBatchReceived struct {
	ChunksPerTick codec.Float
}

func (*ChunkBatchReceived) ID() int32 {
	return 0x07
}

func (*ChunkBatchReceived) Name() string {
	return "ChunkBatchReceived"
}

func (*ChunkBatchReceived) State() codec.State {
	return codec.StatePlay
}

func (*ChunkBatchReceived) Direction() codec.Direction {
	return codec.Serverbound
}

func (p ChunkBatchReceived) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.ChunksPerTick)
}

func (p *ChunkBatchReceived) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.ChunksPerTick)
}
