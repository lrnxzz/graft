package ultralight_test

import (
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/lrnxzz/go-craft/viewer/ultralight"
)

const resources = "sdk/resources"

// The engine belongs to the thread that opened it. A viewer satisfies that for
// free, since the window already pins the main thread, but the test runner hands
// every test a fresh goroutine and moves them between threads at will — which
// crashes roughly one run in seven. So every call into the engine is sent to one
// pinned goroutine, and only the assertions stay on the test's own.
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

// frame is a page rendered to pixels, copied out so the assertions can read it
// after the engine's lock on the bitmap is gone
type frame struct {
	pixels []byte
	stride int
	dirty  ultralight.Rect
}

func (f frame) at(x, y int) (blue, green, red, alpha byte) {
	offset := y*f.stride + x*4

	return f.pixels[offset], f.pixels[offset+1], f.pixels[offset+2], f.pixels[offset+3]
}

func paint(t *testing.T, width, height int, html string) frame {
	t.Helper()

	var painted frame
	var failure error
	var loading bool

	on(func() {
		opened, err := ultralight.Open(resources)
		if err != nil {
			failure = err

			return
		}

		view := opened.View(width, height)
		defer view.Close()

		view.LoadHTML(html)

		for range 400 {
			opened.Update()
			if !view.Loading() {
				break
			}
		}

		loading = view.Loading()
		if loading {
			return
		}

		opened.Render()

		painted.dirty = view.Dirty()
		view.Pixels(func(pixels []byte, stride int) {
			painted.pixels = append([]byte(nil), pixels...)
			painted.stride = stride
		})
	})

	if failure != nil {
		t.Fatalf("open: %v", failure)
	}
	if loading {
		t.Fatal("the page never finished loading")
	}
	if painted.pixels == nil {
		t.Fatal("the page rendered no pixels")
	}

	return painted
}

func TestVersion(t *testing.T) {
	if !strings.HasPrefix(ultralight.Version(), "1.") {
		t.Fatalf("unexpected engine version %q", ultralight.Version())
	}
}

// a solid page proves CSS reached the framebuffer, and proves the channel order
// the GL upload will have to respect
func TestPaintsCSS(t *testing.T) {
	blue, green, red, alpha := paint(t, 64, 64, `<body style="margin:0;background:#3366cc">`).at(32, 32)

	if red != 0x33 || green != 0x66 || blue != 0xcc || alpha != 0xff {
		t.Fatalf("centre pixel is %02x%02x%02x%02x, want 3366ccff in BGRA order", blue, green, red, alpha)
	}
}

// transparency is what lets the page sit over the world instead of hiding it
func TestKeepsBackgroundTransparent(t *testing.T) {
	_, _, _, alpha := paint(t, 64, 64, `<body style="margin:0">`).at(32, 32)

	if alpha != 0 {
		t.Fatalf("an empty page painted alpha %d, want 0", alpha)
	}
}

// the colour arrives already multiplied by its alpha, which is what forces the
// blend to be ONE rather than SRC_ALPHA; getting it wrong leaves dark fringes
// around every translucent panel
func TestPaintsPremultiplied(t *testing.T) {
	_, _, red, alpha := paint(t, 64, 64, `<body style="margin:0;background:rgba(255,0,0,0.5)">`).at(32, 32)

	if red > alpha {
		t.Fatalf("red %02x exceeds alpha %02x, so the colour is not premultiplied", red, alpha)
	}
}

// the gradient is the whole reason for the engine: it is what the painter it
// replaces could not draw at all
func TestPaintsGradient(t *testing.T) {
	painted := paint(t, 64, 64, `<body style="margin:0;background:linear-gradient(to right,#000,#fff)">`)

	left, _, _, _ := painted.at(4, 32)
	right, _, _, _ := painted.at(59, 32)

	if right <= left+0x40 {
		t.Fatalf("gradient went from %02x to %02x, which is not a ramp", left, right)
	}
}

// the dirty region is what keeps a frame from uploading the whole page, so it
// has to report a change and then forget it once cleared
func TestReportsDirtyRegion(t *testing.T) {
	drawn := paint(t, 128, 128, `<body style="margin:0;background:#000">`).dirty
	if drawn.Empty() {
		t.Fatal("a freshly painted page reported nothing dirty")
	}
	if drawn.Width() > 128 || drawn.Height() > 128 {
		t.Fatalf("dirty region %v is larger than the view", drawn)
	}

	var stale ultralight.Rect

	on(func() {
		opened, err := ultralight.Open(resources)
		if err != nil {
			return
		}

		view := opened.View(64, 64)
		defer view.Close()

		view.LoadHTML(`<body style="margin:0;background:#000">`)

		for range 400 {
			opened.Update()
			if !view.Loading() {
				break
			}
		}

		opened.Render()
		view.Clean()

		stale = view.Dirty()
	})

	if !stale.Empty() {
		t.Fatalf("region stayed dirty after Clean: %v", stale)
	}
}

