package v765

import (
	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/codec"
)

type ClientInformation struct {
	Locale              codec.String
	ViewDistance        codec.Byte
	ChatMode            codec.VarInt
	ChatColors          codec.Bool
	DisplayedSkinParts  codec.UByte
	MainHand            codec.VarInt
	EnableTextFiltering codec.Bool
	EnableServerListing codec.Bool
}

func (*ClientInformation) ID() int32 {
	return 0x00
}

func (*ClientInformation) Name() string {
	return "ClientInformation"
}

func (*ClientInformation) State() codec.State {
	return codec.StateConfiguration
}

func (*ClientInformation) Direction() codec.Direction {
	return codec.Serverbound
}

func (p ClientInformation) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.Locale, p.ViewDistance, p.ChatMode, p.ChatColors,
		p.DisplayedSkinParts, p.MainHand, p.EnableTextFiltering, p.EnableServerListing)
}

func (p *ClientInformation) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.Locale, &p.ViewDistance, &p.ChatMode, &p.ChatColors,
		&p.DisplayedSkinParts, &p.MainHand, &p.EnableTextFiltering, &p.EnableServerListing)
}

type ConfigDisconnect struct {
	Reason codec.NBT
}

func (*ConfigDisconnect) ID() int32 {
	return 0x01
}

func (*ConfigDisconnect) Name() string {
	return "ConfigDisconnect"
}

func (*ConfigDisconnect) State() codec.State {
	return codec.StateConfiguration
}

func (*ConfigDisconnect) Direction() codec.Direction {
	return codec.Clientbound
}

func (p ConfigDisconnect) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.Reason)
}

func (p *ConfigDisconnect) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.Reason)
}

type FinishConfiguration struct{}

func (*FinishConfiguration) ID() int32 {
	return 0x02
}

func (*FinishConfiguration) Name() string {
	return "FinishConfiguration"
}

func (*FinishConfiguration) State() codec.State {
	return codec.StateConfiguration
}

func (*FinishConfiguration) Direction() codec.Direction {
	return codec.Clientbound
}

func (FinishConfiguration) Append(dst []byte) []byte {
	return dst
}

func (*FinishConfiguration) Decode(*codec.Reader) error {
	return nil
}

type AcknowledgeConfiguration struct{}

func (*AcknowledgeConfiguration) ID() int32 {
	return 0x02
}

func (*AcknowledgeConfiguration) Name() string {
	return "AcknowledgeConfiguration"
}

func (*AcknowledgeConfiguration) State() codec.State {
	return codec.StateConfiguration
}

func (*AcknowledgeConfiguration) Direction() codec.Direction {
	return codec.Serverbound
}

func (AcknowledgeConfiguration) Append(dst []byte) []byte {
	return dst
}

func (*AcknowledgeConfiguration) Decode(*codec.Reader) error {
	return nil
}

type ConfigKeepAlive struct {
	KeepAliveID codec.Long
}

func (*ConfigKeepAlive) ID() int32 {
	return 0x03
}

func (*ConfigKeepAlive) Name() string {
	return "ConfigKeepAlive"
}

func (*ConfigKeepAlive) State() codec.State {
	return codec.StateConfiguration
}

func (*ConfigKeepAlive) Direction() codec.Direction {
	return codec.Clientbound
}

func (p ConfigKeepAlive) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.KeepAliveID)
}

func (p *ConfigKeepAlive) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.KeepAliveID)
}

type ConfigKeepAliveResponse struct {
	KeepAliveID codec.Long
}

func (*ConfigKeepAliveResponse) ID() int32 {
	return 0x03
}

func (*ConfigKeepAliveResponse) Name() string {
	return "ConfigKeepAliveResponse"
}

func (*ConfigKeepAliveResponse) State() codec.State {
	return codec.StateConfiguration
}

func (*ConfigKeepAliveResponse) Direction() codec.Direction {
	return codec.Serverbound
}

func (p ConfigKeepAliveResponse) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.KeepAliveID)
}

func (p *ConfigKeepAliveResponse) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.KeepAliveID)
}

type ConfigPing struct {
	PingID codec.Int
}

func (*ConfigPing) ID() int32 {
	return 0x04
}

func (*ConfigPing) Name() string {
	return "ConfigPing"
}

func (*ConfigPing) State() codec.State {
	return codec.StateConfiguration
}

func (*ConfigPing) Direction() codec.Direction {
	return codec.Clientbound
}

func (p ConfigPing) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.PingID)
}

func (p *ConfigPing) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.PingID)
}

type ConfigPong struct {
	PingID codec.Int
}

func (*ConfigPong) ID() int32 {
	return 0x04
}

func (*ConfigPong) Name() string {
	return "ConfigPong"
}

func (*ConfigPong) State() codec.State {
	return codec.StateConfiguration
}

func (*ConfigPong) Direction() codec.Direction {
	return codec.Serverbound
}

func (p ConfigPong) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.PingID)
}

func (p *ConfigPong) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.PingID)
}

type RegistryData struct {
	Codec codec.NBT
}

func (*RegistryData) ID() int32 {
	return 0x05
}

func (*RegistryData) Name() string {
	return "RegistryData"
}

func (*RegistryData) State() codec.State {
	return codec.StateConfiguration
}

func (*RegistryData) Direction() codec.Direction {
	return codec.Clientbound
}

func (p RegistryData) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.Codec)
}

func (p *RegistryData) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.Codec)
}

type FeatureFlags struct {
	Features codec.Slice[graft.Identifier]
}

func (*FeatureFlags) ID() int32 {
	return 0x07
}

func (*FeatureFlags) Name() string {
	return "FeatureFlags"
}

func (*FeatureFlags) State() codec.State {
	return codec.StateConfiguration
}

func (*FeatureFlags) Direction() codec.Direction {
	return codec.Clientbound
}

func (p FeatureFlags) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.Features)
}

func (p *FeatureFlags) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.Features)
}
