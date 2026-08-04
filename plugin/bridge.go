package plugin

import "github.com/dop251/goja"

func (r *Runtime) reading(value goja.Value) reading {
	return reader(r.vm, value)
}

func (r *Runtime) install() error {
	api := r.vm.NewObject()

	bridge := into(r.vm, api).all(map[string]any{
		"plugin":  r.declare,
		"command": r.command,
		"when":    r.when,
	})

	for _, spec := range Components() {
		bridge.set(spec.Tag, r.component(spec.Build))
	}
	for _, spec := range Goals() {
		bridge.set(string(spec.Kind), r.goal(spec.Build))
	}
	for _, spec := range Markers() {
		bridge.set(string(spec.Kind), r.marker(spec.Build))
	}

	bridge.all(map[string]any{
		string(GoalSequence): r.nesting(GoalSequence),
		string(GoalRace):     r.nesting(GoalRace),
		string(GoalRepeat):   r.repeat,
		"within":             r.within,
	})

	if err := bridge.done(); err != nil {
		return err
	}

	return into(r.vm, r.vm.GlobalObject()).all(map[string]any{
		"require":   r.requiring(api),
		jsxFactory:  r.element,
		jsxFragment: r.fragment,
	}).done()
}

func (r *Runtime) requiring(api *goja.Object) func(string) (goja.Value, error) {
	return func(name string) (goja.Value, error) {
		if name != apiModule {
			return nil, &UnknownImport{
				Plugin: r.source.Name,
				Module: name,
			}
		}

		return api, nil
	}
}

type UnknownImport struct {
	Plugin string
	Module string
}

func (u *UnknownImport) Error() string {
	return "plugin " + u.Plugin + ": import of " + u.Module + " is not allowed"
}

func (r *Runtime) declare(spec *goja.Object) *goja.Object {
	return spec
}

func (r *Runtime) component(build func(reading) Node) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		return r.vm.ToValue(build(r.reading(call.Argument(0))))
	}
}

func (r *Runtime) goal(build func(reading, reading) Goal) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		return r.vm.ToValue(build(r.reading(call.Argument(0)), r.reading(call.Argument(1))))
	}
}

func (r *Runtime) nesting(kind GoalKind) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		goal := Goal{Kind: kind}
		for _, argument := range call.Arguments {
			goal.Inner = append(goal.Inner, asGoal(r.reading(argument)))
		}

		return r.vm.ToValue(goal)
	}
}

func (r *Runtime) repeat(call goja.FunctionCall) goja.Value {
	return r.vm.ToValue(Goal{
		Kind:  GoalRepeat,
		Times: r.reading(call.Argument(1)).count(),
		Inner: []Goal{asGoal(r.reading(call.Argument(0)))},
	})
}

func (r *Runtime) within(call goja.FunctionCall) goja.Value {
	area := r.vm.NewObject()
	_ = area.Set("radius", call.Argument(0))

	return area
}

func (r *Runtime) marker(build func([]reading) Marker) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		args := make([]reading, 0, len(call.Arguments))
		for _, argument := range call.Arguments {
			args = append(args, r.reading(argument))
		}

		return r.vm.ToValue(build(args))
	}
}

func (r *Runtime) element(call goja.FunctionCall) goja.Value {
	build := r.reading(call.Argument(0)).callable()
	if build == nil {
		panic(r.vm.ToValue("unknown element in jsx"))
	}

	attributes := r.vm.NewObject()
	given := r.reading(call.Argument(1))
	for _, name := range given.keys() {
		_ = attributes.Set(name, given.field(name).value)
	}
	if len(call.Arguments) > 2 {
		_ = attributes.Set("children", r.vm.ToValue(call.Arguments[2:]))
	}

	built, err := build(goja.Undefined(), attributes)
	if err != nil {
		panic(r.vm.ToValue(err.Error()))
	}

	return built
}

func (r *Runtime) fragment(call goja.FunctionCall) goja.Value {
	return r.vm.ToValue(stack{
		frame: frame{children: childrenOf(r.reading(call.Argument(0)))},
	})
}

func (r *Runtime) botObject(bot Bot, permissions []Permission) (*goja.Object, error) {
	object := r.vm.NewObject()
	if bot == nil {
		return object, nil
	}

	bridge := into(r.vm, object)

	for _, sense := range Senses() {
		if !granted(permissions, sense.Needs) {
			continue
		}

		read := sense.Read(bot)
		if read == nil {
			continue
		}

		bridge.getter(sense.Name, read)
	}

	for _, ability := range Abilities() {
		if !granted(permissions, ability.Needs) {
			continue
		}

		bound := ability.Bind(r, bot)
		if bound == nil {
			continue
		}

		bridge.set(ability.Name, bound)
	}

	return object, bridge.done()
}

func surfaceObject(vm *goja.Runtime, surface Surface) *goja.Object {
	object := vm.NewObject()
	width, height := surface.Size()

	_ = into(vm, object).all(map[string]any{
		"width":  width,
		"height": height,
		"fill": func(x, y, w, h float32, color string) {
			surface.Fill(x, y, w, h, ParseColor(color))
		},
		"text": func(text string, x, y float32, color string, scale float32) {
			surface.Text(text, x, y, scaleOr(scale), ParseColor(color).or(white))
		},
		"icon": func(item string, x, y, size float32) {
			surface.Icon(Sprite(item), x, y, size)
		},
		"measure": func(text string, scale float32) float32 {
			return surface.Measure(text, scaleOr(scale))
		},
	}).done()

	return object
}

func scaleOr(scale float32) float32 {
	if scale <= 0 {
		return defaultScale
	}

	return scale
}
