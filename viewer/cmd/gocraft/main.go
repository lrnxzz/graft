package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net"
	"os"
	"runtime"

	"github.com/lrnxzz/go-craft/agent"
	"github.com/lrnxzz/go-craft/host"
	"github.com/lrnxzz/go-craft/plugin"
	"github.com/lrnxzz/go-craft/viewer"
	"github.com/lrnxzz/go-craft/viewer/gpu"
	"github.com/lrnxzz/go-craft/viewer/ultralight"
)

func init() {
	runtime.LockOSThread()
}

// unreachable is the exit code for a server that is not answering yet, told
// apart from every other failure so a launcher can wait rather than give up, and
// give up rather than wait forever on something that is actually broken
const unreachable = 3

func main() {
	err := run()
	if err == nil {
		return
	}

	slog.Error(err.Error())

	var dial *net.OpError
	if errors.As(err, &dial) {
		os.Exit(unreachable)
	}

	os.Exit(1)
}

func run() error {
	folder := flag.String("plugins", "plugins", "folder scanned for .ts and .js plugins")
	username := flag.String("username", "gocraft", "bot username")
	screenshot := flag.String("screenshot", "", "render a single frame to this PNG and exit")
	diagnose := flag.Bool("check", false, "test the html engine on its own and exit")
	flag.Parse()

	if *diagnose {
		return check()
	}

	address := flag.Arg(0)
	if address == "" {
		return errors.New("usage: gocraft [flags] <host[:port]>")
	}

	play := func(ctx context.Context, bot *agent.Agent) error {
		return show(ctx, bot, *folder, *screenshot)
	}

	return host.Run(context.Background(), address, *username, play)
}

func show(ctx context.Context, robot *agent.Agent, folder, shot string) error {
	plugins, refused := plugin.LoadAll(folder, bot{
		ctx:   ctx,
		agent: robot,
	})
	report("load", refused)
	slog.Info("plugins loaded", "names", plugins.Names())

	view, err := viewer.New(robot, shot == "")
	if err != nil {
		return err
	}
	defer plugins.Teardown()

	// the engine reports most of its failures to its own log and nowhere else,
	// so it gets one before it is asked to do anything
	ultralight.Log("ultralight.log")

	// a plugin that draws with html needs the engine, and one that does not
	// should still run, so a missing engine is a warning rather than the end
	engine, err := ultralight.Open("resources")
	if err != nil {
		slog.Warn("html engine", "err", err)
	}

	opening := menus{
		view:    view,
		plugins: plugins,
		engine:  engine,
	}

	commandsFor(ctx, view, commands{
		plugins: plugins,
		chat:    view.Chat(),
	})

	attach(view, opening)
	report("setup", plugins.Setup(opening))

	if shot != "" {
		return view.Screenshot(shot)
	}

	view.Run(ctx)

	return nil
}

func attach(view *viewer.Viewer, opening menus) {
	plugins := opening.plugins

	view.AddLayer(repaint{engine: opening.engine})
	view.AddLayer(pulse{plugins: plugins})
	view.AddLayer(overlay{plugins: plugins})
	view.AddWorldLayer(marks{plugins: plugins})

	for _, claimed := range plugins.Keys() {
		key, known := keyed(claimed)
		if !known {
			slog.Warn("plugin key", "key", claimed, "err", "the viewer has no such key")

			continue
		}
		if reserved(key) {
			slog.Warn("plugin key", "key", claimed, "err", "taken over from the viewer")
		}

		view.Bind(key, opening.pressing(key))
	}
}

func keyed(name string) (gpu.Key, bool) {
	for _, key := range gpu.Keys {
		if gpu.Name(key) == name {
			return key, true
		}
	}

	return 0, false
}

func reserved(key gpu.Key) bool {
	for _, taken := range [...]gpu.Key{gpu.KeyE, gpu.KeyP, gpu.KeyT, gpu.KeySlash} {
		if key == taken {
			return true
		}
	}

	return false
}

type menus struct {
	view    *viewer.Viewer
	plugins *plugin.Plugins
	engine  *ultralight.Renderer
}

func (m menus) pressing(key gpu.Key) func() {
	return func() {
		menu, opened, err := m.plugins.Press(gpu.Name(key), m)
		if err != nil {
			slog.Warn("plugin key", "key", gpu.Name(key), "err", err)

			return
		}
		if opened {
			m.Open(menu)
		}
	}
}

// a menu says how it wants to be drawn by which of the two it declares. The
// canvas is not a fallback for a page: it can only draw a node tree, and a menu
// written as a page has none, so opening one on it shows an empty screen the
// player still has to dismiss.
func (m menus) Open(menu *plugin.Menu) {
	if !menu.Written() {
		m.view.Open(&screen{menu: menu})

		return
	}
	if m.engine == nil {
		slog.Warn("plugin page", "err", "the html engine never started, so the menu cannot be shown")
		m.Toast("§cthis menu is a page, and the html engine never started")

		return
	}

	markup, err := document(menu)
	if err != nil {
		slog.Warn("plugin page", "err", err)

		return
	}

	m.view.Open(newPaged(m.engine, m.view.Viewport(), menu, markup))
}

func (m menus) Dismiss() {
	m.view.Dismiss()
}

func (m menus) Showing() bool {
	return m.view.Showing()
}

func (m menus) Toast(text string) {
	m.view.Chat().Push(text)
}
