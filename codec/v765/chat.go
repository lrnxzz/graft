package v765

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/lrnxzz/graft/codec"
	"github.com/lrnxzz/graft/nbt"
)

const filterPartial codec.VarInt = 2

type chatStamp struct {
	timestamp codec.Long
	salt      codec.Long
}

func stampChat() chatStamp {
	return chatStamp{
		timestamp: codec.Long(time.Now().UnixMilli()),
		salt:      codec.Long(rand.Int64()),
	}
}

type chatSignature [256]byte

func (s chatSignature) Append(dst []byte) []byte {
	return append(dst, s[:]...)
}

func (s *chatSignature) Decode(r *codec.Reader) error {
	for index := range s {
		raw, err := r.ReadByte()
		if err != nil {
			return err
		}
		s[index] = raw
	}

	return nil
}

type chatAcknowledged [3]codec.UByte

func (a chatAcknowledged) Append(dst []byte) []byte {
	for _, part := range a {
		dst = part.Append(dst)
	}

	return dst
}

func (a *chatAcknowledged) Decode(r *codec.Reader) error {
	for index := range a {
		if err := a[index].Decode(r); err != nil {
			return err
		}
	}

	return nil
}

type ChatMessage struct {
	Message      codec.String
	Timestamp    codec.Long
	Salt         codec.Long
	Signature    codec.Option[chatSignature]
	Count        codec.VarInt
	Acknowledged chatAcknowledged
}

func (*ChatMessage) ID() int32 {
	return 0x05
}

func (*ChatMessage) Name() string {
	return "ChatMessage"
}

func (*ChatMessage) State() codec.State {
	return codec.StatePlay
}

func (*ChatMessage) Direction() codec.Direction {
	return codec.Serverbound
}

func (p ChatMessage) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.Message, p.Timestamp, p.Salt, p.Signature, p.Count, p.Acknowledged)
}

func (p *ChatMessage) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.Message, &p.Timestamp, &p.Salt, &p.Signature, &p.Count, &p.Acknowledged)
}

type argumentSignature struct {
	Argument  codec.String
	Signature chatSignature
}

func (s argumentSignature) Append(dst []byte) []byte {
	dst = s.Argument.Append(dst)

	return s.Signature.Append(dst)
}

func (s *argumentSignature) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &s.Argument, &s.Signature)
}

type ChatCommand struct {
	Command      codec.String
	Timestamp    codec.Long
	Salt         codec.Long
	Signatures   codec.Slice[argumentSignature]
	Count        codec.VarInt
	Acknowledged chatAcknowledged
}

func (*ChatCommand) ID() int32 {
	return 0x04
}

func (*ChatCommand) Name() string {
	return "ChatCommand"
}

func (*ChatCommand) State() codec.State {
	return codec.StatePlay
}

func (*ChatCommand) Direction() codec.Direction {
	return codec.Serverbound
}

func (p ChatCommand) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.Command, p.Timestamp, p.Salt, p.Signatures, p.Count, p.Acknowledged)
}

func (p *ChatCommand) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.Command, &p.Timestamp, &p.Salt, &p.Signatures, &p.Count, &p.Acknowledged)
}

type SystemChat struct {
	Content codec.Text
	Overlay codec.Bool
}

func (*SystemChat) ID() int32 {
	return 0x69
}

func (*SystemChat) Name() string {
	return "SystemChat"
}

func (*SystemChat) State() codec.State {
	return codec.StatePlay
}

func (*SystemChat) Direction() codec.Direction {
	return codec.Clientbound
}

func (p SystemChat) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.Content, p.Overlay)
}

func (p *SystemChat) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.Content, &p.Overlay)
}

type previousMessage struct {
	ID        codec.VarInt
	Signature chatSignature
}

func (m previousMessage) Append(dst []byte) []byte {
	dst = m.ID.Append(dst)
	if m.ID == 0 {
		dst = m.Signature.Append(dst)
	}

	return dst
}

func (m *previousMessage) Decode(r *codec.Reader) error {
	if err := m.ID.Decode(r); err != nil {
		return err
	}
	if m.ID != 0 {
		return nil
	}

	return m.Signature.Decode(r)
}

