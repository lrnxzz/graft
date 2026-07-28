package v765

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/nbt"
)

const filterPartial gocraft.VarInt = 2

type chatStamp struct {
	timestamp gocraft.Long
	salt      gocraft.Long
}

func stampChat() chatStamp {
	return chatStamp{
		timestamp: gocraft.Long(time.Now().UnixMilli()),
		salt:      gocraft.Long(rand.Int64()),
	}
}

type chatSignature [256]byte

func (s chatSignature) Append(dst []byte) []byte {
	return append(dst, s[:]...)
}

func (s *chatSignature) Decode(r *gocraft.Reader) error {
	for index := range s {
		raw, err := r.ReadByte()
		if err != nil {
			return err
		}
		s[index] = raw
	}

	return nil
}

type chatAcknowledged [3]gocraft.UByte

func (a chatAcknowledged) Append(dst []byte) []byte {
	for _, part := range a {
		dst = part.Append(dst)
	}

	return dst
}

func (a *chatAcknowledged) Decode(r *gocraft.Reader) error {
	for index := range a {
		if err := a[index].Decode(r); err != nil {
			return err
		}
	}

	return nil
}

type ChatMessage struct {
	Message      gocraft.String
	Timestamp    gocraft.Long
	Salt         gocraft.Long
	Signature    gocraft.Option[chatSignature]
	Count        gocraft.VarInt
	Acknowledged chatAcknowledged
}

func (*ChatMessage) ID() int32 {
	return 0x05
}

func (*ChatMessage) Name() string {
	return "ChatMessage"
}

func (*ChatMessage) State() gocraft.State {
	return gocraft.StatePlay
}

func (*ChatMessage) Direction() gocraft.Direction {
	return gocraft.Serverbound
}

func (p ChatMessage) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.Message, p.Timestamp, p.Salt, p.Signature, p.Count, p.Acknowledged)
}

func (p *ChatMessage) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.Message, &p.Timestamp, &p.Salt, &p.Signature, &p.Count, &p.Acknowledged)
}

type argumentSignature struct {
	Argument  gocraft.String
	Signature chatSignature
}

func (s argumentSignature) Append(dst []byte) []byte {
	dst = s.Argument.Append(dst)

	return s.Signature.Append(dst)
}

func (s *argumentSignature) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &s.Argument, &s.Signature)
}

type ChatCommand struct {
	Command      gocraft.String
	Timestamp    gocraft.Long
	Salt         gocraft.Long
	Signatures   gocraft.Slice[argumentSignature]
	Count        gocraft.VarInt
	Acknowledged chatAcknowledged
}

func (*ChatCommand) ID() int32 {
	return 0x04
}

func (*ChatCommand) Name() string {
	return "ChatCommand"
}

func (*ChatCommand) State() gocraft.State {
	return gocraft.StatePlay
}

func (*ChatCommand) Direction() gocraft.Direction {
	return gocraft.Serverbound
}

func (p ChatCommand) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.Command, p.Timestamp, p.Salt, p.Signatures, p.Count, p.Acknowledged)
}

func (p *ChatCommand) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.Command, &p.Timestamp, &p.Salt, &p.Signatures, &p.Count, &p.Acknowledged)
}

type SystemChat struct {
	Content gocraft.Text
	Overlay gocraft.Bool
}

func (*SystemChat) ID() int32 {
	return 0x69
}

func (*SystemChat) Name() string {
	return "SystemChat"
}

func (*SystemChat) State() gocraft.State {
	return gocraft.StatePlay
}

func (*SystemChat) Direction() gocraft.Direction {
	return gocraft.Clientbound
}

func (p SystemChat) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.Content, p.Overlay)
}

func (p *SystemChat) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.Content, &p.Overlay)
}

type previousMessage struct {
	ID        gocraft.VarInt
	Signature chatSignature
}

func (m previousMessage) Append(dst []byte) []byte {
	dst = m.ID.Append(dst)
	if m.ID == 0 {
		dst = m.Signature.Append(dst)
	}

	return dst
}

func (m *previousMessage) Decode(r *gocraft.Reader) error {
	if err := m.ID.Decode(r); err != nil {
		return err
	}
	if m.ID != 0 {
		return nil
	}

	return m.Signature.Decode(r)
}