// one engine has to serve many menus opening and closing over a session
func TestServesManyViews(t *testing.T) {
	greens := make([]byte, 0, 4)

	on(func() {
		opened, err := ultralight.Open(resources)
		if err != nil {
			return
		}

		for range 4 {
			view := opened.View(64, 64)

			view.LoadHTML(`<body style="margin:0;background:#0f0">`)

			for range 400 {
				opened.Update()
				if !view.Loading() {
					break
				}
			}

			opened.Render()
			view.Pixels(func(pixels []byte, stride int) {
				greens = append(greens, pixels[32*stride+32*4+1])
			})

			view.Close()
		}
	})

	if len(greens) != 4 {
		t.Fatalf("painted %d views, want 4", len(greens))
	}
	for round, green := range greens {
		if green != 0xff {
			t.Fatalf("round %d painted green %02x, want ff", round, green)
		}
	}
}

// reopening elsewhere has to be refused rather than silently ignored, since the
// caller would otherwise be reading resources it never asked for
func TestRefusesASecondResourcesDirectory(t *testing.T) {
	var moved *ultralight.Renderer
	var err error

	on(func() {
		_, err = ultralight.Open(resources)
		if err != nil {
			return
		}

		moved, err = ultralight.Open("sdk/somewhere-else")
	})

	if err == nil {
		t.Fatal("opening a second resources directory was allowed")
	}
	if moved != nil {
		t.Fatal("a refused open still handed back a renderer")
	}
}

// react loads a page, delivers input to it, and paints the result, so a test can
// tell whether the event reached the document rather than merely being accepted
func react(t *testing.T, html string, deliver func(*ultralight.View)) frame {
	t.Helper()

	var painted frame
	var failure error

	on(func() {
		opened, err := ultralight.Open(resources)
		if err != nil {
			failure = err

			return
		}

		view := opened.View(64, 64)
		defer view.Close()

		view.LoadHTML(html)

		for range 400 {
			opened.Update()
			if !view.Loading() {
				break
			}
		}

		opened.Render()
		view.Focus(true)

		deliver(view)

		for range 20 {
			opened.Update()
		}

		opened.Render()

		view.Pixels(func(pixels []byte, stride int) {
			painted.pixels = append([]byte(nil), pixels...)
			painted.stride = stride
		})
	})

	if failure != nil {
		t.Fatalf("open: %v", failure)
	}
	if painted.pixels == nil {
		t.Fatal("the page rendered no pixels")
	}

	return painted
}

const clickable = `<body style="margin:0;background:#f00"
	onmousedown="document.body.style.background='#0f0'">`

func TestDeliversClicks(t *testing.T) {
	pressed := func(view *ultralight.View) {
		view.Moved(32, 32)
		view.Pressed(32, 32, ultralight.Left)
		view.Released(32, 32, ultralight.Left)
	}

	_, green, red, _ := react(t, clickable, pressed).at(32, 32)

	if green != 0xff || red != 0x00 {
		t.Fatalf("after a click the page is r=%02x g=%02x, want r=00 g=ff", red, green)
	}
}

// a click that misses must not fire, or every press anywhere would count as a
// press on whatever the page happens to hold
func TestIgnoresClicksOutsideTheTarget(t *testing.T) {
	const spotted = `<body style="margin:0;background:#f00"><div style="width:8px;height:8px"
		onmousedown="document.body.style.background='#0f0'"></div>`

	pressed := func(view *ultralight.View) {
		view.Moved(50, 50)
		view.Pressed(50, 50, ultralight.Left)
		view.Released(50, 50, ultralight.Left)
	}

	_, green, red, _ := react(t, spotted, pressed).at(32, 32)

	if red != 0xff || green != 0x00 {
		t.Fatalf("a click on empty space still fired: r=%02x g=%02x", red, green)
	}
}

