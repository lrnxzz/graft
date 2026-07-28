package v765

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/codec/v765/blocks"
	"github.com/lrnxzz/go-craft/codec/v765/items"
	"github.com/lrnxzz/go-craft/lib"
	"github.com/lrnxzz/go-craft/mojang"
	"github.com/lrnxzz/go-craft/nbt"
)

type dimensionBounds struct {
	minY   int
	height int
}

var overworld = dimensionBounds{
	minY:   -64,
	height: 384,
}

type JoinHandler func(*gocraft.Client, *JoinGame) error

type ChatListener func(line string)

type Session struct {
	client     *gocraft.Client
	world      *gocraft.World
	player     *gocraft.Player
	ready      JoinHandler
	spawned    bool
	dimensions map[gocraft.Identifier]dimensionBounds
	bounds     dimensionBounds
	view       int
	inventory  gocraft.Inventory
	carried    gocraft.ItemStack
	stateID    int32
	pending    lib.Pending[blockPrediction]
	chat       ChatListener
}

type blockPrediction struct {
	position gocraft.Position
	server   gocraft.BlockState
}

func Join(client *gocraft.Client, host string, port uint16, username string, onReady JoinHandler) (*Session, error) {
	offline := mojang.Offline{Username: username}

	profile, err := offline.Authenticate(context.Background())
	if err != nil {
		return nil, err
	}

	var uuid gocraft.UUID
	if raw, err := hex.DecodeString(profile.Profile.ID); err == nil {
		copy(uuid[:], raw)
	}

	session := &Session{
		client:     client,
		world:      gocraft.NewWorld(),
		player:     &gocraft.Player{},
		ready:      onReady,
		dimensions: map[gocraft.Identifier]dimensionBounds{},
		bounds:     overworld,
	}
	session.listen()

	if err := client.Send(&Handshake{
		ProtocolVersion: ProtocolVersion,
		ServerAddress:   gocraft.String(host),
		ServerPort:      gocraft.UShort(port),
		NextState:       gocraft.VarInt(gocraft.StateLogin),
	}); err != nil {
		return nil, err
	}

	client.SetState(gocraft.StateLogin)

	return session, client.Send(&LoginStart{
		Username: gocraft.String(username),
		UUID:     uuid,
	})
}

func (s *Session) World() *gocraft.World {
	return s.world
}

func (s *Session) ViewDistance() int {
	return s.view
}

func (s *Session) Player() *gocraft.Player {
	return s.player
}

func (s *Session) Spawned() bool {
	return s.spawned
}

func (s *Session) listen() {
	gocraft.On(s.client, s.onCompression)
	gocraft.On(s.client, s.onEncryption)
	gocraft.On(s.client, s.onLoginSuccess)
	gocraft.On(s.client, s.onLoginDisconnect)

	gocraft.On(s.client, s.onConfigKeepAlive)
	gocraft.On(s.client, s.onConfigPing)
	gocraft.On(s.client, s.onRegistryData)
	gocraft.On(s.client, s.onFinishConfiguration)
	gocraft.On(s.client, s.onConfigDisconnect)

	gocraft.On(s.client, s.onJoinGame)
	gocraft.On(s.client, s.onChunkBatch)
	gocraft.On(s.client, s.onKeepAlive)
	gocraft.On(s.client, s.onSyncPosition)
	gocraft.On(s.client, s.onChunkData)
	gocraft.On(s.client, s.onUnloadChunk)
	gocraft.On(s.client, s.onBlockUpdate)
	gocraft.On(s.client, s.onSectionBlocks)
	gocraft.On(s.client, s.onHealth)
	gocraft.On(s.client, s.onAbilities)
	gocraft.On(s.client, s.onExperience)
	gocraft.On(s.client, s.onContainerContent)
	gocraft.On(s.client, s.onContainerSlot)
	gocraft.On(s.client, s.onHeldItem)
	gocraft.On(s.client, s.onSystemChat)
	gocraft.On(s.client, s.onPlayerChat)
	gocraft.On(s.client, s.onDisguisedChat)
	gocraft.On(s.client, s.onBlockAck)
	gocraft.On(s.client, s.onPlayDisconnect)
}

