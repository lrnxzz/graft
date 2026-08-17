package agent

import (
	"sync"

	graft "github.com/lrnxzz/graft"
)

// steering is what a caller wants the body to do, written from anywhere and read
// once a tick. It is the only direction the caller writes in: everything the bot
// reports back goes the other way, through latest.
type steering struct {
	mu       sync.Mutex
	controls graft.Controls
	yaw      float32
	pitch    float32
	aimed    bool
}

func (s *steering) hold(controls graft.Controls) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.controls = controls
}

func (s *steering) aim(yaw, pitch float32) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.yaw, s.pitch, s.aimed = yaw, pitch, true
}

// wanted is read once at the top of the tick so the rest of it works from a
// settled intent rather than one a caller may still be changing
func (s *steering) wanted() (graft.Controls, float32, float32, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.controls, s.yaw, s.pitch, s.aimed
}
