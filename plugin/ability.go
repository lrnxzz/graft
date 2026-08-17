package plugin

import (
	"slices"

	"github.com/dop251/goja"
)

const always Permission = ""

type Sense struct {
	Name  string
	Needs Permission
	Read  func(Bot) func() any
}

type Ability struct {
	Name  string
	Needs Permission
	Bind  func(*Runtime, Bot) any
}

func sensed[C any](read func(C) any) func(Bot) func() any {
	return func(bot Bot) func() any {
		capable, is := bot.(C)
		if !is {
			return nil
		}

		return func() any {
			return read(capable)
		}
	}
}

func able[C any](bind func(*Runtime, C) any) func(*Runtime, Bot) any {
	return func(r *Runtime, bot Bot) any {
		capable, is := bot.(C)
		if !is {
			return nil
		}

		return bind(r, capable)
	}
}

func Senses() []Sense {
	return []Sense{
		{
			Name:  "name",
			Needs: always,
			Read:  sensed(readName),
		},
		{
			Name:  "position",
			Needs: always,
			Read:  sensed(readPosition),
		},
		{
			Name:  "health",
			Needs: always,
			Read:  sensed(readHealth),
		},
		{
			Name:  "food",
			Needs: always,
			Read:  sensed(readFood),
		},
		{
			Name:  "onGround",
			Needs: always,
			Read:  sensed(readOnGround),
		},
		{
			Name:  "looking",
			Needs: always,
			Read:  sensed(readLooking),
		},
		{
			Name:  "held",
			Needs: MayInventory,
			Read:  sensed(readHeld),
		},
		{
			Name:  "inventory",
			Needs: MayInventory,
			Read:  sensed(readInventory),
		},
	}
}

func readName(bot Bot) any {
	return bot.Name()
}

func readPosition(bot Bot) any {
	return bot.Position()
}

func readHealth(vitals Vitals) any {
	return vitals.Health()
}

func readFood(vitals Vitals) any {
	return vitals.Food()
}

func readOnGround(vitals Vitals) any {
	return vitals.OnGround()
}

func readHeld(holder Holder) any {
	return string(holder.Held())
}

func readInventory(holder Holder) any {
	return holder.Inventory()
}

func readLooking(eyes Sighted) any {
	target, sighted := eyes.Looking()
	if !sighted {
		return nil
	}

	return target
}

func Abilities() []Ability {
	return []Ability{
		{
			Name:  "goto",
			Needs: MayMove,
			Bind:  able(bindGoto),
		},
		{
			Name:  "look",
			Needs: MayMove,
			Bind:  able(bindLook),
		},
		{
			Name:  "pursue",
			Needs: MayMove,
			Bind:  able(bindPursue),
		},
		{
			Name:  "abandon",
			Needs: MayMove,
			Bind:  able(bindAbandon),
		},
		{
			Name:  "dig",
			Needs: MayDig,
			Bind:  able(bindDig),
		},
		{
			Name:  "place",
			Needs: MayPlace,
			Bind:  able(bindPlace),
		},
		{
			Name:  "say",
			Needs: MayChat,
			Bind:  able(bindSay),
		},
		{
			Name:  "hold",
			Needs: MayInventory,
			Bind:  able(bindHold),
		},
		{
			Name:  "count",
			Needs: MayInventory,
			Bind:  able(bindCount),
		},
		{
			Name:  "blockAt",
			Needs: always,
			Bind:  able(bindBlockAt),
		},
		{
			Name:  "on",
			Needs: always,
			Bind:  able(bindOn),
		},
		{
			Name:  "before",
			Needs: always,
			Bind:  able(bindBefore),
		},
	}
}

func bindGoto(r *Runtime, walker Walker) any {
	return func(at goja.Value) (Vec3, error) {
		return walker.Goto(r.vec3(at))
	}
}

func bindLook(r *Runtime, walker Walker) any {
	return func(at goja.Value) {
		walker.Look(r.vec3(at))
	}
}

func bindPursue(r *Runtime, pursuer Pursuer) any {
	return func(goal goja.Value) error {
		return pursuer.Pursue(asGoal(r.reading(goal)))
	}
}

func bindAbandon(_ *Runtime, pursuer Pursuer) any {
	return pursuer.Abandon
}

func bindDig(r *Runtime, digger Digger) any {
	return func(at goja.Value) error {
		return digger.Dig(r.vec3(at))
	}
}

func bindPlace(r *Runtime, builder Builder) any {
	return func(at goja.Value) error {
		return builder.Place(r.vec3(at))
	}
}

func bindSay(_ *Runtime, speaker Speaker) any {
	return speaker.Say
}

func bindHold(_ *Runtime, holder Holder) any {
	return func(item string) error {
		return holder.Hold(Item(item))
	}
}

func bindCount(_ *Runtime, holder Holder) any {
	return func(item string) int {
		return holder.Count(Item(item))
	}
}

func bindBlockAt(r *Runtime, eyes Sighted) any {
	return func(at goja.Value) string {
		return string(eyes.BlockAt(r.vec3(at)))
	}
}

func bindOn(r *Runtime, watcher Watcher) any {
	return func(event string, handle goja.Callable) bool {
		raise := func(notice Notice) {
			_, _ = handle(goja.Undefined(), r.vm.ToValue(notice))
		}

		return watcher.Watch(event, raise)
	}
}

func bindBefore(r *Runtime, guard Guard) any {
	return func(intent string, handle goja.Callable) bool {
		ask := func(about Intent) string {
			return r.veto(handle, about)
		}

		return guard.Guard(intent, ask)
	}
}

func (r *Runtime) vec3(value goja.Value) Vec3 {
	return vec3Of(r.reading(value))
}

// an empty answer lets the intent through; anything else is the refusal
func (r *Runtime) veto(handle goja.Callable, about Intent) string {
	refused := ""

	cancel := func(reason string) {
		if reason == "" {
			reason = "refused by a plugin"
		}

		refused = reason
	}

	// the intent arrives as a goja-wrapped struct, which takes no new properties,
	// so its fields are copied onto a plain object that cancel can join
	wrapped := r.vm.ToValue(about).ToObject(r.vm)

	event := r.vm.NewObject()
	bridge := into(r.vm, event)

	for _, name := range wrapped.Keys() {
		bridge.set(name, wrapped.Get(name))
	}

	bridge.set("cancel", cancel)

	if bridge.done() != nil {
		return ""
	}

	if _, err := handle(goja.Undefined(), event); err != nil {
		return ""
	}

	return refused
}

func granted(permissions []Permission, needs Permission) bool {
	return needs == always || slices.Contains(permissions, needs)
}
