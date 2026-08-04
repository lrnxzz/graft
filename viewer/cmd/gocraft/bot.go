package main

import (
	"context"
	"fmt"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/agent"
	"github.com/lrnxzz/go-craft/codec/v765/blocks"
	"github.com/lrnxzz/go-craft/codec/v765/items"
	"github.com/lrnxzz/go-craft/pathfinder"
	"github.com/lrnxzz/go-craft/plugin"
)

type bot struct {
	ctx   context.Context
	agent *agent.Agent
}

func (b bot) Name() string {
	return b.agent.Player().Username
}

func (b bot) Position() plugin.Vec3 {
	return spot(b.agent.Snapshot().Position)
}

func (b bot) Health() float32 {
	return b.agent.Snapshot().Health
}

func (b bot) Food() float32 {
	return float32(b.agent.Player().Food)
}

func (b bot) OnGround() bool {
	return b.agent.Snapshot().OnGround
}

func (b bot) Looking() (plugin.Target, bool) {
	hit, sighted := b.agent.Target(gocraft.BlockReach)
	if !sighted {
		return plugin.Target{}, false
	}

	return plugin.Target{
		At:    spot(hit.Block.Corner()),
		Face:  plugin.Face(hit.Face.String()),
		Block: b.BlockAt(spot(hit.Block.Corner())),
	}, true
}

func (b bot) BlockAt(at plugin.Vec3) plugin.Block {
	state, loaded := b.agent.World().BlockAt(block(at))
	if !loaded {
		return ""
	}

	known, named := blocks.Of(state)
	if !named {
		return ""
	}

	return plugin.Block(known.Name)
}

func (b bot) Held() plugin.Item {
	return named(b.agent.Inventory().Held().Item)
}

func (b bot) Inventory() []plugin.Stack {
	inventory := b.agent.Inventory()

	var carried []plugin.Stack
	for slot := range gocraft.InventorySize {
		stack := inventory.Slot(slot)
		if stack.Empty() {
			continue
		}

		carried = append(carried, plugin.Stack{
			Item:  named(stack.Item),
			Count: stack.Count,
			Slot:  slot,
		})
	}

	return carried
}

func (b bot) Count(item plugin.Item) int {
	id, known := identified(item)
	if !known {
		return 0
	}

	return b.agent.Inventory().Count(id)
}

func (b bot) Hold(item plugin.Item) error {
	id, known := identified(item)
	if !known {
		return fmt.Errorf("gocraft: no item named %s", item)
	}

	return b.agent.Hold(id)
}

func (b bot) Say(text string) error {
	return b.agent.Chat(text)
}

func (b bot) Look(at plugin.Vec3) {
	b.agent.LookAt(gocraft.Vec3(at.X, at.Y, at.Z))
}

func (b bot) Goto(at plugin.Vec3) (plugin.Vec3, error) {
	arrived, err := b.agent.Navigate(b.ctx, pathfinder.GoalAt(block(at)))
	if err != nil {
		return plugin.Vec3{}, err
	}

	return spot(arrived.Corner()), nil
}

func (b bot) Dig(at plugin.Vec3) error {
	target := block(at)
	b.agent.LookAt(target.Center())

	hit, err := b.agent.Dig(b.ctx, gocraft.BlockReach)
	if err != nil {
		return err
	}
	if hit.Block != target {
		return fmt.Errorf("gocraft: broke %v instead of %v", hit.Block, target)
	}

	return nil
}

func (b bot) Place(at plugin.Vec3) error {
	b.agent.LookAt(block(at).Center())

	return b.agent.Place(gocraft.BlockReach)
}

func (b bot) Pursue(goal plugin.Goal) error {
	reached, err := aim(goal)
	if err != nil {
		return err
	}

	_, err = b.agent.Navigate(b.ctx, reached)

	return err
}

func (b bot) Abandon() {
	b.agent.Stop()
}

func aim(goal plugin.Goal) (pathfinder.Goal, error) {
	switch goal.Type {
	case plugin.GoalAt:
		return pathfinder.GoalAt(block(goal.At)), nil
	case plugin.GoalNear:
		return pathfinder.GoalNear(block(goal.At), int(goal.Radius)), nil
	default:
		return nil, fmt.Errorf("gocraft: the %s goal is not served yet", goal.Type)
	}
}

func spot(at gocraft.Vec3d) plugin.Vec3 {
	return plugin.Vec3{
		X: at.X,
		Y: at.Y,
		Z: at.Z,
	}
}

func block(at plugin.Vec3) gocraft.Position {
	return gocraft.At(int(at.X), int(at.Y), int(at.Z))
}

func named(id gocraft.ItemID) plugin.Item {
	known, is := items.Of(id)
	if !is {
		return ""
	}

	return plugin.Item(known.Name)
}

func identified(item plugin.Item) (gocraft.ItemID, bool) {
	known, is := items.Named(gocraft.Identifier(item))
	if !is {
		return 0, false
	}

	return known.ID, true
}
