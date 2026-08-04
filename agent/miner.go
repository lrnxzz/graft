package agent

import (
	"sync"

	"errors"
	"fmt"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/codec/v765/blocks"
)

var errDigAbandoned = errors.New("agent: digging abandoned")

type excavator interface {
	StartDigging(gocraft.RayHit) error
	CancelDigging(gocraft.RayHit) error
	FinishDigging(gocraft.RayHit) error
}

type excavation struct {
	hit      gocraft.RayHit
	reach    float64
	progress float64
	instant  bool
	finished chan error
}

// the lock lives here rather than on the agent: the tick and a caller reach the
// same excavation from different goroutines, and nothing else waits on it.
//
// Every method that takes it delegates to an unlocked core, because tick calls
// what abandon and finish do and a mutex here is not reentrant.
type miner struct {
	mu     sync.Mutex
	digger excavator
	dig    *excavation
}

// begin refuses outright when the block cannot be mined at all; whatever depends
// on the world lands on the returned channel instead, which already carries the
// answer when the block gave way on the first strike
func (m *miner) begin(hit gocraft.RayHit, reach float64, mode gocraft.GameMode, held gocraft.ItemID) (<-chan error, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.stop(); err != nil {
		return nil, err
	}

	// creative breaks on the start packet alone, so its first strike is total
	instant := mode == gocraft.Creative
	damage := 1.0

	if !instant {
		dealt, breakable := blocks.DigDamage(hit.State, held)
		if !breakable {
			return nil, fmt.Errorf("agent: block state %d cannot be broken", hit.State)
		}

		damage = dealt
	}

	if err := m.digger.StartDigging(hit); err != nil {
		return nil, err
	}

	dig := &excavation{
		hit:      hit,
		reach:    reach,
		progress: damage,
		instant:  instant,
		finished: make(chan error, 1),
	}
	m.dig = dig

	if damage < 1 {
		return dig.finished, nil
	}

	return dig.finished, m.complete()
}

// finish retires the dig in progress and tells whoever asked for the block how it
// went; the channel is buffered for exactly this one send, so the tick loop that
// calls it never waits for a reader
func (m *miner) complete() error {
	dig := m.dig
	m.dig = nil

	var err error
	if !dig.instant {
		err = m.digger.FinishDigging(dig.hit)
	}

	dig.finished <- err

	return err
}

func (m *miner) abandon() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.stop()
}

func (m *miner) stop() error {
	dig := m.dig
	if dig == nil {
		return nil
	}

	m.dig = nil
	dig.finished <- errDigAbandoned

	return m.digger.CancelDigging(dig.hit)
}

func (m *miner) excavating() (float64, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.dig == nil {
		return 0, false
	}

	return m.dig.reach, true
}

func (m *miner) excavation() (Excavating, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.dig == nil {
		return Excavating{}, false
	}

	underway := Excavating{
		Block:    m.dig.hit.Block,
		Progress: m.dig.progress,
	}

	return underway, true
}

func (m *miner) tick(target gocraft.RayHit, sighted bool, held gocraft.ItemID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	dig := m.dig
	if dig == nil {
		return nil
	}

	if !sighted || target.Block != dig.hit.Block {
		return m.stop()
	}

	damage, breakable := blocks.DigDamage(target.State, held)
	if !breakable {
		return m.stop()
	}

	dig.progress += damage
	if dig.progress < 1 {
		return nil
	}

	return m.complete()
}
