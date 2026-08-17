package v765

import (
	"github.com/lrnxzz/graft/codec"
)

type SetContainerContent struct {
	WindowID codec.UByte
	StateID  codec.VarInt
	Slots    codec.Slice[Slot]
	Carried  Slot
}

func (*SetContainerContent) ID() int32 {
	return 0x13
}

func (*SetContainerContent) Name() string {
	return "SetContainerContent"
}

func (*SetContainerContent) State() codec.State {
	return codec.StatePlay
}

func (*SetContainerContent) Direction() codec.Direction {
	return codec.Clientbound
}

func (p SetContainerContent) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.WindowID, p.StateID, p.Slots, p.Carried)
}

func (p *SetContainerContent) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.WindowID, &p.StateID, &p.Slots, &p.Carried)
}

type SetContainerSlot struct {
	WindowID codec.Byte
	StateID  codec.VarInt
	Index    codec.Short
	Data     Slot
}

func (*SetContainerSlot) ID() int32 {
	return 0x15
}

func (*SetContainerSlot) Name() string {
	return "SetContainerSlot"
}

func (*SetContainerSlot) State() codec.State {
	return codec.StatePlay
}

func (*SetContainerSlot) Direction() codec.Direction {
	return codec.Clientbound
}

func (p SetContainerSlot) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.WindowID, p.StateID, p.Index, p.Data)
}

func (p *SetContainerSlot) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.WindowID, &p.StateID, &p.Index, &p.Data)
}

type SetHeldItem struct {
	Slot codec.Byte
}

func (*SetHeldItem) ID() int32 {
	return 0x51
}

func (*SetHeldItem) Name() string {
	return "SetHeldItem"
}

func (*SetHeldItem) State() codec.State {
	return codec.StatePlay
}

func (*SetHeldItem) Direction() codec.Direction {
	return codec.Clientbound
}

func (p SetHeldItem) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.Slot)
}

func (p *SetHeldItem) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.Slot)
}

type HeldItemChange struct {
	Slot codec.Short
}

func (*HeldItemChange) ID() int32 {
	return 0x2C
}

func (*HeldItemChange) Name() string {
	return "HeldItemChange"
}

func (*HeldItemChange) State() codec.State {
	return codec.StatePlay
}

func (*HeldItemChange) Direction() codec.Direction {
	return codec.Serverbound
}

func (p HeldItemChange) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.Slot)
}

func (p *HeldItemChange) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.Slot)
}

const (
	clickPickup   codec.VarInt = 0
	clickSwap     codec.VarInt = 2
	offhandButton codec.Byte   = 40
)

type ClickContainer struct {
	WindowID codec.UByte
	StateID  codec.VarInt
	Index    codec.Short
	Button   codec.Byte
	Mode     codec.VarInt
	Changed  codec.Slice[ChangedSlot]
	Carried  Slot
}

func (*ClickContainer) ID() int32 {
	return 0x0D
}

func (*ClickContainer) Name() string {
	return "ClickContainer"
}

func (*ClickContainer) State() codec.State {
	return codec.StatePlay
}

func (*ClickContainer) Direction() codec.Direction {
	return codec.Serverbound
}

func (p ClickContainer) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.WindowID, p.StateID, p.Index, p.Button, p.Mode, p.Changed, p.Carried)
}

func (p *ClickContainer) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.WindowID, &p.StateID, &p.Index, &p.Button, &p.Mode, &p.Changed, &p.Carried)
}

type CloseContainer struct {
	WindowID codec.UByte
}

func (*CloseContainer) ID() int32 {
	return 0x0E
}

func (*CloseContainer) Name() string {
	return "CloseContainer"
}

func (*CloseContainer) State() codec.State {
	return codec.StatePlay
}

func (*CloseContainer) Direction() codec.Direction {
	return codec.Serverbound
}

func (p CloseContainer) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.WindowID)
}

func (p *CloseContainer) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.WindowID)
}
