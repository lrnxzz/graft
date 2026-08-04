package plugin

import (
	"errors"
	"fmt"

	"github.com/dop251/goja"
)

var (
	errLooseReaction = errors.New("a reaction was not built with when()")
	errLooseCommands = errors.New("commands must be an object")
	errLooseKeys     = errors.New("keys must be an object")
	errNotAPlugin    = errors.New("the default export is not a plugin object")
)

type Runtime struct {
	vm     *goja.Runtime
	source Source
	decl   Declaration
}

func Load(source Source, bot Bot) (*Runtime, error) {
	vm := goja.New()
	vm.SetFieldNameMapper(goja.UncapFieldNameMapper())

	runtime := &Runtime{
		vm:     vm,
		source: source,
	}
	if err := runtime.install(); err != nil {
		return nil, err
	}

	declared, err := runtime.run()
	if err != nil {
		return nil, err
	}

	decl, err := runtime.read(declared)
	if err != nil {
		return nil, fmt.Errorf("plugin %s: %w", source.Name, err)
	}
	runtime.decl = decl

	// the bot goes in last, once the declaration has been read: an ability the
	// plugin never asked permission for is then never bound at all, rather than
	// bound and checked
	if err := runtime.attach(bot); err != nil {
		return nil, fmt.Errorf("plugin %s: %w", source.Name, err)
	}

	return runtime, nil
}

// run gives the script the commonjs module esbuild compiles against, and hands
// back whatever it exported as default.
//
// The exports object is read after the script and not before: the wrapper esbuild
// emits reassigns module.exports on its own last line, so anything captured up
// front is a stale object that never sees the plugin.
func (r *Runtime) run() (goja.Value, error) {
	exports := r.vm.NewObject()
	module := r.vm.NewObject()

	exporting := map[string]any{"exports": exports}
	commonjs := map[string]any{
		"module":  module,
		"exports": exports,
	}

	if err := into(r.vm, module).all(exporting).done(); err != nil {
		return nil, err
	}
	if err := into(r.vm, r.vm.GlobalObject()).all(commonjs).done(); err != nil {
		return nil, err
	}

	if _, err := r.vm.RunScript(r.source.Path, r.source.Code); err != nil {
		return nil, fmt.Errorf("plugin %s: %w", r.source.Name, err)
	}

	settled, is := module.Get("exports").(*goja.Object)
	if !is {
		return nil, fmt.Errorf("plugin %s: exported nothing", r.source.Name)
	}

	declared := settled.Get("default")
	if declared == nil || goja.IsUndefined(declared) {
		return nil, fmt.Errorf("plugin %s: no default export", r.source.Name)
	}

	return declared, nil
}

func (r *Runtime) attach(bot Bot) error {
	object, err := r.botObject(bot, r.decl.Permissions)
	if err != nil {
		return err
	}

	return r.vm.Set("bot", object)
}

func (r *Runtime) Exposed() map[string]bool {
	surface := map[string]bool{}
	for _, name := range r.reading(r.vm.Get("bot")).keys() {
		surface[name] = true
	}

	return surface
}

func (r *Runtime) Declaration() Declaration {
	return r.decl
}

func (r *Runtime) Name() string {
	if r.decl.Name != "" {
		return r.decl.Name
	}

	return r.source.Name
}

func (r *Runtime) read(declared goja.Value) (Declaration, error) {
	spec := r.reading(declared)
	if spec.object() == nil {
		return Declaration{}, errNotAPlugin
	}

	decl := Declaration{
		Name:     spec.field("name").text(),
		Version:  spec.field("version").text(),
		Describe: spec.field("describe").text(),
		Commands: map[string]*Command{},
		Keys:     map[string]goja.Callable{},

		hud:   spec.field("hud").callable(),
		world: spec.field("world").callable(),
		setup: spec.field("setup").callable(),
		down:  spec.field("teardown").callable(),
	}

	for _, granted := range spec.field("permissions").items() {
		decl.Permissions = append(decl.Permissions, Permission(granted.text()))
	}
	for _, reacting := range spec.field("reactions").items() {
		reaction, built := exported[*Reaction](reacting)
		if !built {
			return Declaration{}, errLooseReaction
		}

		decl.Reactions = append(decl.Reactions, reaction)
	}

	if err := readCommands(spec, &decl); err != nil {
		return Declaration{}, err
	}
	if err := readKeys(spec, &decl); err != nil {
		return Declaration{}, err
	}

	if decl.Name == "" {
		decl.Name = r.source.Name
	}

	return decl, nil
}

func readCommands(spec reading, decl *Declaration) error {
	commands := spec.field("commands")
	if commands.missing() {
		return nil
	}
	if commands.object() == nil {
		return errLooseCommands
	}

	for _, word := range commands.keys() {
		command, built := exported[*Command](commands.field(word))
		if !built {
			return fmt.Errorf("command %q was not built with command()", word)
		}

		decl.Commands[word] = command
	}

	return nil
}

func readKeys(spec reading, decl *Declaration) error {
	keys := spec.field("keys")
	if keys.missing() {
		return nil
	}
	if keys.object() == nil {
		return errLooseKeys
	}

	for _, name := range keys.keys() {
		bound := keys.field(name).callable()
		if bound == nil {
			return fmt.Errorf("key %q is not bound to a function", name)
		}

		decl.Keys[name] = bound
	}

	return nil
}

func (r *Runtime) Setup(ui UI) error {
	if r.decl.setup == nil {
		return nil
	}

	_, err := r.decl.setup(goja.Undefined(), r.vm.Get("bot"), r.uiObject(ui))

	return err
}

func (r *Runtime) Teardown() {
	if r.decl.down == nil {
		return
	}

	_, _ = r.decl.down(goja.Undefined())
}

func (r *Runtime) Tick() []error {
	var failures []error
	for _, reaction := range r.decl.Reactions {
		if !reaction.Fires() {
			continue
		}
		if _, err := reaction.Act(goja.Undefined(), r.vm.Get("bot")); err != nil {
			failures = append(failures, err)
		}
	}

	return failures
}

func (r *Runtime) Hud() (Node, bool, error) {
	if r.decl.hud == nil {
		return nil, false, nil
	}

	built, err := r.decl.hud(goja.Undefined(), r.vm.Get("bot"))
	if err != nil {
		return nil, false, err
	}

	return rootOf(r.reading(built))
}

func (r *Runtime) World() ([]Marker, error) {
	if r.decl.world == nil {
		return nil, nil
	}

	built, err := r.decl.world(goja.Undefined(), r.vm.Get("bot"))
	if err != nil {
		return nil, err
	}

	var marks []Marker
	for _, item := range r.reading(built).items() {
		marker, is := exported[Marker](item)
		if is {
			marks = append(marks, marker)
		}
	}

	return marks, nil
}

func (r *Runtime) Parse(command *Command, words []string) (goja.Value, error) {
	return command.Parse(r.vm, words)
}

func (r *Runtime) Run(command *Command, words []string) error {
	parsed, err := r.Parse(command, words)
	if err != nil {
		return err
	}

	_, err = command.Run(goja.Undefined(), r.vm.Get("bot"), parsed)

	return err
}

func (r *Runtime) Press(key string, ui UI) (*Menu, bool, error) {
	bound, claimed := r.decl.Keys[key]
	if !claimed {
		return nil, false, nil
	}

	opened := &capturing{inner: ui}
	if _, err := bound(goja.Undefined(), r.vm.Get("bot"), r.uiObject(opened)); err != nil {
		return nil, false, err
	}

	return opened.menu, opened.menu != nil, nil
}
