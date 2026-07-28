package agent

import (
	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/pathfinder"
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
	At gocraft.Vec3d
}

func (Spawned) Name() string {
	return "spawned"
}

// Arrived carries why a walk ended, because ending short of the goal is a normal
// outcome and a plugin usually cares which it was
type Arrived struct {
	At     gocraft.Position
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
	At    gocraft.Position
	State gocraft.BlockState
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
	Block gocraft.Position
	State gocraft.BlockState
	Tool  gocraft.ItemID
}

func (*Digging) Name() string {
	return "dig"
}

type Placing struct {
	decision
	Block gocraft.Position
	Item  gocraft.ItemID
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
	var prototype N

	a.mu.Lock()
	a.watching[prototype.Name()] = append(a.watching[prototype.Name()], func(raised Notice) {
		typed, ok := raised.(N)
		if ok {
			handle(typed)
		}
	})
	a.mu.Unlock()
}

// Before subscribes to something about to happen, with the chance to refuse it.
// The handler runs on whichever goroutine asked for the action, before anything
// has been sent.
func Before[I Intent](a *Agent, handle func(I)) {
	var prototype I

	a.mu.Lock()
	a.guarding[prototype.Name()] = append(a.guarding[prototype.Name()], func(proposed Intent) {
		typed, ok := proposed.(I)
		if ok {
			handle(typed)
		}
	})
	a.mu.Unlock()
}

// post queues a notice for the next tick
func (a *Agent) post(raised Notice) {
	a.mu.Lock()
	a.pending = append(a.pending, raised)
	a.mu.Unlock()
}

// deliver drains the queue on the tick. Handlers run outside the lock so one is
// free to call back into the bot.
func (a *Agent) deliver() {
	a.mu.Lock()
	queued := a.pending
	a.pending = nil
	a.mu.Unlock()

	for _, raised := range queued {
		a.mu.Lock()
		handlers := a.watching[raised.Name()]
		a.mu.Unlock()

		for _, handle := range handlers {
			handle(raised)
		}
	}
}

// allowed runs the guards for an intent and reports whether it survived them
func (a *Agent) allowed(proposed Intent) error {
	a.mu.Lock()
	handlers := a.guarding[proposed.Name()]
	a.mu.Unlock()

	for _, handle := range handlers {
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