type PlayerChat struct {
	Sender      gocraft.UUID
	Index       gocraft.VarInt
	Signature   gocraft.Option[chatSignature]
	Message     gocraft.String
	Timestamp   gocraft.Long
	Salt        gocraft.Long
	Previous    gocraft.Slice[previousMessage]
	Unsigned    gocraft.Option[gocraft.Text]
	Filter      gocraft.VarInt
	FilterBits  gocraft.Slice[gocraft.Long]
	ChatType    gocraft.VarInt
	NetworkName gocraft.Text
	Target      gocraft.Option[gocraft.Text]
}

func (*PlayerChat) ID() int32 {
	return 0x37
}

func (*PlayerChat) Name() string {
	return "PlayerChat"
}

func (*PlayerChat) State() gocraft.State {
	return gocraft.StatePlay
}

func (*PlayerChat) Direction() gocraft.Direction {
	return gocraft.Clientbound
}

func (p PlayerChat) Append(dst []byte) []byte {
	dst = gocraft.AppendAll(dst, p.Sender, p.Index, p.Signature, p.Message, p.Timestamp, p.Salt, p.Previous, p.Unsigned, p.Filter)
	if p.Filter == filterPartial {
		dst = p.FilterBits.Append(dst)
	}

	return gocraft.AppendAll(dst, p.ChatType, p.NetworkName, p.Target)
}

func (p *PlayerChat) Decode(r *gocraft.Reader) error {
	if err := gocraft.DecodeAll(r, &p.Sender, &p.Index, &p.Signature, &p.Message, &p.Timestamp, &p.Salt, &p.Previous, &p.Unsigned, &p.Filter); err != nil {
		return err
	}
	if p.Filter == filterPartial {
		if err := p.FilterBits.Decode(r); err != nil {
			return err
		}
	}

	return gocraft.DecodeAll(r, &p.ChatType, &p.NetworkName, &p.Target)
}

type DisguisedChat struct {
	Message    gocraft.Text
	ChatType   gocraft.VarInt
	SenderName gocraft.Text
	Target     gocraft.Option[gocraft.Text]
}

func (*DisguisedChat) ID() int32 {
	return 0x1E
}

func (*DisguisedChat) Name() string {
	return "DisguisedChat"
}

func (*DisguisedChat) State() gocraft.State {
	return gocraft.StatePlay
}

func (*DisguisedChat) Direction() gocraft.Direction {
	return gocraft.Clientbound
}

func (p DisguisedChat) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.Message, p.ChatType, p.SenderName, p.Target)
}

func (p *DisguisedChat) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.Message, &p.ChatType, &p.SenderName, &p.Target)
}

// the session renders player and disguised chat with these same shapes, so they
// are named here rather than spelled out again at each call site
const (
	formatChat         = "<%s> %s"
	formatAnnouncement = "[%s] %s"
	formatEmote        = "* %s %s"
	formatJoined       = "%s joined the game"
	formatLeft         = "%s left the game"
)

var chatFormats = map[string]string{
	"chat.type.text":            formatChat,
	"chat.type.announcement":    formatAnnouncement,
	"chat.type.emote":           formatEmote,
	"multiplayer.player.joined": formatJoined,
	"multiplayer.player.left":   formatLeft,
}

func plainText(tag nbt.Tag) string {
	switch component := tag.(type) {
	case nbt.String:
		return string(component)
	case nbt.Compound:
		translate, translatable := nbt.Get[nbt.String](component, "translate")
		if translatable {
			return translated(component, string(translate))
		}

		var plain strings.Builder

		text, literal := nbt.Get[nbt.String](component, "text")
		if literal {
			plain.WriteString(string(text))
		}

		extra, nested := nbt.Get[nbt.List](component, "extra")
		if nested {
			for _, item := range extra.Items {
				plain.WriteString(plainText(item))
			}
		}

		return plain.String()
	default:
		return ""
	}
}

func translated(component nbt.Compound, key string) string {
	var arguments []string

	with, parameterized := nbt.Get[nbt.List](component, "with")
	if parameterized {
		for _, item := range with.Items {
			arguments = append(arguments, plainText(item))
		}
	}

	format, known := chatFormats[key]
	if !known {
		parts := make([]string, 0, len(arguments)+1)
		parts = append(parts, key)
		parts = append(parts, arguments...)

		return strings.Join(parts, " ")
	}

	values := make([]any, len(arguments))
	for index, argument := range arguments {
		values[index] = argument
	}

	return fmt.Sprintf(format, values...)
}
