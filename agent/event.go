package agent

import (
	"sync"

	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/pathfinder"
)

// Notice is something that already happened. Notices are queued and delivered on
// the tick, so a handler always reads a settled world and may call back into the
// bot without racing the loop that produced the notice.
type Notice interface {
	Name() string
}

// Intent is something about to happen, delivered synchronously at the point of
// decision — which is the only place a refusal can still mean anything. A notice
// arriving a tick late is fine; a veto arriving a tick late is a lie.
type Intent interface {
	Name() string
	Refuse(reason string)
	Refused() (string, bool)
}

// decision is embedded by every intent so the refusal is written once
type decision struct {
	reason  string
	refused bool
}

func (d *decision) Refuse(reason string) {
	d.refused = true
	d.reason = reason
}

func (d *decision) Refused() (string, bool) {
	return d.reason, d.refused
}

type Spawned struct {
	At graft.Vec3d
}

func (Spawned) Name() string {
	return "spawned"
}

// Arrived carries why a walk ended, because ending short of the goal is a normal
// outcome and a plugin usually cares which it was
type Arrived struct {
	At     graft.Position
	Reason error
}

func (Arrived) Name() string {
	return "arrived"
}

type ChatReceived struct {
	Line string
}

func (ChatReceived) Name() string {
	return "chat"
}

// BlockChanged reports the settled state only. The previous one is gone by the
// time this is raised: the session applies the packet before the agent sees it.
type BlockChanged struct {
	At    graft.Position
	State graft.BlockState
}

func (BlockChanged) Name() string {
	return "blockChanged"
}

type HealthChanged struct {
	Health float32
	Food   float32
}

func (HealthChanged) Name() string {
	return "health"
}

type Disconnected struct {
	Reason string
}

func (Disconnected) Name() string {
	return "disconnected"
}

type Digging struct {
	decision
	Block graft.Position
	State graft.BlockState
	Tool  graft.ItemID
}

func (*Digging) Name() string {
	return "dig"
}

type Placing struct {
	decision
	Block graft.Position
	Item  graft.ItemID
}

func (*Placing) Name() string {
	return "place"
}

type Chatting struct {
	decision
	Message string
}

func (*Chatting) Name() string {
	return "chat"
}

type Navigating struct {
	decision
	Goal pathfinder.Goal
}

func (*Navigating) Name() string {
	return "navigate"
}

// On subscribes to something that already happened. Handlers run on the tick, in
// the order they were added, and one that panics or blocks only delays the bot —
// it can never drop the connection.
func On[N Notice](a *Agent, handle func(N)) {
	// the prototype is a zero value, and for the pointer notices that is nil —
	// safe only because every Name() ignores its receiver
	var prototype N

	typed := func(raised Notice) {
		only, is := raised.(N)
		if is {
			handle(only)
		}
	}

	a.audience.watch(prototype.Name(), typed)
}

// Before subscribes to something about to happen, with the chance to refuse it.
// The handler runs on whichever goroutine asked for the action, before anything
// has been sent.
func Before[I Intent](a *Agent, handle func(I)) {
	// nil-receiver Name(), same as On
	var prototype I

	typed := func(proposed Intent) {
		only, is := proposed.(I)
		if is {
			handle(only)
		}
	}

	a.audience.guard(prototype.Name(), typed)
}

// post queues a notice for the next tick
func (a *Agent) post(raised Notice) {
	a.audience.queue(raised)
}

// deliver drains the queue on the tick. Handlers run outside the lock so one is
// free to call back into the bot.
func (a *Agent) deliver() {
	for _, raised := range a.audience.drain() {
		for _, handle := range a.audience.watchers(raised.Name()) {
			handle(raised)
		}
	}
}

// allowed runs the guards for an intent and reports whether it survived them
func (a *Agent) allowed(proposed Intent) error {
	for _, handle := range a.audience.guards(proposed.Name()) {
		handle(proposed)
	}

	reason, refused := proposed.Refused()
	if !refused {
		return nil
	}

	return &Refusal{
		Intent: proposed.Name(),
		Reason: reason,
	}
}

// Refusal is what an action answers when a guard vetoed it, so a caller can tell
// a refusal apart from a failure
type Refusal struct {
	Intent string
	Reason string
}

func (r *Refusal) Error() string {
	if r.Reason == "" {
		return "agent: " + r.Intent + " was refused"
	}

	return "agent: " + r.Intent + " was refused: " + r.Reason
}

// audience is who is listening. It carries its own lock because a handler runs
// outside it and is free to call back into the bot: holding the agent's lock
// across a plugin's code would deadlock the tick on the first one that did.
type audience struct {
	mu       sync.Mutex
	watching map[string][]func(Notice)
	guarding map[string][]func(Intent)
	pending  []Notice
}

func newAudience() audience {
	return audience{
		watching: map[string][]func(Notice){},
		guarding: map[string][]func(Intent){},
	}
}

func (a *audience) watch(name string, handle func(Notice)) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.watching[name] = append(a.watching[name], handle)
}

func (a *audience) guard(name string, handle func(Intent)) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.guarding[name] = append(a.guarding[name], handle)
}

func (a *audience) queue(raised Notice) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.pending = append(a.pending, raised)
}

func (a *audience) drain() []Notice {
	a.mu.Lock()
	defer a.mu.Unlock()

	queued := a.pending
	a.pending = nil

	return queued
}

func (a *audience) watchers(name string) []func(Notice) {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.watching[name]
}

func (a *audience) guards(name string) []func(Intent) {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.guarding[name]
}