// the wheel is what the old menu could never get right, so it is worth pinning
// that the page itself now moves
func TestDeliversScroll(t *testing.T) {
	// the background of a body propagates to the canvas and stays put, so the
	// bands that are meant to move live in the flow instead
	const bands = `<div style="height:64px;background:#000"></div>
		<div style="height:400px;background:#fff"></div>`

	const scrollable = `<body style="margin:0">
		<div style="height:64px;overflow-y:scroll">` + bands + `</div>`

	wheeled := func(view *ultralight.View) {
		view.Moved(32, 32)
		view.Scrolled(0, -200)
	}

	blue, _, _, _ := react(t, scrollable, wheeled).at(32, 32)

	if blue < 0x80 {
		t.Fatalf("after scrolling down the centre is still %02x, so the box never moved", blue)
	}
}

// The wheel moves an overflowing element and never the document itself, which
// decides how a menu has to be written: content that should scroll needs a box
// around it, not just a page taller than the window.
func TestLeavesTheDocumentUnscrolled(t *testing.T) {
	const tall = `<body style="margin:0">
		<div style="height:64px;background:#000"></div>
		<div style="height:400px;background:#fff"></div>`

	wheeled := func(view *ultralight.View) {
		view.Moved(32, 32)
		view.Scrolled(0, -200)
	}

	blue, _, _, _ := react(t, tall, wheeled).at(32, 32)

	if blue >= 0x80 {
		t.Fatal("the document scrolled, so a menu no longer needs a box around its content")
	}
}

// text arrives already composed by the window, which is the only path that
// spells accents and non-latin layouts correctly
func TestDeliversTypedText(t *testing.T) {
	const listening = `<body style="margin:0;background:#f00"
		onkeypress="document.body.style.background='#0f0'">`

	typed := func(view *ultralight.View) {
		view.Typed("a", 0)
	}

	_, green, red, _ := react(t, listening, typed).at(32, 32)

	if green != 0xff || red != 0x00 {
		t.Fatalf("after typing the page is r=%02x g=%02x, want r=00 g=ff", red, green)
	}
}

func TestResizeTracksSize(t *testing.T) {
	var width, height int

	on(func() {
		opened, err := ultralight.Open(resources)
		if err != nil {
			return
		}

		view := opened.View(100, 50)
		defer view.Close()

		view.Resize(200, 80)

		width, height = view.Size()
	})

	if width != 200 || height != 80 {
		t.Fatalf("size is %dx%d after resize, want 200x80", width, height)
	}
}

// the page reaching back into Go is what lets a menu do more than look right;
// without it a click is only a repaint
func TestPageSendsToGo(t *testing.T) {
	const talking = `<body style="margin:0" onload="gocraft.send('pick','stone',7)">`

	var heard []string
	var name string

	on(func() {
		opened, err := ultralight.Open(resources)
		if err != nil {
			return
		}

		view := opened.View(32, 32)
		defer view.Close()

		view.Sends(func(sent string, args []string) {
			name = sent
			heard = args
		})

		view.LoadHTML(talking)

		for range 400 {
			opened.Update()
			if !view.Loading() {
				break
			}
		}
	})

	if name != "pick" {
		t.Fatalf("the page sent %q, want pick", name)
	}
	if len(heard) != 2 || heard[0] != "stone" || heard[1] != "7" {
		t.Fatalf("arguments arrived as %q, want [stone 7]", heard)
	}
}

// a view nobody listens to must not grow the function, or a plugin without the
// permission would still be reachable from its own markup
func TestPageWithoutAListenerHasNoBridge(t *testing.T) {
	const asking = `<body style="margin:0" onload="document.title = typeof gocraft">`

	var kind string

	on(func() {
		opened, err := ultralight.Open(resources)
		if err != nil {
			return
		}

		view := opened.View(32, 32)
		defer view.Close()

		view.LoadHTML(asking)

		for range 400 {
			opened.Update()
			if !view.Loading() {
				break
			}
		}

		kind, _ = view.Eval("document.title")
	})

	if kind != "undefined" {
		t.Fatalf("an unlistened page saw gocraft as %q, want undefined", kind)
	}
}

func TestEvalReadsThePage(t *testing.T) {
	var answer string
	var failure error

	on(func() {
		opened, err := ultralight.Open(resources)
		if err != nil {
			return
		}

		view := opened.View(32, 32)
		defer view.Close()

		view.LoadHTML(`<body><span id="n">41</span>`)

		for range 400 {
			opened.Update()
			if !view.Loading() {
				break
			}
		}

		answer, failure = view.Eval("String(Number(document.getElementById('n').textContent) + 1)")
	})

	if failure != nil {
		t.Fatalf("eval: %v", failure)
	}
	if answer != "42" {
		t.Fatalf("eval answered %q, want 42", answer)
	}
}
