package plugin

import (
	"fmt"

	"github.com/dop251/goja"
)

type reading struct {
	vm    *goja.Runtime
	value goja.Value
}

func reader(vm *goja.Runtime, value goja.Value) reading {
	return reading{
		vm:    vm,
		value: value,
	}
}

func (r reading) at(value goja.Value) reading {
	return reading{
		vm:    r.vm,
		value: value,
	}
}

func (r reading) missing() bool {
	return r.value == nil || goja.IsUndefined(r.value) || goja.IsNull(r.value)
}

func (r reading) object() *goja.Object {
	if r.missing() {
		return nil
	}

	object, ok := r.value.(*goja.Object)
	if !ok {
		return nil
	}

	return object
}

func (r reading) field(name string) reading {
	object := r.object()
	if object == nil {
		return reading{}
	}

	return r.at(object.Get(name))
}

func (r reading) keys() []string {
	object := r.object()
	if object == nil {
		return nil
	}

	return object.Keys()
}

func (r reading) text() string {
	if r.missing() {
		return ""
	}

	return r.value.String()
}

func (r reading) number() float32 {
	return float32(r.decimal())
}

func (r reading) decimal() float64 {
	if r.missing() {
		return 0
	}

	return r.value.ToFloat()
}

func (r reading) count() int {
	if r.missing() {
		return 0
	}

	return int(r.value.ToInteger())
}

func (r reading) flag() bool {
	if r.missing() {
		return false
	}

	return r.value.ToBoolean()
}

func (r reading) color() Color {
	return ParseColor(r.text())
}

func (r reading) callable() goja.Callable {
	if r.missing() {
		return nil
	}

	fn, ok := goja.AssertFunction(r.value)
	if !ok {
		return nil
	}

	return fn
}

func (r reading) painter() func(Surface) {
	fn := r.callable()
	if fn == nil {
		return nil
	}

	vm := r.vm

	return func(surface Surface) {
		_, _ = fn(goja.Undefined(), surfaceObject(vm, surface))
	}
}

func (r reading) callback() func() {
	fn := r.callable()
	if fn == nil {
		return nil
	}

	return func() {
		_, _ = fn(goja.Undefined())
	}
}

func (r reading) items() []reading {
	object := r.object()
	if object == nil || object.ClassName() != "Array" {
		return nil
	}

	found := make([]reading, 0, len(object.Keys()))
	for _, key := range object.Keys() {
		found = append(found, r.at(object.Get(key)))
	}

	return found
}

func nodeOf(r reading) (Node, bool) {
	if r.missing() {
		return nil, false
	}

	node, ok := r.value.Export().(Node)

	return node, ok
}

func exported[T any](r reading) (T, bool) {
	var zero T
	if r.missing() {
		return zero, false
	}

	switch value := r.value.Export().(type) {
	case *T:
		return *value, true
	case T:
		return value, true
	default:
		return zero, false
	}
}

// installing puts names on a javascript object and remembers the first failure
// instead of asking every caller to check. It also refuses a name twice: the api
// object is built from several catalogues at once, and one of them quietly
// shadowing another is exactly the mistake that would never raise an error.
type installing struct {
	vm      *goja.Runtime
	object  *goja.Object
	claimed map[string]bool
	err     error
}

func into(vm *goja.Runtime, object *goja.Object) *installing {
	return &installing{
		vm:      vm,
		object:  object,
		claimed: map[string]bool{},
	}
}

func (i *installing) set(name string, value any) {
	if i.err != nil {
		return
	}
	if i.claimed[name] {
		i.err = fmt.Errorf("plugin: two things answer to %s", name)

		return
	}

	i.claimed[name] = true
	i.err = i.object.Set(name, value)
}

func (i *installing) all(values map[string]any) *installing {
	for name, value := range values {
		i.set(name, value)
	}

	return i
}

func (i *installing) getter(name string, read func() any) *installing {
	if i.err != nil {
		return i
	}

	i.err = i.object.DefineAccessorProperty(name,
		i.vm.ToValue(func(goja.FunctionCall) goja.Value {
			return i.vm.ToValue(read())
		}),
		nil, goja.FLAG_FALSE, goja.FLAG_TRUE)

	return i
}

func (i *installing) done() error {
	return i.err
}
