package agent

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	gocraft "github.com/lrnxzz/go-craft"
	v765 "github.com/lrnxzz/go-craft/codec/v765"
	"github.com/lrnxzz/go-craft/codec/v765/blocks"
	"github.com/lrnxzz/go-craft/pathfinder"
)

const (
	readyPoll = 100 * time.Millisecond

	// the planner reads the terrain around the bot, so being spawned is not
	// enough: the columns it will look at have to arrive first. A server may
	// advertise a view distance far wider than that, so whichever is smaller
	// decides how long the wait lasts.
	readyRadius = 3
)

var (
	errNoRoute        = errors.New("agent: no route toward the goal")
	errNoBlockInReach = errors.New("agent: no block within reach")
)

type Agent struct {
	client  *gocraft.Client
	session *v765.Session
	physics *gocraft.Physics

	spawns  sync.Once
	spawned chan struct{}

	mu        sync.Mutex
	controls  gocraft.Controls
	yaw       float32
	pitch     float32
	look      bool
	miner     miner
	navigator navigator

	watching map[string][]func(Notice)
	guarding map[string][]func(Intent)
	pending  []Notice

	ticks    uint64
	snapshot Snapshot
}

type Snapshot struct {
	Tick     uint64
	Position gocraft.Vec3d
	Yaw      float32
	Pitch    float32
	OnGround bool
	Health   float32
}

func (a *Agent) Snapshot() Snapshot {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.snapshot
}

func Join(ctx context.Context, address, username string) (*Agent, error) {
	host, port, err := splitAddress(address)
	if err != nil {
		return nil, err
	}

	conn, err := gocraft.Dial(ctx, net.JoinHostPort(host, strconv.Itoa(int(port))))
	if err != nil {
		return nil, err
	}

	client := gocraft.NewClient(conn, v765.Protocol())
	a := &Agent{
		client:   client,
		physics:  gocraft.NewPhysics(blocks.Collision),
		spawned:  make(chan struct{}),
		watching: map[string][]func(Notice){},
		guarding: map[string][]func(Intent){},
	}

	session, err := v765.Join(client, host, port, username, nil)
	if err != nil {
		client.Close()

		return nil, err
	}
	a.session = session
	a.miner = miner{digger: session}
	a.observe()

	client.Tick(gocraft.TickRate, a.tick)

	return a, nil
}

func (a *Agent) Run(ctx context.Context) error {
	return a.client.Run(ctx)
}

func (a *Agent) Spawned() <-chan struct{} {
	return a.spawned
}

func (a *Agent) Ready(ctx context.Context) error {
	select {
	case <-a.spawned:
	case <-ctx.Done():
		return ctx.Err()
	}

	radius := min(readyRadius, a.session.ViewDistance())
	for !a.session.World().Surrounds(a.session.Player().Position.Floor(), radius) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(readyPoll):
		}
	}

	return nil
}

func (a *Agent) World() *gocraft.World {
	return a.session.World()
}

func (a *Agent) Player() *gocraft.Player {
	return a.session.Player()
}

func (a *Agent) SetControls(controls gocraft.Controls) {
	a.mu.Lock()
	a.controls = controls
	a.mu.Unlock()
}

func (a *Agent) Look(yaw, pitch float32) {
	a.mu.Lock()
	a.yaw, a.pitch, a.look = yaw, pitch, true
	a.mu.Unlock()
}

func (a *Agent) LookAt(target gocraft.Vec3d) {
	player := a.session.Player()
	yaw, pitch := gocraft.LookAngles(player.Eye(), target)

	a.Look(yaw, pitch)
}

func (a *Agent) Target(reach float64) (gocraft.RayHit, bool) {
	player := a.session.Player()

	return a.session.World().Raycast(player.Eye(), player.LookDirection(), reach, blocks.Solid)
}

func (a *Agent) Inventory() *gocraft.Inventory {
	return a.session.Inventory()
}

func (a *Agent) SelectHotbar(index int) error {
	return a.session.SelectHotbar(index)
}

func (a *Agent) Hold(item gocraft.ItemID) error {
	inventory := a.session.Inventory()

	if inventory.Held().Is(item) {
		return nil
	}

	for index := range gocraft.HotbarSize {
		if inventory.Hotbar(index).Is(item) {
			return a.session.SelectHotbar(index)
		}
	}

	slot, ok := inventory.FindItem(item)
	if !ok {
		return fmt.Errorf("agent: item %d is not in the inventory", item)
	}

	return a.session.SwapWithHotbar(slot, inventory.HeldIndex())
}

