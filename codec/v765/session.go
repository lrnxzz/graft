package v765

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"

	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/codec"
	"github.com/lrnxzz/graft/codec/v765/blocks"
	"github.com/lrnxzz/graft/codec/v765/items"
	"github.com/lrnxzz/graft/lib"
	"github.com/lrnxzz/graft/mojang"
	"github.com/lrnxzz/graft/nbt"
)

type dimensionBounds struct {
	minY   int
	height int
}

var overworld = dimensionBounds{
	minY:   -64,
	height: 384,
}

type JoinHandler func(*codec.Client, *JoinGame) error

type ChatListener func(line string)

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

type blockPrediction struct {
	position graft.Position
	server   graft.BlockState
}

func Join(client *codec.Client, host string, port uint16, username string, onReady JoinHandler) (*Session, error) {
	offline := mojang.Offline{Username: username}

	profile, err := offline.Authenticate(context.Background())
	if err != nil {
		return nil, err
	}

	var uuid codec.UUID
	if raw, err := hex.DecodeString(profile.Profile.ID); err == nil {
		copy(uuid[:], raw)
	}

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
	codec.On(s.client, s.onChunkBatch)
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

func (s *Session) onBlockAck(c *codec.Client, p *AcknowledgeBlockChange) error {
	for _, prediction := range s.pending.Ack(p.Sequence.Int32()) {
		s.world.SetBlock(prediction.position.X, prediction.position.Y, prediction.position.Z, prediction.server)
	}

	return nil
}

func (s *Session) StartDigging(hit graft.RayHit) error {
	return s.action(digStart, hit)
}

func (s *Session) CancelDigging(hit graft.RayHit) error {
	return s.action(digCancel, hit)
}

func (s *Session) FinishDigging(hit graft.RayHit) error {
	if err := s.action(digFinish, hit); err != nil {
		return err
	}

	s.world.SetBlock(hit.Block.X, hit.Block.Y, hit.Block.Z, 0)

	return nil
}

func (s *Session) action(status codec.VarInt, hit graft.RayHit) error {
	packet := &PlayerAction{
		Status:   status,
		Location: hit.Block,
		Face:     codec.Byte(hit.Face),
		Sequence: codec.VarInt(s.pending.Push(s.predict(hit.Block))),
	}

	return s.client.Send(packet)
}

func (s *Session) PlaceBlock(hit graft.RayHit) error {
	placed := hit.Block.Neighbor(hit.Face)
	cursor := hit.Point.Sub(hit.Block.Corner())

	use := &UseItemOn{
		Hand:     mainHand,
		Location: hit.Block,
		Face:     codec.VarInt(hit.Face),
		CursorX:  codec.Float(cursor.X),
		CursorY:  codec.Float(cursor.Y),
		CursorZ:  codec.Float(cursor.Z),
		Sequence: codec.VarInt(s.pending.Push(s.predict(placed))),
	}
	if err := s.client.Send(use); err != nil {
		return err
	}

	held, ok := items.Of(s.inventory.Held().Item)
	if !ok {
		return nil
	}

	block, ok := blocks.Named(held.Name)
	if !ok {
		return nil
	}

	s.world.SetBlock(placed.X, placed.Y, placed.Z, block.DefaultState)

	return nil
}

func (s *Session) predict(position graft.Position) blockPrediction {
	server, _ := s.world.BlockAt(position)

	return blockPrediction{
		position: position,
		server:   server,
	}
}

func (s *Session) applyServerBlock(change BlockChange) {
	position := graft.Position{
		X: change.X,
		Y: change.Y,
		Z: change.Z,
	}

	predicted := false
	s.pending.Each(func(prediction *blockPrediction) {
		if prediction.position == position {
			prediction.server = change.State
			predicted = true
		}
	})
	if predicted {
		return
	}

	s.world.SetBlock(change.X, change.Y, change.Z, change.State)
}

func (s *Session) onRegistryData(c *codec.Client, p *RegistryData) error {
	registry, ok := nbt.Get[nbt.Compound](nbt.Compound(p.Codec), "minecraft:dimension_type")
	if !ok {
		return nil
	}

	entries, ok := nbt.Get[nbt.List](registry, "value")
	if !ok {
		return nil
	}

	types, ok := nbt.Items[nbt.Compound](entries)
	if !ok {
		return nil
	}

	for _, entry := range types {
		name, ok := nbt.Get[nbt.String](entry, "name")
		if !ok {
			continue
		}

		element, ok := nbt.Get[nbt.Compound](entry, "element")
		if !ok {
			continue
		}

		minY, ok := nbt.Get[nbt.Int](element, "min_y")
		if !ok {
			continue
		}

		height, ok := nbt.Get[nbt.Int](element, "height")
		if !ok {
			continue
		}

		s.dimensions[graft.Identifier(name)] = dimensionBounds{
			minY:   int(minY),
			height: int(height),
		}
	}

	return nil
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

// the server paces chunk delivery from this number; a bot has no frame budget
// to protect, so it asks for as much as the vanilla client ever would
const chunkBatchRate codec.Float = 64

func (s *Session) onChunkBatch(c *codec.Client, p *ChunkBatchFinished) error {
	return c.Send(&ChunkBatchReceived{ChunksPerTick: chunkBatchRate})
}

func (s *Session) onChunkData(c *codec.Client, p *ChunkData) error {
	column, err := p.Column(s.bounds.minY, s.bounds.height)
	if err != nil {
		return err
	}

	s.world.LoadColumn(column)

	return nil
}

func (s *Session) onUnloadChunk(c *codec.Client, p *UnloadChunk) error {
	s.world.UnloadColumn(graft.Chunk(p.X.Int32(), p.Z.Int32()))

	return nil
}

func (s *Session) onBlockUpdate(c *codec.Client, p *BlockUpdate) error {
	s.applyServerBlock(p.Change())

	return nil
}

func (s *Session) onSectionBlocks(c *codec.Client, p *SectionBlocksUpdate) error {
	for _, change := range p.Changes() {
		s.applyServerBlock(change)
	}

	return nil
}

func (s *Session) onHealth(c *codec.Client, p *SetHealth) error {
	p.Apply(s.player)

	return nil
}

func (s *Session) Inventory() *graft.Inventory {
	return &s.inventory
}

func (s *Session) onContainerContent(c *codec.Client, p *SetContainerContent) error {
	if p.WindowID != 0 {
		return nil
	}

	s.stateID = p.StateID.Int32()

	stacks := make([]graft.ItemStack, len(p.Slots))
	for index, slot := range p.Slots {
		stacks[index] = slot.Stack()
	}
	s.inventory.Load(stacks)
	s.carried = p.Carried.Stack()

	return nil
}

func (s *Session) onContainerSlot(c *codec.Client, p *SetContainerSlot) error {
	s.stateID = p.StateID.Int32()

	switch {
	case p.WindowID == -1 && p.Index == -1:
		s.carried = p.Data.Stack()
	case p.WindowID == 0:
		s.inventory.SetSlot(p.Index.Int(), p.Data.Stack())
	}

	return nil
}

func (s *Session) onHeldItem(c *codec.Client, p *SetHeldItem) error {
	s.inventory.SelectHeld(p.Slot.Int())

	return nil
}

func (s *Session) SelectHotbar(index int) error {
	if index < 0 || index >= graft.HotbarSize {
		return fmt.Errorf("v765: hotbar index %d is out of range", index)
	}

	if err := s.client.Send(&HeldItemChange{Slot: codec.Short(index)}); err != nil {
		return err
	}

	s.inventory.SelectHeld(index)

	return nil
}

func (s *Session) SwapWithHotbar(slot, hotbar int) error {
	if hotbar < 0 || hotbar >= graft.HotbarSize {
		return fmt.Errorf("v765: hotbar index %d is out of range", hotbar)
	}

	return s.swap(slot, codec.Byte(hotbar), graft.HotbarSlot(hotbar))
}

func (s *Session) SwapWithOffhand(slot int) error {
	return s.swap(slot, offhandButton, graft.SlotOffhand)
}

func (s *Session) swap(slot int, button codec.Byte, other int) error {
	if slot < 0 || slot >= graft.InventorySize {
		return fmt.Errorf("v765: inventory slot %d is out of range", slot)
	}
	if slot == other {
		return nil
	}

	first := s.inventory.Slot(slot)
	second := s.inventory.Slot(other)

	click := &ClickContainer{
		StateID: codec.VarInt(s.stateID),
		Index:   codec.Short(slot),
		Button:  button,
		Mode:    clickSwap,
		Changed: codec.Slice[ChangedSlot]{
			{
				Index: codec.Short(slot),
				Item:  slotOf(second),
			},
			{
				Index: codec.Short(other),
				Item:  slotOf(first),
			},
		},
		Carried: slotOf(s.carried),
	}

	if err := s.client.Send(click); err != nil {
		return err
	}

	s.inventory.Swap(slot, other)

	return nil
}

func (s *Session) Carried() graft.ItemStack {
	return s.carried
}

func (s *Session) ClickSlot(slot int) error {
	if slot < 0 || slot >= graft.InventorySize {
		return fmt.Errorf("v765: inventory slot %d is out of range", slot)
	}
	if slot == graft.SlotCraftingOutput && !s.carried.Empty() {
		return nil
	}

	current := s.inventory.Slot(slot)
	if current.Empty() && s.carried.Empty() {
		return nil
	}

	landed, carried := land(current, s.carried)

	changed := ChangedSlot{
		Index: codec.Short(slot),
		Item:  slotOf(landed),
	}
	click := &ClickContainer{
		StateID: codec.VarInt(s.stateID),
		Index:   codec.Short(slot),
		Mode:    clickPickup,
		Changed: codec.Slice[ChangedSlot]{changed},
		Carried: slotOf(carried),
	}
	if err := s.client.Send(click); err != nil {
		return err
	}

	s.inventory.SetSlot(slot, landed)
	s.carried = carried

	return nil
}

func land(slot, carried graft.ItemStack) (graft.ItemStack, graft.ItemStack) {
	if carried.Empty() || slot.Empty() || slot.Item != carried.Item {
		return carried, slot
	}

	size := graft.MaxStackSize
	item, known := items.Of(slot.Item)
	if known {
		size = item.StackSize
	}

	total := slot.Count + carried.Count
	slot.Count = min(total, size)
	carried.Count = total - slot.Count
	if carried.Count == 0 {
		carried = graft.ItemStack{}
	}

	return slot, carried
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
