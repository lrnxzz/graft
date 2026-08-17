package agent

import (
	"sync"

	graft "github.com/lrnxzz/graft"
)

// A Snapshot is what the bot last reported about itself, all of it from the same
// tick so the fields never disagree with each other
type Snapshot struct {
	Tick     uint64
	Position graft.Vec3d
	Yaw      float32
	Pitch    float32
	OnGround bool
	Health   float32
}

// latest is written by the tick and read from anywhere. It carries its own lock
// because it flows the opposite way to steering, and sharing one would make a
// reader wait on a writer it never meets.
type latest struct {
	mu    sync.Mutex
	ticks uint64
	seen  Snapshot
}

func (l *latest) publish(player *graft.Player) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.ticks++
	l.seen = Snapshot{
		Tick:     l.ticks,
		Position: player.Position,
		Yaw:      player.Yaw,
		Pitch:    player.Pitch,
		OnGround: player.OnGround,
		Health:   player.Health,
	}
}

func (l *latest) read() Snapshot {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.seen
}
