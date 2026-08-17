package v765

import (
	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/codec"
)

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

func (p *SyncPlayerPosition) Apply(player *graft.Player) {
	flags := byte(p.Flags)

	target := graft.Vec3(p.X.Float64(), p.Y.Float64(), p.Z.Float64())
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
