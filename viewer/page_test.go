package viewer_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lrnxzz/go-craft/plugin"
	"github.com/lrnxzz/go-craft/viewer"
	"github.com/lrnxzz/go-craft/viewer/ultralight"
)

// A plugin is written in TypeScript, compiled, run in goja, asked for its markup
// and only then handed to the engine. Everything in that chain can be right on
// its own and still produce a blank panel, so this test walks the whole of it and
// looks at the pixels that came out.
//
// The engine belongs to the thread that opened it, which the window guarantees in
// a real run and the test runner does not.
var engine = make(chan func())

func TestMain(m *testing.M) {
	go pin()

	os.Exit(m.Run())
}

func pin() {
	runtime.LockOSThread()

	for work := range engine {
		work()
	}
}

func on(work func()) {
	done := make(chan struct{})

	engine <- func() {
		defer close(done)

		work()
	}

	<-done
}

// miner is the least a bot can be and still let the ore panel load
type miner struct{}

func (miner) Name() string {
	return "miner"
}

func (miner) Position() plugin.Vec3 {
	return plugin.Vec3{}
}

func (miner) Held() plugin.Item {
	return "iron_pickaxe"
}

func (miner) Inventory() []plugin.Stack {
	return nil
}

func (miner) Say(string) error {
	return nil
}

func (miner) Dig(plugin.Vec3) error {
	return nil
}

func (miner) Goto(plugin.Vec3) (plugin.Vec3, error) {
	return plugin.Vec3{}, nil
}

func (miner) Pursue(plugin.Goal) error {
	return nil
}

func (miner) Abandon() {
}

func (miner) Count(plugin.Item) int {
	return 0
}

func (miner) Hold(plugin.Item) error {
	return nil
}

func markupOf(t *testing.T, folder, key string) string {
	t.Helper()

	plugins, refused := plugin.LoadAll(filepath.Join("..", "plugin", folder), miner{})
	for _, failure := range refused {
		t.Fatalf("load: %v", failure)
	}

	menu, opened, err := plugins.Press(key, nil)
	if err != nil || !opened {
		t.Fatalf("press %s: opened=%v err=%v", key, opened, err)
	}

	written, is, err := menu.Page()
	if err != nil {
		t.Fatalf("page: %v", err)
	}
	if !is {
		t.Fatal("the menu declared no page")
	}

	built, err := viewer.Document(written.Style, written.Body)
	if err != nil {
		t.Fatalf("document: %v", err)
	}

	return built
}

func TestAPluginPageReachesThePixels(t *testing.T) {
	markup := markupOf(t, "example", "B")

	const width, height = 640, 360

	var pixels []byte
	var stride int

	on(func() {
		opened, err := ultralight.Open(filepath.Join("ultralight", "sdk", "resources"))
		if err != nil {
			t.Errorf("open: %v", err)

			return
		}

		view := opened.View(width, height)
		defer view.Close()

		view.LoadHTML(markup)

		for range 400 {
			opened.Update()
			if !view.Loading() {
				break
			}
		}

		opened.Render()
		view.Pixels(func(read []byte, row int) {
			pixels = append([]byte(nil), read...)
			stride = row
		})
	})

	if pixels == nil {
		t.Fatal("the plugin's page rendered nothing")
	}

	// the panel is centred and solid; the corners hold only the scrim that dims
	// the world, so a page that laid out at all differs between the two
	_, _, _, middle := at(pixels, stride, width/2, height/2)
	_, _, _, corner := at(pixels, stride, 4, 4)

	// the panel is glass rather than paint, so it is dense without being opaque
	if middle < 0xc0 {
		t.Fatalf("the centre of the page is %02x, so the panel never laid out", middle)
	}
	if corner > 0xa0 {
		t.Fatalf("the corner is %02x, so the world behind the menu is hidden", corner)
	}
}

func at(pixels []byte, stride, x, y int) (blue, green, red, alpha byte) {
	offset := y*stride + x*4

	return pixels[offset], pixels[offset+1], pixels[offset+2], pixels[offset+3]
}
