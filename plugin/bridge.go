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

	for _, word := range Words() {
		bridge.set(word.Name, r.spoken(word.Build))
	}

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

	for _, ability := range Abilities() {
		if !granted(permissions, ability.Needs) {
			continue
		}

		if ability.Read != nil {
			if read := ability.Read(bot); read != nil {
				bridge.getter(ability.Name, read)
			}

			continue
		}

		if bound := ability.Bind(r, bot); bound != nil {
			bridge.set(ability.Name, bound)
		}
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
