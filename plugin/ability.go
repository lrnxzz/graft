package plugin

import "github.com/dop251/goja"

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

		return func() any { return read(capable) }
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
		{Name: "name", Needs: always, Read: sensed(func(b Bot) any { return b.Name() })},
		{Name: "position", Needs: always, Read: sensed(func(b Bot) any { return b.Position() })},
		{Name: "health", Needs: always, Read: sensed(func(v Vitals) any { return v.Health() })},
		{Name: "food", Needs: always, Read: sensed(func(v Vitals) any { return v.Food() })},
		{Name: "onGround", Needs: always, Read: sensed(func(v Vitals) any { return v.OnGround() })},
		{Name: "looking", Needs: always, Read: sensed(sighting)},
		{Name: "held", Needs: MayInventory, Read: sensed(func(h Holder) any { return string(h.Held()) })},
		{Name: "inventory", Needs: MayInventory, Read: sensed(func(h Holder) any { return h.Inventory() })},
	}
}

func sighting(eyes Sighted) any {
	target, sighted := eyes.Looking()
	if !sighted {
		return nil
	}

	return target
}

func Abilities() []Ability {
	return []Ability{
		{Name: "goto", Needs: MayMove, Bind: able(func(r *Runtime, w Walker) any {
			return func(at goja.Value) (Vec3, error) { return w.Goto(r.vec3(at)) }
		})},
		{Name: "look", Needs: MayMove, Bind: able(func(r *Runtime, w Walker) any {
			return func(at goja.Value) { w.Look(r.vec3(at)) }
		})},
		{Name: "pursue", Needs: MayMove, Bind: able(func(r *Runtime, p Pursuer) any {
			return func(goal goja.Value) error { return p.Pursue(asGoal(r.reading(goal))) }
		})},
		{Name: "abandon", Needs: MayMove, Bind: able(func(_ *Runtime, p Pursuer) any {
			return p.Abandon
		})},
		{Name: "dig", Needs: MayDig, Bind: able(func(r *Runtime, d Digger) any {
			return func(at goja.Value) error { return d.Dig(r.vec3(at)) }
		})},
		{Name: "place", Needs: MayPlace, Bind: able(func(r *Runtime, b Builder) any {
			return func(at goja.Value) error { return b.Place(r.vec3(at)) }
		})},
		{Name: "say", Needs: MayChat, Bind: able(func(_ *Runtime, s Speaker) any {
			return s.Say
		})},
		{Name: "hold", Needs: MayInventory, Bind: able(func(_ *Runtime, h Holder) any {
			return func(item string) error { return h.Hold(Item(item)) }
		})},
		{Name: "count", Needs: MayInventory, Bind: able(func(_ *Runtime, h Holder) any {
			return func(item string) int { return h.Count(Item(item)) }
		})},
		{Name: "blockAt", Needs: always, Bind: able(func(r *Runtime, s Sighted) any {
			return func(at goja.Value) string { return string(s.BlockAt(r.vec3(at))) }
		})},
		{Name: "on", Needs: always, Bind: able(func(r *Runtime, w Watcher) any {
			return func(event string, handle goja.Callable) bool {
				return w.Watch(event, func(data map[string]any) {
					_, _ = handle(goja.Undefined(), r.vm.ToValue(data))
				})
			}
		})},
		{Name: "before", Needs: always, Bind: able(func(r *Runtime, g Guard) any {
			return func(intent string, handle goja.Callable) bool {
				return g.Guard(intent, func(data map[string]any) string {
					return r.veto(handle, data)
				})
			}
		})},
	}
}

func (r *Runtime) vec3(value goja.Value) Vec3 {
	return vec3Of(r.reading(value))
}

func (r *Runtime) veto(handle goja.Callable, data map[string]any) string {
	refused := ""

	event := r.vm.NewObject()
	bridge := into(r.vm, event).all(data)
	bridge.set("cancel", func(reason string) {
		if reason == "" {
			reason = "refused by a plugin"
		}

		refused = reason
	})
	if bridge.done() != nil {
		return ""
	}

	if _, err := handle(goja.Undefined(), event); err != nil {
		return ""
	}

	return refused
}

func granted(permissions []Permission, needs Permission) bool {
	if needs == always {
		return true
	}

	for _, permission := range permissions {
		if permission == needs {
			return true
		}
	}

	return false
}
