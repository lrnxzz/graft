package v765

import gocraft "github.com/lrnxzz/go-craft"

type DeathLocation struct {
	DimensionName gocraft.Identifier
	Location      gocraft.Position
}

func (p DeathLocation) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.DimensionName, p.Location)
}

func (p *DeathLocation) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.DimensionName, &p.Location)
}

type JoinGame struct {
	EntityID            gocraft.Int
	Hardcore            gocraft.Bool
	Worlds              gocraft.Slice[gocraft.Identifier]
	MaxPlayers          gocraft.VarInt
	ViewDistance        gocraft.VarInt
	SimulationDistance  gocraft.VarInt
	ReducedDebugInfo    gocraft.Bool
	EnableRespawnScreen gocraft.Bool
	LimitedCrafting     gocraft.Bool
	DimensionType       gocraft.Identifier
	DimensionName       gocraft.Identifier
	HashedSeed          gocraft.Long
	GameMode            gocraft.UByte
	PreviousGameMode    gocraft.Byte
	Debug               gocraft.Bool
	Flat                gocraft.Bool
	Death               gocraft.Option[DeathLocation]
	PortalCooldown      gocraft.VarInt
}

func (*JoinGame) ID() int32 {
	return 0x29
}

func (*JoinGame) Name() string {
	return "JoinGame"
}

func (*JoinGame) State() gocraft.State {
	return gocraft.StatePlay
}

func (*JoinGame) Direction() gocraft.Direction {
	return gocraft.Clientbound
}

func (p JoinGame) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.EntityID, p.Hardcore, p.Worlds, p.MaxPlayers, p.ViewDistance,
		p.SimulationDistance, p.ReducedDebugInfo, p.EnableRespawnScreen, p.LimitedCrafting,
		p.DimensionType, p.DimensionName, p.HashedSeed, p.GameMode, p.PreviousGameMode,
		p.Debug, p.Flat, p.Death, p.PortalCooldown)
}

func (p *JoinGame) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.EntityID, &p.Hardcore, &p.Worlds, &p.MaxPlayers, &p.ViewDistance,
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
	KeepAliveID gocraft.Long
}

func (*PlayKeepAlive) ID() int32 {
	return 0x24
}

func (*PlayKeepAlive) Name() string {
	return "PlayKeepAlive"
}

func (*PlayKeepAlive) State() gocraft.State {
	return gocraft.StatePlay
}

func (*PlayKeepAlive) Direction() gocraft.Direction {
	return gocraft.Clientbound
}

func (p PlayKeepAlive) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.KeepAliveID)
}

func (p *PlayKeepAlive) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.KeepAliveID)
}

type PlayKeepAliveResponse struct {
	KeepAliveID gocraft.Long
}

func (*PlayKeepAliveResponse) ID() int32 {
	return 0x15
}

func (*PlayKeepAliveResponse) Name() string {
	return "PlayKeepAliveResponse"
}

func (*PlayKeepAliveResponse) State() gocraft.State {
	return gocraft.StatePlay
}

func (*PlayKeepAliveResponse) Direction() gocraft.Direction {
	return gocraft.Serverbound
}

func (p PlayKeepAliveResponse) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.KeepAliveID)
}

func (p *PlayKeepAliveResponse) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.KeepAliveID)
}

type SyncPlayerPosition struct {
	X          gocraft.Double
	Y          gocraft.Double
	Z          gocraft.Double
	Yaw        gocraft.Float
	Pitch      gocraft.Float
	Flags      gocraft.Byte
	TeleportID gocraft.VarInt
}

func (*SyncPlayerPosition) ID() int32 {
	return 0x3E
}

func (*SyncPlayerPosition) Name() string {
	return "SyncPlayerPosition"
}

func (*SyncPlayerPosition) State() gocraft.State {
	return gocraft.StatePlay
}

func (*SyncPlayerPosition) Direction() gocraft.Direction {
	return gocraft.Clientbound
}

func (p SyncPlayerPosition) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.X, p.Y, p.Z, p.Yaw, p.Pitch, p.Flags, p.TeleportID)
}

func (p *SyncPlayerPosition) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.X, &p.Y, &p.Z, &p.Yaw, &p.Pitch, &p.Flags, &p.TeleportID)
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
	TeleportID gocraft.VarInt
}

func (*ConfirmTeleport) ID() int32 {
	return 0x00
}

func (*ConfirmTeleport) Name() string {
	return "ConfirmTeleport"
}

func (*ConfirmTeleport) State() gocraft.State {
	return gocraft.StatePlay
}

func (*ConfirmTeleport) Direction() gocraft.Direction {
	return gocraft.Serverbound
}

func (p ConfirmTeleport) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.TeleportID)
}

