package agent

import (
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

type miner struct {
	digger excavator
	dig    *excavation
}

// begin refuses outright when the block cannot be mined at all; whatever depends
// on the world lands on the returned channel instead, which already carries the
// answer when the block gave way on the first strike
func (m *miner) begin(hit gocraft.RayHit, reach float64, mode gocraft.GameMode, held gocraft.ItemID) (<-chan error, error) {
	if err := m.abandon(); err != nil {
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

	return dig.finished, m.finish()
}

// finish retires the dig in progress and tells whoever asked for the block how it
// went; the channel is buffered for exactly this one send, so the tick loop that
// calls it never waits for a reader
func (m *miner) finish() error {
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
	dig := m.dig
	if dig == nil {
		return nil
	}

	m.dig = nil
	dig.finished <- errDigAbandoned

	return m.digger.CancelDigging(dig.hit)
}

func (m *miner) excavating() (float64, bool) {
	if m.dig == nil {
		return 0, false
	}

	return m.dig.reach, true
}

func (m *miner) excavation() (gocraft.Position, float64, bool) {
	if m.dig == nil {
		return gocraft.Position{}, 0, false
	}

	return m.dig.hit.Block, m.dig.progress, true
}

func (m *miner) tick(target gocraft.RayHit, sighted bool, held gocraft.ItemID) error {
	dig := m.dig
	if dig == nil {
		return nil
	}

	if !sighted || target.Block != dig.hit.Block {
		return m.abandon()
	}

	damage, breakable := blocks.DigDamage(target.State, held)
	if !breakable {
		return m.abandon()
	}

	dig.progress += damage
	if dig.progress < 1 {
		return nil
	}

	return m.finish()
}