type PlayerChat struct {
	Sender      codec.UUID
	Index       codec.VarInt
	Signature   codec.Option[chatSignature]
	Message     codec.String
	Timestamp   codec.Long
	Salt        codec.Long
	Previous    codec.Slice[previousMessage]
	Unsigned    codec.Option[codec.Text]
	Filter      codec.VarInt
	FilterBits  codec.Slice[codec.Long]
	ChatType    codec.VarInt
	NetworkName codec.Text
	Target      codec.Option[codec.Text]
}

func (*PlayerChat) ID() int32 {
	return 0x37
}

func (*PlayerChat) Name() string {
	return "PlayerChat"
}

func (*PlayerChat) State() codec.State {
	return codec.StatePlay
}

func (*PlayerChat) Direction() codec.Direction {
	return codec.Clientbound
}

func (p PlayerChat) Append(dst []byte) []byte {
	dst = codec.AppendAll(dst, p.Sender, p.Index, p.Signature, p.Message, p.Timestamp, p.Salt, p.Previous, p.Unsigned, p.Filter)
	if p.Filter == filterPartial {
		dst = p.FilterBits.Append(dst)
	}

	return codec.AppendAll(dst, p.ChatType, p.NetworkName, p.Target)
}

func (p *PlayerChat) Decode(r *codec.Reader) error {
	if err := codec.DecodeAll(r, &p.Sender, &p.Index, &p.Signature, &p.Message, &p.Timestamp, &p.Salt, &p.Previous, &p.Unsigned, &p.Filter); err != nil {
		return err
	}
	if p.Filter == filterPartial {
		if err := p.FilterBits.Decode(r); err != nil {
			return err
		}
	}

	return codec.DecodeAll(r, &p.ChatType, &p.NetworkName, &p.Target)
}

type DisguisedChat struct {
	Message    codec.Text
	ChatType   codec.VarInt
	SenderName codec.Text
	Target     codec.Option[codec.Text]
}

func (*DisguisedChat) ID() int32 {
	return 0x1E
}

func (*DisguisedChat) Name() string {
	return "DisguisedChat"
}

func (*DisguisedChat) State() codec.State {
	return codec.StatePlay
}

func (*DisguisedChat) Direction() codec.Direction {
	return codec.Clientbound
}

func (p DisguisedChat) Append(dst []byte) []byte {
	return codec.AppendAll(dst, p.Message, p.ChatType, p.SenderName, p.Target)
}

func (p *DisguisedChat) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &p.Message, &p.ChatType, &p.SenderName, &p.Target)
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

type ChatListener func(line string)

func (s *Session) OnChat(listener ChatListener) {
	s.chat = listener
}

func (s *Session) notifyChat(line string) {
	if s.chat != nil {
		s.chat(line)
	}
}

func (s *Session) onSystemChat(c *codec.Client, p *SystemChat) error {
	if !p.Overlay.Bool() {
		s.notifyChat(plainText(p.Content.Tag))
	}

	return nil
}

func (s *Session) onPlayerChat(c *codec.Client, p *PlayerChat) error {
	s.notifyChat(fmt.Sprintf(formatChat, plainText(p.NetworkName.Tag), p.Message))

	return nil
}

func (s *Session) onDisguisedChat(c *codec.Client, p *DisguisedChat) error {
	s.notifyChat(fmt.Sprintf(formatAnnouncement, plainText(p.SenderName.Tag), plainText(p.Message.Tag)))

	return nil
}

func (s *Session) SendChat(message string) error {
	stamp := stampChat()

	return s.client.Send(&ChatMessage{
		Message:   codec.String(message),
		Timestamp: stamp.timestamp,
		Salt:      stamp.salt,
	})
}

func (s *Session) SendCommand(command string) error {
	stamp := stampChat()

	return s.client.Send(&ChatCommand{
		Command:   codec.String(command),
		Timestamp: stamp.timestamp,
		Salt:      stamp.salt,
	})
}