func (s *Session) OnChat(listener ChatListener) {
	s.chat = listener
}

func (s *Session) notifyChat(line string) {
	if s.chat != nil {
		s.chat(line)
	}
}

func (s *Session) onSystemChat(c *gocraft.Client, p *SystemChat) error {
	if !p.Overlay.Bool() {
		s.notifyChat(plainText(p.Content.Tag))
	}

	return nil
}

func (s *Session) onPlayerChat(c *gocraft.Client, p *PlayerChat) error {
	s.notifyChat(fmt.Sprintf(formatChat, plainText(p.NetworkName.Tag), p.Message))

	return nil
}

func (s *Session) onDisguisedChat(c *gocraft.Client, p *DisguisedChat) error {
	s.notifyChat(fmt.Sprintf(formatAnnouncement, plainText(p.SenderName.Tag), plainText(p.Message.Tag)))

	return nil
}

func (s *Session) SendChat(message string) error {
	stamp := stampChat()

	return s.client.Send(&ChatMessage{
		Message:   gocraft.String(message),
		Timestamp: stamp.timestamp,
		Salt:      stamp.salt,
	})
}

func (s *Session) SendCommand(command string) error {
	stamp := stampChat()

	return s.client.Send(&ChatCommand{
		Command:   gocraft.String(command),
		Timestamp: stamp.timestamp,
		Salt:      stamp.salt,
	})
}

func (s *Session) onBlockAck(c *gocraft.Client, p *AcknowledgeBlockChange) error {
	for _, prediction := range s.pending.Ack(p.Sequence.Int32()) {
		s.world.SetBlock(prediction.position.X, prediction.position.Y, prediction.position.Z, prediction.server)
	}

	return nil
}

func (s *Session) StartDigging(hit gocraft.RayHit) error {
	return s.action(digStart, hit)
}

func (s *Session) CancelDigging(hit gocraft.RayHit) error {
	return s.action(digCancel, hit)
}

func (s *Session) FinishDigging(hit gocraft.RayHit) error {
	if err := s.action(digFinish, hit); err != nil {
		return err
	}

	s.world.SetBlock(hit.Block.X, hit.Block.Y, hit.Block.Z, 0)

	return nil
}

func (s *Session) action(status gocraft.VarInt, hit gocraft.RayHit) error {
	packet := &PlayerAction{
		Status:   status,
		Location: hit.Block,
		Face:     gocraft.Byte(hit.Face),
		Sequence: gocraft.VarInt(s.pending.Push(s.predict(hit.Block))),
	}

	return s.client.Send(packet)
}

