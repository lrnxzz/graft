package v765

import gocraft "github.com/lrnxzz/go-craft"

type SetContainerContent struct {
	WindowID gocraft.UByte
	StateID  gocraft.VarInt
	Slots    gocraft.Slice[Slot]
	Carried  Slot
}

func (*SetContainerContent) ID() int32 {
	return 0x13
}

func (*SetContainerContent) Name() string {
	return "SetContainerContent"
}

func (*SetContainerContent) State() gocraft.State {
	return gocraft.StatePlay
}

func (*SetContainerContent) Direction() gocraft.Direction {
	return gocraft.Clientbound
}

func (p SetContainerContent) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.WindowID, p.StateID, p.Slots, p.Carried)
}

func (p *SetContainerContent) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.WindowID, &p.StateID, &p.Slots, &p.Carried)
}

type SetContainerSlot struct {
	WindowID gocraft.Byte
	StateID  gocraft.VarInt
	Index    gocraft.Short
	Data     Slot
}

func (*SetContainerSlot) ID() int32 {
	return 0x15
}

func (*SetContainerSlot) Name() string {
	return "SetContainerSlot"
}

func (*SetContainerSlot) State() gocraft.State {
	return gocraft.StatePlay
}

func (*SetContainerSlot) Direction() gocraft.Direction {
	return gocraft.Clientbound
}

func (p SetContainerSlot) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.WindowID, p.StateID, p.Index, p.Data)
}

func (p *SetContainerSlot) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.WindowID, &p.StateID, &p.Index, &p.Data)
}

type SetHeldItem struct {
	Slot gocraft.Byte
}

func (*SetHeldItem) ID() int32 {
	return 0x51
}

func (*SetHeldItem) Name() string {
	return "SetHeldItem"
}

func (*SetHeldItem) State() gocraft.State {
	return gocraft.StatePlay
}

func (*SetHeldItem) Direction() gocraft.Direction {
	return gocraft.Clientbound
}

func (p SetHeldItem) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.Slot)
}

func (p *SetHeldItem) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.Slot)
}

type HeldItemChange struct {
	Slot gocraft.Short
}

func (*HeldItemChange) ID() int32 {
	return 0x2C
}

func (*HeldItemChange) Name() string {
	return "HeldItemChange"
}

func (*HeldItemChange) State() gocraft.State {
	return gocraft.StatePlay
}

func (*HeldItemChange) Direction() gocraft.Direction {
	return gocraft.Serverbound
}

func (p HeldItemChange) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.Slot)
}

func (p *HeldItemChange) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.Slot)
}

const (
	clickPickup   gocraft.VarInt = 0
	clickSwap     gocraft.VarInt = 2
	offhandButton gocraft.Byte   = 40
)

type ClickContainer struct {
	WindowID gocraft.UByte
	StateID  gocraft.VarInt
	Index    gocraft.Short
	Button   gocraft.Byte
	Mode     gocraft.VarInt
	Changed  gocraft.Slice[ChangedSlot]
	Carried  Slot
}

func (*ClickContainer) ID() int32 {
	return 0x0D
}

func (*ClickContainer) Name() string {
	return "ClickContainer"
}

func (*ClickContainer) State() gocraft.State {
	return gocraft.StatePlay
}

func (*ClickContainer) Direction() gocraft.Direction {
	return gocraft.Serverbound
}

func (p ClickContainer) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.WindowID, p.StateID, p.Index, p.Button, p.Mode, p.Changed, p.Carried)
}

func (p *ClickContainer) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.WindowID, &p.StateID, &p.Index, &p.Button, &p.Mode, &p.Changed, &p.Carried)
}

type CloseContainer struct {
	WindowID gocraft.UByte
}

func (*CloseContainer) ID() int32 {
	return 0x0E
}

func (*CloseContainer) Name() string {
	return "CloseContainer"
}

func (*CloseContainer) State() gocraft.State {
	return gocraft.StatePlay
}

func (*CloseContainer) Direction() gocraft.Direction {
	return gocraft.Serverbound
}

func (p CloseContainer) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.WindowID)
}

func (p *CloseContainer) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.WindowID)
}
