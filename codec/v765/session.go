package v765

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"

	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/codec"
	"github.com/lrnxzz/graft/lib"
	"github.com/lrnxzz/graft/mojang"
)

type JoinHandler func(*codec.Client, *JoinGame) error

type Session struct {
	client     *codec.Client
	world      *graft.World
	player     *graft.Player
	ready      JoinHandler
	spawned    bool
	dimensions map[graft.Identifier]dimensionBounds
	bounds     dimensionBounds
	view       int
	inventory  graft.Inventory
	carried    graft.ItemStack
	stateID    int32
	pending    lib.Pending[blockPrediction]
	chat       ChatListener
}

func Join(client *codec.Client, host string, port uint16, username string, onReady JoinHandler) (*Session, error) {
	offline := mojang.Offline{Username: username}

	profile, err := offline.Authenticate(context.Background())
	if err != nil {
		return nil, err
	}

	// a profile id that does not decode is a bug in whoever produced the
	// profile, not a condition to play on with a zeroed uuid
	raw, err := hex.DecodeString(profile.Profile.ID)
	if err != nil {
		return nil, fmt.Errorf("v765: profile id %q is not hex: %w", profile.Profile.ID, err)
	}

	var uuid codec.UUID
	copy(uuid[:], raw)

	session := &Session{
		client:     client,
		world:      graft.NewWorld(),
		player:     &graft.Player{},
		ready:      onReady,
		dimensions: map[graft.Identifier]dimensionBounds{},
		bounds:     overworld,
	}
	session.listen()

	if err := client.Send(&Handshake{
		ProtocolVersion: ProtocolVersion,
		ServerAddress:   codec.String(host),
		ServerPort:      codec.UShort(port),
		NextState:       codec.VarInt(codec.StateLogin),
	}); err != nil {
		return nil, err
	}

	client.SetState(codec.StateLogin)

	return session, client.Send(&LoginStart{
		Username: codec.String(username),
		UUID:     uuid,
	})
}

func (s *Session) World() *graft.World {
	return s.world
}

func (s *Session) ViewDistance() int {
	return s.view
}

func (s *Session) Player() *graft.Player {
	return s.player
}

func (s *Session) Spawned() bool {
	return s.spawned
}

func (s *Session) listen() {
	codec.On(s.client, s.onCompression)
	codec.On(s.client, s.onEncryption)
	codec.On(s.client, s.onLoginSuccess)
	codec.On(s.client, s.onLoginDisconnect)

	codec.On(s.client, s.onConfigKeepAlive)
	codec.On(s.client, s.onConfigPing)
	codec.On(s.client, s.onRegistryData)
	codec.On(s.client, s.onFinishConfiguration)
	codec.On(s.client, s.onConfigDisconnect)

	codec.On(s.client, s.onJoinGame)
	codec.On(s.client, s.onChunkBatchFinished)
	codec.On(s.client, s.onKeepAlive)
	codec.On(s.client, s.onSyncPosition)
	codec.On(s.client, s.onChunkData)
	codec.On(s.client, s.onUnloadChunk)
	codec.On(s.client, s.onBlockUpdate)
	codec.On(s.client, s.onSectionBlocks)
	codec.On(s.client, s.onHealth)
	codec.On(s.client, s.onAbilities)
	codec.On(s.client, s.onExperience)
	codec.On(s.client, s.onContainerContent)
	codec.On(s.client, s.onContainerSlot)
	codec.On(s.client, s.onHeldItem)
	codec.On(s.client, s.onSystemChat)
	codec.On(s.client, s.onPlayerChat)
	codec.On(s.client, s.onDisguisedChat)
	codec.On(s.client, s.onBlockAck)
	codec.On(s.client, s.onPlayDisconnect)
}

func (s *Session) onCompression(c *codec.Client, p *SetCompression) error {
	c.SetCompression(p.Threshold.Int())

	return nil
}

func (s *Session) onEncryption(c *codec.Client, p *EncryptionBegin) error {
	return errors.New("v765: server requested encryption (online-mode); auth and encryption are not implemented")
}

func (s *Session) onLoginSuccess(c *codec.Client, p *LoginSuccess) error {
	p.Apply(s.player)

	if err := c.Send(&LoginAcknowledged{}); err != nil {
		return err
	}

	c.SetState(codec.StateConfiguration)

	return c.Send(&ClientInformation{
		Locale:              "en_us",
		ViewDistance:        8,
		MainHand:            1,
		EnableServerListing: true,
	})
}

func (s *Session) onLoginDisconnect(c *codec.Client, p *LoginDisconnect) error {
	return fmt.Errorf("v765: kicked during login: %s", p.Reason)
}

func (s *Session) onConfigKeepAlive(c *codec.Client, p *ConfigKeepAlive) error {
	return c.Send(&ConfigKeepAliveResponse{KeepAliveID: p.KeepAliveID})
}

func (s *Session) onConfigPing(c *codec.Client, p *ConfigPing) error {
	return c.Send(&ConfigPong{PingID: p.PingID})
}

func (s *Session) onFinishConfiguration(c *codec.Client, p *FinishConfiguration) error {
	if err := c.Send(&AcknowledgeConfiguration{}); err != nil {
		return err
	}

	c.SetState(codec.StatePlay)

	return nil
}

func (s *Session) onConfigDisconnect(c *codec.Client, p *ConfigDisconnect) error {
	return errors.New("v765: kicked during configuration")
}

func (s *Session) onJoinGame(c *codec.Client, p *JoinGame) error {
	p.Apply(s.player)

	// falling back keeps the session alive, but the wrong minY and height would
	// silently corrupt every chunk decoded from here on, so it is worth saying
	bounds, known := s.dimensions[p.DimensionType]
	if !known {
		slog.Warn("unknown dimension, assuming overworld bounds", "dimension", p.DimensionType)

		bounds = overworld
	}
	s.bounds = bounds
	s.view = p.ViewDistance.Int()

	if s.ready != nil {
		return s.ready(c, p)
	}

	return nil
}

func (s *Session) onKeepAlive(c *codec.Client, p *PlayKeepAlive) error {
	return c.Send(&PlayKeepAliveResponse{KeepAliveID: p.KeepAliveID})
}

func (s *Session) onSyncPosition(c *codec.Client, p *SyncPlayerPosition) error {
	p.Apply(s.player)
	s.spawned = true

	if err := c.Send(&ConfirmTeleport{TeleportID: p.TeleportID}); err != nil {
		return err
	}

	return s.SendPosition()
}

func (s *Session) SendPosition() error {
	return s.client.Send(&SetPlayerPositionRotation{
		X:        codec.Double(s.player.Position.X),
		Y:        codec.Double(s.player.Position.Y),
		Z:        codec.Double(s.player.Position.Z),
		Yaw:      codec.Float(s.player.Yaw),
		Pitch:    codec.Float(s.player.Pitch),
		OnGround: codec.Bool(s.player.OnGround),
	})
}

func (s *Session) onHealth(c *codec.Client, p *SetHealth) error {
	p.Apply(s.player)

	return nil
}

func (s *Session) onAbilities(c *codec.Client, p *PlayerAbilities) error {
	p.Apply(s.player)

	return nil
}

func (s *Session) onExperience(c *codec.Client, p *SetExperience) error {
	p.Apply(s.player)

	return nil
}

func (s *Session) onPlayDisconnect(c *codec.Client, p *PlayDisconnect) error {
	return errors.New("v765: kicked during play")
}
