package v765

import "github.com/lrnxzz/graft/codec"

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