func (p *ConfirmTeleport) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.TeleportID)
}

type PlayDisconnect struct {
	Reason gocraft.NBT
}

func (*PlayDisconnect) ID() int32 {
	return 0x1B
}

func (*PlayDisconnect) Name() string {
	return "PlayDisconnect"
}

func (*PlayDisconnect) State() gocraft.State {
	return gocraft.StatePlay
}

func (*PlayDisconnect) Direction() gocraft.Direction {
	return gocraft.Clientbound
}

func (p PlayDisconnect) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.Reason)
}

func (p *PlayDisconnect) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.Reason)
}

type ChunkData struct {
	X          gocraft.Int
	Z          gocraft.Int
	Heightmaps gocraft.NBT
	Sections   gocraft.Bytes
}

func (*ChunkData) ID() int32 {
	return 0x25
}

func (*ChunkData) Name() string {
	return "ChunkData"
}

func (*ChunkData) State() gocraft.State {
	return gocraft.StatePlay
}

func (*ChunkData) Direction() gocraft.Direction {
	return gocraft.Clientbound
}

func (p ChunkData) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.X, p.Z, p.Heightmaps, p.Sections)
}

func (p *ChunkData) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.X, &p.Z, &p.Heightmaps, &p.Sections)
}

func (p *ChunkData) Column(minY, height int) (*gocraft.Column, error) {
	column := gocraft.ChunkColumn(p.X.Int32(), p.Z.Int32(), minY, height)
	if err := column.Decode(gocraft.NewReader(p.Sections)); err != nil {
		return nil, err
	}

	return column, nil
}

type UnloadChunk struct {
	Z gocraft.Int
	X gocraft.Int
}

func (*UnloadChunk) ID() int32 {
	return 0x1F
}

func (*UnloadChunk) Name() string {
	return "UnloadChunk"
}

func (*UnloadChunk) State() gocraft.State {
	return gocraft.StatePlay
}

func (*UnloadChunk) Direction() gocraft.Direction {
	return gocraft.Clientbound
}

func (p UnloadChunk) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.Z, p.X)
}

func (p *UnloadChunk) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.Z, &p.X)
}

type BlockUpdate struct {
	Location gocraft.Position
	Block    gocraft.VarInt
}

func (*BlockUpdate) ID() int32 {
	return 0x09
}

func (*BlockUpdate) Name() string {
	return "BlockUpdate"
}

func (*BlockUpdate) State() gocraft.State {
	return gocraft.StatePlay
}

func (*BlockUpdate) Direction() gocraft.Direction {
	return gocraft.Clientbound
}

func (p BlockUpdate) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.Location, p.Block)
}

func (p *BlockUpdate) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.Location, &p.Block)
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
	Section gocraft.Long
	Packed  gocraft.Slice[gocraft.VarLong]
}

func (*SectionBlocksUpdate) ID() int32 {
	return 0x47
}

func (*SectionBlocksUpdate) Name() string {
	return "SectionBlocksUpdate"
}

func (*SectionBlocksUpdate) State() gocraft.State {
	return gocraft.StatePlay
}

func (*SectionBlocksUpdate) Direction() gocraft.Direction {
	return gocraft.Clientbound
}

func (p SectionBlocksUpdate) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.Section, p.Packed)
}

func (p *SectionBlocksUpdate) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.Section, &p.Packed)
}

func (p *SectionBlocksUpdate) Changes() []BlockChange {
	baseX := int(p.Section.Signed(42, 22)) * 16
	baseZ := int(p.Section.Signed(20, 22)) * 16
	baseY := int(p.Section.Signed(0, 20)) * 16

	changes := make([]BlockChange, len(p.Packed))
	for i, packed := range p.Packed {
		block := gocraft.Long(packed)
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
	Flags       gocraft.Byte
	FlyingSpeed gocraft.Float
	FieldOfView gocraft.Float
}

func (*PlayerAbilities) ID() int32 {
	return 0x36
}

func (*PlayerAbilities) Name() string {
	return "PlayerAbilities"
}

func (*PlayerAbilities) State() gocraft.State {
	return gocraft.StatePlay
}

func (*PlayerAbilities) Direction() gocraft.Direction {
	return gocraft.Clientbound
}

func (p PlayerAbilities) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.Flags, p.FlyingSpeed, p.FieldOfView)
}

func (p *PlayerAbilities) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.Flags, &p.FlyingSpeed, &p.FieldOfView)
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
	Bar             gocraft.Float
	Level           gocraft.VarInt
	TotalExperience gocraft.VarInt
}

func (*SetExperience) ID() int32 {
	return 0x5A
}

func (*SetExperience) Name() string {
	return "SetExperience"
}

func (*SetExperience) State() gocraft.State {
	return gocraft.StatePlay
}