func (a *Agent) SwapHands() error {
	inventory := a.session.Inventory()

	return a.session.SwapWithOffhand(gocraft.HotbarSlot(inventory.HeldIndex()))
}

func (a *Agent) ClickSlot(slot int) error {
	return a.session.ClickSlot(slot)
}

func (a *Agent) Chat(message string) error {
	proposed := &Chatting{Message: message}
	if err := a.allowed(proposed); err != nil {
		return err
	}

	command, addressed := strings.CutPrefix(message, "/")
	if addressed {
		return a.session.SendCommand(command)
	}

	return a.session.SendChat(message)
}

// observe turns the wire into notices. The session keeps its single chat
// listener — this one — and the fan-out to many happens here, which is what
// stops two plugins from silently unsubscribing each other.
func (a *Agent) observe() {
	a.session.OnChat(func(line string) {
		a.post(ChatReceived{Line: line})
	})

	gocraft.On(a.client, a.noticeBlock)
	gocraft.On(a.client, a.noticeHealth)
}

func (a *Agent) noticeBlock(_ *gocraft.Client, p *v765.BlockUpdate) error {
	change := p.Change()

	a.post(BlockChanged{
		At:    gocraft.At(change.X, change.Y, change.Z),
		State: change.State,
	})

	return nil
}

func (a *Agent) noticeHealth(_ *gocraft.Client, p *v765.SetHealth) error {
	a.post(HealthChanged{
		Health: p.Health.Float32(),
		Food:   float32(p.Food.Int()),
	})

	return nil
}

func (a *Agent) Carried() gocraft.ItemStack {
	return a.session.Carried()
}

// Dig blocks until the world settles the block, so a caller that must not wait
// runs it in a goroutine and drops the result
func (a *Agent) Dig(ctx context.Context, reach float64) (gocraft.RayHit, error) {
	hit, ok := a.Target(reach)
	if !ok {
		return gocraft.RayHit{}, errNoBlockInReach
	}

	proposed := &Digging{
		Block: hit.Block,
		State: hit.State,
		Tool:  a.session.Inventory().Held().Item,
	}
	if err := a.allowed(proposed); err != nil {
		return gocraft.RayHit{}, err
	}

	a.mu.Lock()
	finished, err := a.miner.begin(hit, reach, a.session.Player().GameMode, a.session.Inventory().Held().Item)
	a.mu.Unlock()

	if err != nil {
		return gocraft.RayHit{}, err
	}

	select {
	case err := <-finished:
		if err != nil {
			return gocraft.RayHit{}, err
		}

		return hit, nil
	case <-ctx.Done():
		_ = a.StopDigging()

		return gocraft.RayHit{}, ctx.Err()
	}
}

func (a *Agent) StopDigging() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.miner.abandon()
}

func (a *Agent) Digging() bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	_, active := a.miner.excavating()

	return active
}

func (a *Agent) Excavation() (gocraft.Position, float64, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.miner.excavation()
}

func (a *Agent) excavate() {
	a.mu.Lock()
	defer a.mu.Unlock()

	reach, active := a.miner.excavating()
	if !active {
		return
	}

	hit, sighted := a.Target(reach)
	_ = a.miner.tick(hit, sighted, a.session.Inventory().Held().Item)
}

func (a *Agent) Place(reach float64) error {
	hit, ok := a.Target(reach)
	if !ok {
		return errNoBlockInReach
	}

	proposed := &Placing{
		Block: hit.Block.Neighbor(hit.Face),
		Item:  a.session.Inventory().Held().Item,
	}
	if err := a.allowed(proposed); err != nil {
		return err
	}

	return a.session.PlaceBlock(hit)
}

// Navigate blocks until the bot arrives, gives up, or the context ends; the walk
// itself is driven by the tick loop, so a caller that wants to do something else
// meanwhile runs this in a goroutine
func (a *Agent) Navigate(ctx context.Context, goal pathfinder.Goal) (gocraft.Position, error) {
	proposed := &Navigating{Goal: goal}
	if err := a.allowed(proposed); err != nil {
		return gocraft.Position{}, err
	}

	from := a.session.Player().Position.Floor()
	planner := pathfinder.NewPlanner(a.session.World(), blocks.Terrain(), a.loadout())

	route, ok := planner.Plan(from, goal)
	if !ok {
		return gocraft.Position{}, errNoRoute
	}

	done := make(chan arrival, 1)

	a.mu.Lock()
	a.navigator.follow(route, done)
	a.mu.Unlock()

	select {
	case reached := <-done:
		a.post(Arrived{
			At:     reached.at,
			Reason: reached.err,
		})

		return reached.at, reached.err
	case <-ctx.Done():
		a.Stop()

		return gocraft.Position{}, ctx.Err()
	}
}

