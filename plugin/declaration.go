package plugin

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dop251/goja"
)

type Declaration struct {
	Name        string
	Version     string
	Describe    string
	Permissions []Permission

	Reactions []*Reaction
	Commands  map[string]*Command
	Keys      map[string]goja.Callable

	hud   goja.Callable
	world goja.Callable
	setup goja.Callable
	down  goja.Callable
}

type Permission string

const (
	MayMove      Permission = "move"
	MayDig       Permission = "dig"
	MayPlace     Permission = "place"
	MayChat      Permission = "chat"
	MayInventory Permission = "inventory"
	MayUI        Permission = "ui"
)

func (d Declaration) Allows(permission Permission) bool {
	for _, granted := range d.Permissions {
		if granted == permission {
			return true
		}
	}

	return false
}

type Reaction struct {
	Condition goja.Callable
	Act       goja.Callable
	firing    bool
}

func (r *Reaction) Fires() bool {
	held, err := r.Condition(goja.Undefined())
	if err != nil {
		return false
	}

	now := held.ToBoolean()
	edge := now && !r.firing
	r.firing = now

	return edge
}

type Command struct {
	Describe string
	Args     []Argument
	Run      goja.Callable
}

type Argument struct {
	Name     string
	Kind     ArgKind
	Optional bool
}

type ArgKind string

const (
	ArgText     ArgKind = "string"
	ArgNumber   ArgKind = "number"
	ArgBlock    ArgKind = "block"
	ArgItem     ArgKind = "item"
	ArgPlayer   ArgKind = "player"
	ArgPosition ArgKind = "position"
)

func (c Command) Usage(name string) string {
	parts := make([]string, 0, len(c.Args)+1)
	parts = append(parts, name)

	for _, arg := range c.Args {
		if arg.Optional {
			parts = append(parts, "["+arg.Name+"]")

			continue
		}

		parts = append(parts, "<"+arg.Name+">")
	}

	return strings.Join(parts, " ")
}

func (c Command) Parse(vm *goja.Runtime, words []string) (goja.Value, error) {
	parsed := vm.NewObject()
	bridge := into(vm, parsed)

	for index, arg := range c.Args {
		if index >= len(words) {
			if !arg.Optional {
				return nil, fmt.Errorf("missing %s", arg.Name)
			}

			continue
		}

		value, err := arg.coerce(words[index])
		if err != nil {
			return nil, err
		}

		bridge.set(arg.Name, value)
	}

	return parsed, bridge.done()
}

func (a Argument) coerce(word string) (any, error) {
	switch a.Kind {
	case ArgNumber:
		number, err := strconv.ParseFloat(word, 64)
		if err != nil {
			return nil, fmt.Errorf("%s must be a number, got %q", a.Name, word)
		}

		return number, nil
	case ArgPosition:
		spot, err := parsePosition(word)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", a.Name, err)
		}

		return spot, nil
	case ArgText, ArgBlock, ArgItem, ArgPlayer:
		return word, nil
	default:
		return nil, fmt.Errorf("%s declares an unknown kind %q", a.Name, a.Kind)
	}
}

func parsePosition(word string) (Vec3, error) {
	parts := strings.Split(word, ",")
	if len(parts) != 3 {
		return Vec3{}, fmt.Errorf("position must be written as x,y,z, got %q", word)
	}

	var axes [3]float64
	for index, part := range parts {
		value, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
		if err != nil {
			return Vec3{}, fmt.Errorf("invalid coordinate %q", part)
		}

		axes[index] = value
	}

	return Vec3{
		X: axes[0],
		Y: axes[1],
		Z: axes[2],
	}, nil
}

func (r *Runtime) command(call goja.FunctionCall) goja.Value {
	run := r.reading(call.Argument(1)).callable()
	if run == nil {
		panic(r.vm.ToValue("command() needs a function to run"))
	}

	built := &Command{
		Describe: r.reading(call.Argument(2)).text(),
		Run:      run,
	}

	spec := r.reading(call.Argument(0))
	for _, name := range spec.keys() {
		kind := spec.field(name).text()

		built.Args = append(built.Args, Argument{
			Name:     name,
			Kind:     ArgKind(strings.TrimSuffix(kind, "?")),
			Optional: strings.HasSuffix(kind, "?"),
		})
	}

	return r.vm.ToValue(built)
}

func (r *Runtime) when(call goja.FunctionCall) goja.Value {
	condition := r.reading(call.Argument(0)).callable()
	act := r.reading(call.Argument(1)).callable()
	if condition == nil || act == nil {
		panic(r.vm.ToValue("when() needs a condition and an action"))
	}

	return r.vm.ToValue(&Reaction{
		Condition: condition,
		Act:       act,
	})
}