func (*SetExperience) Direction() gocraft.Direction {
	return gocraft.Clientbound
}

func (p SetExperience) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.Bar, p.Level, p.TotalExperience)
}

func (p *SetExperience) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.Bar, &p.Level, &p.TotalExperience)
}

func (p *SetExperience) Apply(player *gocraft.Player) {
	player.Experience = p.Bar.Float32()
	player.Level = p.Level.Int32()
	player.TotalExperience = p.TotalExperience.Int32()
}

type SetHealth struct {
	Health     gocraft.Float
	Food       gocraft.VarInt
	Saturation gocraft.Float
}

func (*SetHealth) ID() int32 {
	return 0x5B
}

func (*SetHealth) Name() string {
	return "SetHealth"
}

func (*SetHealth) State() gocraft.State {
	return gocraft.StatePlay
}

func (*SetHealth) Direction() gocraft.Direction {
	return gocraft.Clientbound
}

func (p SetHealth) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.Health, p.Food, p.Saturation)
}

func (p *SetHealth) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.Health, &p.Food, &p.Saturation)
}

func (p *SetHealth) Apply(player *gocraft.Player) {
	player.Health = p.Health.Float32()
	player.Food = p.Food.Int32()
	player.Saturation = p.Saturation.Float32()
}

type SetPlayerPosition struct {
	X        gocraft.Double
	Y        gocraft.Double
	Z        gocraft.Double
	OnGround gocraft.Bool
}

func (*SetPlayerPosition) ID() int32 {
	return 0x17
}

func (*SetPlayerPosition) Name() string {
	return "SetPlayerPosition"
}

func (*SetPlayerPosition) State() gocraft.State {
	return gocraft.StatePlay
}

func (*SetPlayerPosition) Direction() gocraft.Direction {
	return gocraft.Serverbound
}

func (p SetPlayerPosition) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.X, p.Y, p.Z, p.OnGround)
}

func (p *SetPlayerPosition) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.X, &p.Y, &p.Z, &p.OnGround)
}

type SetPlayerPositionRotation struct {
	X        gocraft.Double
	Y        gocraft.Double
	Z        gocraft.Double
	Yaw      gocraft.Float
	Pitch    gocraft.Float
	OnGround gocraft.Bool
}

func (*SetPlayerPositionRotation) ID() int32 {
	return 0x18
}

func (*SetPlayerPositionRotation) Name() string {
	return "SetPlayerPositionRotation"
}

func (*SetPlayerPositionRotation) State() gocraft.State {
	return gocraft.StatePlay
}

func (*SetPlayerPositionRotation) Direction() gocraft.Direction {
	return gocraft.Serverbound
}

func (p SetPlayerPositionRotation) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.X, p.Y, p.Z, p.Yaw, p.Pitch, p.OnGround)
}

func (p *SetPlayerPositionRotation) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.X, &p.Y, &p.Z, &p.Yaw, &p.Pitch, &p.OnGround)
}

type ChunkBatchStart struct{}

func (*ChunkBatchStart) ID() int32 {
	return 0x0D
}

func (*ChunkBatchStart) Name() string {
	return "ChunkBatchStart"
}

func (*ChunkBatchStart) State() gocraft.State {
	return gocraft.StatePlay
}

func (*ChunkBatchStart) Direction() gocraft.Direction {
	return gocraft.Clientbound
}

func (ChunkBatchStart) Append(dst []byte) []byte {
	return dst
}

func (*ChunkBatchStart) Decode(*gocraft.Reader) error {
	return nil
}

type ChunkBatchFinished struct {
	BatchSize gocraft.VarInt
}

func (*ChunkBatchFinished) ID() int32 {
	return 0x0C
}

func (*ChunkBatchFinished) Name() string {
	return "ChunkBatchFinished"
}

func (*ChunkBatchFinished) State() gocraft.State {
	return gocraft.StatePlay
}

func (*ChunkBatchFinished) Direction() gocraft.Direction {
	return gocraft.Clientbound
}

func (p ChunkBatchFinished) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.BatchSize)
}

func (p *ChunkBatchFinished) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.BatchSize)
}

type ChunkBatchReceived struct {
	ChunksPerTick gocraft.Float
}

func (*ChunkBatchReceived) ID() int32 {
	return 0x07
}

func (*ChunkBatchReceived) Name() string {
	return "ChunkBatchReceived"
}

func (*ChunkBatchReceived) State() gocraft.State {
	return gocraft.StatePlay
}

func (*ChunkBatchReceived) Direction() gocraft.Direction {
	return gocraft.Serverbound
}

func (p ChunkBatchReceived) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.ChunksPerTick)
}

func (p *ChunkBatchReceived) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.ChunksPerTick)
}
