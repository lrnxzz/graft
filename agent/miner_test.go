package agent

import (
	"errors"
	"testing"

	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/codec/v765/blocks"
	"github.com/lrnxzz/graft/codec/v765/items"
)

type digs struct {
	started  int
	canceled int
	finished int
}

func (d *digs) StartDigging(graft.RayHit) error {
	d.started++

	return nil
}

func (d *digs) CancelDigging(graft.RayHit) error {
	d.canceled++

	return nil
}

func (d *digs) FinishDigging(graft.RayHit) error {
	d.finished++

	return nil
}

func TestMinerFinishesAfterVanillaBreakTime(t *testing.T) {
	d := &digs{}
	m := miner{digger: d}

	hit := graft.RayHit{
		Block: graft.Position{
			X: 1,
			Y: 64,
			Z: 1,
		},
		State: blocks.Stone,
	}

	finished, err := m.begin(hit, 4.5, graft.Survival, items.Air)
	if err != nil {
		t.Fatal(err)
	}
	if d.started != 1 {
		t.Fatalf("started = %d, want 1", d.started)
	}

	ticks, _ := blocks.BreakTicks(blocks.Stone, items.Air)
	for range ticks - 2 {
		if err := m.tick(hit, true, items.Air); err != nil {
			t.Fatal(err)
		}
	}
	if d.finished != 0 {
		t.Fatal("finished before the vanilla break time")
	}

	if err := m.tick(hit, true, items.Air); err != nil {
		t.Fatal(err)
	}
	if d.finished != 1 {
		t.Errorf("finished = %d, want 1 exactly at the break time", d.finished)
	}

	if err := <-finished; err != nil {
		t.Fatal(err)
	}
	if _, active := m.current(); active {
		t.Error("miner should be idle after finishing")
	}
}

func TestMinerSwitchingToolsSpeedsUpMidDig(t *testing.T) {
	d := &digs{}
	m := miner{digger: d}

	hit := graft.RayHit{
		Block: graft.Position{
			X: 0,
			Y: 10,
			Z: 0,
		},
		State: blocks.Stone,
	}

	_, err := m.begin(hit, 4.5, graft.Survival, items.Air)
	if err != nil {
		t.Fatal(err)
	}

	ticks, _ := blocks.BreakTicks(blocks.Stone, items.NetheritePickaxe)
	for range ticks {
		if err := m.tick(hit, true, items.NetheritePickaxe); err != nil {
			t.Fatal(err)
		}
	}

	if d.finished != 1 {
		t.Errorf("finished = %d, want 1 (netherite pickaxe must speed up an ongoing dig)", d.finished)
	}
}

func TestMinerCancelsWhenLookingAway(t *testing.T) {
	d := &digs{}
	m := miner{digger: d}

	hit := graft.RayHit{
		Block: graft.Position{
			X: 3,
			Y: 64,
			Z: 3,
		},
		State: blocks.Stone,
	}

	finished, err := m.begin(hit, 4.5, graft.Survival, items.Air)
	if err != nil {
		t.Fatal(err)
	}

	elsewhere := hit
	elsewhere.Block = hit.Block.Add(1, 0, 0)
	if err := m.tick(elsewhere, true, items.Air); err != nil {
		t.Fatal(err)
	}

	if d.canceled != 1 {
		t.Errorf("canceled = %d, want 1 when the target changes", d.canceled)
	}

	if err := <-finished; !errors.Is(err, errDigAbandoned) {
		t.Errorf("dig err = %v, want abandoned", err)
	}
	if _, active := m.current(); active {
		t.Error("miner should be idle after abandoning")
	}
}

func TestMinerRestartResolvesThePreviousDigAsAbandoned(t *testing.T) {
	d := &digs{}
	m := miner{digger: d}

	first := graft.RayHit{
		Block: graft.Position{
			X: 0,
			Y: 64,
			Z: 0,
		},
		State: blocks.Stone,
	}
	second := first
	second.Block = first.Block.Add(0, 0, 1)

	abandoned, err := m.begin(first, 4.5, graft.Survival, items.Air)
	if err != nil {
		t.Fatal(err)
	}
	_, restarted := m.begin(second, 4.5, graft.Survival, items.Air)
	if restarted != nil {
		t.Fatal(restarted)
	}

	if d.canceled != 1 || d.started != 2 {
		t.Errorf("canceled = %d, started = %d, want 1 and 2", d.canceled, d.started)
	}

	if err := <-abandoned; !errors.Is(err, errDigAbandoned) {
		t.Errorf("first dig err = %v, want abandoned", err)
	}
}

func TestMinerCreativeBreaksWithStartOnly(t *testing.T) {
	d := &digs{}
	m := miner{digger: d}

	hit := graft.RayHit{
		Block: graft.Position{
			X: 5,
			Y: 64,
			Z: 5,
		},
		State: blocks.Stone,
	}

	finished, err := m.begin(hit, 4.5, graft.Creative, items.Air)
	if err != nil {
		t.Fatal(err)
	}

	if d.started != 1 || d.finished != 0 {
		t.Errorf("started = %d, finished = %d, want 1 and 0", d.started, d.finished)
	}

	if err := <-finished; err != nil {
		t.Errorf("creative dig err = %v, want nil", err)
	}
	if _, active := m.current(); active {
		t.Error("creative digging should not leave the miner active")
	}
}

func TestMinerRefusesUnbreakableBlocks(t *testing.T) {
	d := &digs{}
	m := miner{digger: d}

	hit := graft.RayHit{
		Block: graft.Position{
			X: 0,
			Y: -60,
			Z: 0,
		},
		State: blocks.Bedrock,
	}

	_, err := m.begin(hit, 4.5, graft.Survival, items.NetheritePickaxe)
	if err == nil {
		t.Fatal("bedrock should be refused")
	}
	if d.started != 0 {
		t.Errorf("started = %d, want 0", d.started)
	}
}