func (s *Session) PlaceBlock(hit gocraft.RayHit) error {
	placed := hit.Block.Neighbor(hit.Face)
	cursor := hit.Point.Sub(hit.Block.Corner())

	use := &UseItemOn{
		Hand:     mainHand,
		Location: hit.Block,
		Face:     gocraft.VarInt(hit.Face),
		CursorX:  gocraft.Float(cursor.X),
		CursorY:  gocraft.Float(cursor.Y),
		CursorZ:  gocraft.Float(cursor.Z),
		Sequence: gocraft.VarInt(s.pending.Push(s.predict(placed))),
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

func (s *Session) predict(position gocraft.Position) blockPrediction {
	server, _ := s.world.BlockAt(position)

	return blockPrediction{
		position: position,
		server:   server,
	}
}

func (s *Session) applyServerBlock(change BlockChange) {
	position := gocraft.Position{
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

func (s *Session) onRegistryData(c *gocraft.Client, p *RegistryData) error {
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

		s.dimensions[gocraft.Identifier(name)] = dimensionBounds{
			minY:   int(minY),
			height: int(height),
		}
	}

	return nil
}

func (s *Session) onCompression(c *gocraft.Client, p *SetCompression) error {
	c.SetCompression(p.Threshold.Int())

	return nil
}

func (s *Session) onEncryption(c *gocraft.Client, p *EncryptionBegin) error {
	return errors.New("v765: server requested encryption (online-mode); auth and encryption are not implemented")
}

func (s *Session) onLoginSuccess(c *gocraft.Client, p *LoginSuccess) error {
	p.Apply(s.player)

	if err := c.Send(&LoginAcknowledged{}); err != nil {
		return err
	}

	c.SetState(gocraft.StateConfiguration)

	return c.Send(&ClientInformation{
		Locale:              "en_us",
		ViewDistance:        8,
		MainHand:            1,
		EnableServerListing: true,
	})
}

func (s *Session) onLoginDisconnect(c *gocraft.Client, p *LoginDisconnect) error {
	return fmt.Errorf("v765: kicked during login: %s", p.Reason)
}

func (s *Session) onConfigKeepAlive(c *gocraft.Client, p *ConfigKeepAlive) error {
	return c.Send(&ConfigKeepAliveResponse{KeepAliveID: p.KeepAliveID})
}

func (s *Session) onConfigPing(c *gocraft.Client, p *ConfigPing) error {
	return c.Send(&ConfigPong{PingID: p.PingID})
}

func (s *Session) onFinishConfiguration(c *gocraft.Client, p *FinishConfiguration) error {
	if err := c.Send(&AcknowledgeConfiguration{}); err != nil {
		return err
	}

	c.SetState(gocraft.StatePlay)

	return nil
}

func (s *Session) onConfigDisconnect(c *gocraft.Client, p *ConfigDisconnect) error {
	return errors.New("v765: kicked during configuration")
}

func (s *Session) onJoinGame(c *gocraft.Client, p *JoinGame) error {
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

func (s *Session) onKeepAlive(c *gocraft.Client, p *PlayKeepAlive) error {
	return c.Send(&PlayKeepAliveResponse{KeepAliveID: p.KeepAliveID})
}

func (s *Session) onSyncPosition(c *gocraft.Client, p *SyncPlayerPosition) error {
	p.Apply(s.player)
	s.spawned = true

	if err := c.Send(&ConfirmTeleport{TeleportID: p.TeleportID}); err != nil {
		return err
	}

	return s.SendPosition()
}

func (s *Session) SendPosition() error {
	return s.client.Send(&SetPlayerPositionRotation{
		X:        gocraft.Double(s.player.Position.X),
		Y:        gocraft.Double(s.player.Position.Y),
		Z:        gocraft.Double(s.player.Position.Z),
		Yaw:      gocraft.Float(s.player.Yaw),
		Pitch:    gocraft.Float(s.player.Pitch),
		OnGround: gocraft.Bool(s.player.OnGround),
	})
}

// the server paces chunk delivery from this number; a bot has no frame budget
// to protect, so it asks for as much as the vanilla client ever would
const chunkBatchRate gocraft.Float = 64

func (s *Session) onChunkBatch(c *gocraft.Client, p *ChunkBatchFinished) error {
	return c.Send(&ChunkBatchReceived{ChunksPerTick: chunkBatchRate})
}

func (s *Session) onChunkData(c *gocraft.Client, p *ChunkData) error {
	column, err := p.Column(s.bounds.minY, s.bounds.height)
	if err != nil {
		return err
	}

	s.world.LoadColumn(column)

	return nil
}

func (s *Session) onUnloadChunk(c *gocraft.Client, p *UnloadChunk) error {
	s.world.UnloadColumn(gocraft.Chunk(p.X.Int32(), p.Z.Int32()))

	return nil
}

func (s *Session) onBlockUpdate(c *gocraft.Client, p *BlockUpdate) error {
	s.applyServerBlock(p.Change())

	return nil
}

func (s *Session) onSectionBlocks(c *gocraft.Client, p *SectionBlocksUpdate) error {
	for _, change := range p.Changes() {
		s.applyServerBlock(change)
	}

	return nil
}

func (s *Session) onHealth(c *gocraft.Client, p *SetHealth) error {
	p.Apply(s.player)

	return nil
}

func (s *Session) Inventory() *gocraft.Inventory {
	return &s.inventory
}

func (s *Session) onContainerContent(c *gocraft.Client, p *SetContainerContent) error {
	if p.WindowID != 0 {
		return nil
	}

	s.stateID = p.StateID.Int32()

	stacks := make([]gocraft.ItemStack, len(p.Slots))
	for index, slot := range p.Slots {
		stacks[index] = slot.Stack()
	}
	s.inventory.Load(stacks)
	s.carried = p.Carried.Stack()

	return nil
}

func (s *Session) onContainerSlot(c *gocraft.Client, p *SetContainerSlot) error {
	s.stateID = p.StateID.Int32()

	switch {
	case p.WindowID == -1 && p.Index == -1:
		s.carried = p.Data.Stack()
	case p.WindowID == 0:
		s.inventory.SetSlot(p.Index.Int(), p.Data.Stack())
	}

	return nil
}

func (s *Session) onHeldItem(c *gocraft.Client, p *SetHeldItem) error {
	s.inventory.SelectHeld(p.Slot.Int())

	return nil
}

func (s *Session) SelectHotbar(index int) error {
	if index < 0 || index >= gocraft.HotbarSize {
		return fmt.Errorf("v765: hotbar index %d is out of range", index)
	}

	if err := s.client.Send(&HeldItemChange{Slot: gocraft.Short(index)}); err != nil {
		return err
	}

	s.inventory.SelectHeld(index)

	return nil
}

func (s *Session) SwapWithHotbar(slot, hotbar int) error {
	if hotbar < 0 || hotbar >= gocraft.HotbarSize {
		return fmt.Errorf("v765: hotbar index %d is out of range", hotbar)
	}

	return s.swap(slot, gocraft.Byte(hotbar), gocraft.HotbarSlot(hotbar))
}

func (s *Session) SwapWithOffhand(slot int) error {
	return s.swap(slot, offhandButton, gocraft.SlotOffhand)
}

func (s *Session) swap(slot int, button gocraft.Byte, other int) error {
	if slot < 0 || slot >= gocraft.InventorySize {
		return fmt.Errorf("v765: inventory slot %d is out of range", slot)
	}
	if slot == other {
		return nil
	}

	first := s.inventory.Slot(slot)
	second := s.inventory.Slot(other)

	click := &ClickContainer{
		StateID: gocraft.VarInt(s.stateID),
		Index:   gocraft.Short(slot),
		Button:  button,
		Mode:    clickSwap,
		Changed: gocraft.Slice[ChangedSlot]{
			{
				Index: gocraft.Short(slot),
				Item:  slotOf(second),
			},
			{
				Index: gocraft.Short(other),
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

func (s *Session) Carried() gocraft.ItemStack {
	return s.carried
}

func (s *Session) ClickSlot(slot int) error {
	if slot < 0 || slot >= gocraft.InventorySize {
		return fmt.Errorf("v765: inventory slot %d is out of range", slot)
	}
	if slot == gocraft.SlotCraftingOutput && !s.carried.Empty() {
		return nil
	}

	current := s.inventory.Slot(slot)
	if current.Empty() && s.carried.Empty() {
		return nil
	}

	landed, carried := land(current, s.carried)

	changed := ChangedSlot{
		Index: gocraft.Short(slot),
		Item:  slotOf(landed),
	}
	click := &ClickContainer{
		StateID: gocraft.VarInt(s.stateID),
		Index:   gocraft.Short(slot),
		Mode:    clickPickup,
		Changed: gocraft.Slice[ChangedSlot]{changed},
		Carried: slotOf(carried),
	}
	if err := s.client.Send(click); err != nil {
		return err
	}

	s.inventory.SetSlot(slot, landed)
	s.carried = carried

	return nil
}

func land(slot, carried gocraft.ItemStack) (gocraft.ItemStack, gocraft.ItemStack) {
	if carried.Empty() || slot.Empty() || slot.Item != carried.Item {
		return carried, slot
	}

	size := gocraft.MaxStackSize
	item, known := items.Of(slot.Item)
	if known {
		size = item.StackSize
	}

	total := slot.Count + carried.Count
	slot.Count = min(total, size)
	carried.Count = total - slot.Count
	if carried.Count == 0 {
		carried = gocraft.ItemStack{}
	}

	return slot, carried
}

func (s *Session) onAbilities(c *gocraft.Client, p *PlayerAbilities) error {
	p.Apply(s.player)

	return nil
}

func (s *Session) onExperience(c *gocraft.Client, p *SetExperience) error {
	p.Apply(s.player)

	return nil
}

func (s *Session) onPlayDisconnect(c *gocraft.Client, p *PlayDisconnect) error {
	return errors.New("v765: kicked during play")
}
