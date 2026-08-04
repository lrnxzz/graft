package plugin

import "fmt"

type Failure struct {
	Plugin string
	Err    error
}

func (f Failure) Error() string {
	return "plugin " + f.Plugin + ": " + f.Err.Error()
}

func (f Failure) Unwrap() error {
	return f.Err
}

type Plugins struct {
	loaded   []*Runtime
	commands map[string]claim
	keys     map[string]*Runtime
}

type claim struct {
	runtime *Runtime
	command *Command
}

func LoadAll(dir string, bot Bot) (*Plugins, []Failure) {
	plugins := &Plugins{
		commands: map[string]claim{},
		keys:     map[string]*Runtime{},
	}

	sources, err := Collect(dir)
	if err != nil {
		return plugins, []Failure{{Plugin: dir, Err: err}}
	}

	var refused []Failure
	for _, source := range sources {
		runtime, err := Load(source, bot)
		if err != nil {
			refused = append(refused, Failure{
				Plugin: source.Name,
				Err:    err,
			})

			continue
		}

		refused = append(refused, plugins.adopt(runtime)...)
	}

	return plugins, refused
}

func (p *Plugins) adopt(runtime *Runtime) []Failure {
	var refused []Failure

	for word, command := range runtime.Declaration().Commands {
		taken, claimed := p.commands[word]
		if claimed {
			refused = append(refused, Failure{
				Plugin: runtime.Name(),
				Err:    fmt.Errorf("the command %q already belongs to %s", word, taken.runtime.Name()),
			})

			continue
		}

		p.commands[word] = claim{
			runtime: runtime,
			command: command,
		}
	}

	for key := range runtime.Declaration().Keys {
		taken, claimed := p.keys[key]
		if claimed {
			refused = append(refused, Failure{
				Plugin: runtime.Name(),
				Err:    fmt.Errorf("the key %q already belongs to %s", key, taken.Name()),
			})

			continue
		}

		p.keys[key] = runtime
	}

	p.loaded = append(p.loaded, runtime)

	return refused
}

func (p *Plugins) Names() []string {
	names := make([]string, 0, len(p.loaded))
	for _, runtime := range p.loaded {
		names = append(names, runtime.Name())
	}

	return names
}

func (p *Plugins) Keys() []string {
	claimed := make([]string, 0, len(p.keys))
	for key := range p.keys {
		claimed = append(claimed, key)
	}

	return claimed
}

func (p *Plugins) Setup(ui UI) []Failure {
	var failures []Failure
	for _, runtime := range p.loaded {
		if err := runtime.Setup(ui); err != nil {
			failures = append(failures, Failure{
				Plugin: runtime.Name(),
				Err:    err,
			})
		}
	}

	return failures
}

func (p *Plugins) Tick() []Failure {
	var failures []Failure
	for _, runtime := range p.loaded {
		for _, err := range runtime.Tick() {
			failures = append(failures, Failure{
				Plugin: runtime.Name(),
				Err:    err,
			})
		}
	}

	return failures
}

func (p *Plugins) Hud() ([]Node, []Failure) {
	var (
		roots    []Node
		failures []Failure
	)

	for _, runtime := range p.loaded {
		root, declared, err := runtime.Hud()
		if err != nil {
			failures = append(failures, Failure{
				Plugin: runtime.Name(),
				Err:    err,
			})

			continue
		}
		if declared {
			roots = append(roots, root)
		}
	}

	return roots, failures
}

func (p *Plugins) World() ([]Marker, []Failure) {
	var (
		marks    []Marker
		failures []Failure
	)

	for _, runtime := range p.loaded {
		drawn, err := runtime.World()
		if err != nil {
			failures = append(failures, Failure{
				Plugin: runtime.Name(),
				Err:    err,
			})

			continue
		}

		marks = append(marks, drawn...)
	}

	return marks, failures
}

func (p *Plugins) Run(word string, words []string) (bool, error) {
	claimed, known := p.commands[word]
	if !known {
		return false, nil
	}

	return true, claimed.runtime.Run(claimed.command, words)
}

func (p *Plugins) Usage() map[string]string {
	usage := make(map[string]string, len(p.commands))
	for word, claimed := range p.commands {
		usage[word] = claimed.command.Usage(word)
	}

	return usage
}

func (p *Plugins) Press(key string, ui UI) (*Menu, bool, error) {
	runtime, claimed := p.keys[key]
	if !claimed {
		return nil, false, nil
	}

	return runtime.Press(key, ui)
}

func (p *Plugins) Teardown() {
	for _, runtime := range p.loaded {
		runtime.Teardown()
	}
}