// the route is planned against what the bot is carrying right now, and the held
// tool is what sets how expensive mining looks.
//
// Scaffold stays at zero on purpose: the planner can bridge, but placing the
// first block over a drop needs the vanilla sneak-bridge stance — body hanging
// past the edge, aiming back and down — because from on top of the support the
// crosshair always reaches its upper face first. Until the navigator can hold
// that stance, asking for a bridge would only plan a step it cannot carry out.
func (a *Agent) loadout() pathfinder.Loadout {
	return pathfinder.Loadout{
		Tool:    a.session.Inventory().Held().Item,
		Digging: a.session.Player().GameMode != gocraft.Adventure,
	}
}

func (a *Agent) Route() ([]gocraft.Position, int) {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.navigator.route()
}

func (a *Agent) Stop() {
	a.mu.Lock()
	a.navigator.abandon(errNavigationStopped)
	a.controls = gocraft.Controls{}
	a.mu.Unlock()
}

func (a *Agent) tick() {
	if !a.session.Spawned() {
		return
	}

	player := a.session.Player()
	a.spawns.Do(func() {
		a.post(Spawned{At: player.Position})
		close(a.spawned)
	})

	a.deliver()
	a.navigate(player)

	a.mu.Lock()
	controls := a.controls
	yaw, pitch, look := a.yaw, a.pitch, a.look
	a.mu.Unlock()

	// the aim has to land on the player before the miner raycasts, otherwise a
	// dig started this tick is judged against last tick's crosshair and dropped
	if look {
		player.Yaw = yaw
		player.Pitch = pitch
	}

	a.excavate()

	a.physics.Tick(a.session.World(), player, controls)
	_ = a.session.SendPosition()

	a.mu.Lock()
	a.ticks++
	a.snapshot = Snapshot{
		Tick:     a.ticks,
		Position: player.Position,
		Yaw:      player.Yaw,
		Pitch:    player.Pitch,
		OnGround: player.OnGround,
		Health:   player.Health,
	}
	a.mu.Unlock()
}

func (a *Agent) navigate(player *gocraft.Player) {
	a.mu.Lock()
	command, navigating := a.navigator.tick(a.session.World(), player)
	if navigating {
		a.controls = command.controls
		a.yaw = command.yaw
		a.pitch = command.pitch
		a.look = true
	}
	a.mu.Unlock()

	if !navigating {
		return
	}

	switch command.action {
	case pathfinder.Break:
		a.mine(player, command)
	case pathfinder.Place:
		a.build(player, command)
	}
}

// the crosshair only lands on the block a tick after the look is sent, so both
// actions verify what they are actually pointing at before committing
func (a *Agent) mine(player *gocraft.Player, command order) {
	if a.Digging() {
		return
	}

	hit, sighted := a.sight(player, command)
	if !sighted || hit.Block != command.target {
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// the navigator watches the world to know the block fell, so the dig channel
	// has no reader here and the buffered send simply drops on the floor
	_, _ = a.miner.begin(hit, gocraft.BlockReach, player.GameMode, a.session.Inventory().Held().Item)
}

func (a *Agent) build(player *gocraft.Player, command order) {
	hit, sighted := a.sight(player, command)
	if !sighted || hit.Block.Neighbor(hit.Face) != command.target {
		return
	}

	_ = a.session.PlaceBlock(hit)
}

func (a *Agent) sight(player *gocraft.Player, command order) (gocraft.RayHit, bool) {
	direction := command.aim.Sub(player.Eye())

	return a.session.World().Raycast(player.Eye(), direction, gocraft.BlockReach, blocks.Solid)
}

func splitAddress(address string) (string, uint16, error) {
	host, raw, err := net.SplitHostPort(address)
	if err != nil {
		return address, gocraft.DefaultPort.Uint16(), nil
	}

	port, err := strconv.ParseUint(raw, 10, 16)
	if err != nil {
		return "", 0, fmt.Errorf("agent: invalid port in %q", address)
	}

	return host, uint16(port), nil
}
