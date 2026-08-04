package v765

import (
	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/codec"
)

type LoginStart struct {
	Username codec.String
	UUID     codec.UUID
}

func (*LoginStart) ID() int32 {
	return 0x00
}

func (*LoginStart) Name() string {
	return "LoginStart"
}

func (*LoginStart) State() codec.State {
	return codec.StateLogin
}

func (*LoginStart) Direction() codec.Direction {
	return codec.Serverbound
}

func (p LoginStart) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.Username, p.UUID)
}

func (p *LoginStart) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.Username, &p.UUID)
}

type EncryptionBegin struct {
	ServerID    codec.String
	PublicKey   codec.Bytes
	VerifyToken codec.Bytes
}

func (*EncryptionBegin) ID() int32 {
	return 0x01
}

func (*EncryptionBegin) Name() string {
	return "EncryptionBegin"
}

func (*EncryptionBegin) State() codec.State {
	return codec.StateLogin
}

func (*EncryptionBegin) Direction() codec.Direction {
	return codec.Clientbound
}

func (p EncryptionBegin) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.ServerID, p.PublicKey, p.VerifyToken)
}

func (p *EncryptionBegin) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.ServerID, &p.PublicKey, &p.VerifyToken)
}

type LoginAcknowledged struct{}

func (*LoginAcknowledged) ID() int32 {
	return 0x03
}

func (*LoginAcknowledged) Name() string {
	return "LoginAcknowledged"
}

func (*LoginAcknowledged) State() codec.State {
	return codec.StateLogin
}

func (*LoginAcknowledged) Direction() codec.Direction {
	return codec.Serverbound
}

func (LoginAcknowledged) Append(dst []byte) []byte {
	return dst
}

func (*LoginAcknowledged) Decode(*codec.Reader) error {
	return nil
}

type Property struct {
	Name      codec.String
	Value     codec.String
	Signature codec.Option[codec.String]
}

func (p Property) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.Name, p.Value, p.Signature)
}

func (p *Property) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.Name, &p.Value, &p.Signature)
}

type LoginSuccess struct {
	UUID       codec.UUID
	Username   codec.String
	Properties codec.Slice[Property]
}

func (*LoginSuccess) ID() int32 {
	return 0x02
}

func (*LoginSuccess) Name() string {
	return "LoginSuccess"
}

func (*LoginSuccess) State() codec.State {
	return codec.StateLogin
}

func (*LoginSuccess) Direction() codec.Direction {
	return codec.Clientbound
}

func (p LoginSuccess) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.UUID, p.Username, p.Properties)
}

func (p *LoginSuccess) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.UUID, &p.Username, &p.Properties)
}

func (p *LoginSuccess) Apply(player *gocraft.Player) {
	player.UUID = p.UUID
	player.Username = p.Username.String()
}

type SetCompression struct {
	Threshold codec.VarInt
}

func (*SetCompression) ID() int32 {
	return 0x03
}

func (*SetCompression) Name() string {
	return "SetCompression"
}

func (*SetCompression) State() codec.State {
	return codec.StateLogin
}

func (*SetCompression) Direction() codec.Direction {
	return codec.Clientbound
}

func (p SetCompression) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.Threshold)
}

func (p *SetCompression) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.Threshold)
}

type LoginDisconnect struct {
	Reason codec.String
}

func (*LoginDisconnect) ID() int32 {
	return 0x00
}

func (*LoginDisconnect) Name() string {
	return "LoginDisconnect"
}

func (*LoginDisconnect) State() codec.State {
	return codec.StateLogin
}

func (*LoginDisconnect) Direction() codec.Direction {
	return codec.Clientbound
}

func (p LoginDisconnect) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.Reason)
}

func (p *LoginDisconnect) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.Reason)
}
