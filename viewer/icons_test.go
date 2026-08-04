package viewer_test

import (
	"strings"
	"testing"
	"time"

	"github.com/lrnxzz/go-craft/viewer"
	"github.com/lrnxzz/go-craft/viewer/ultralight"
)

func TestIconStylesheetPaints(t *testing.T) {
	var sheet string
	var opaque int
	var elapsed time.Duration

	on(func() {
		written, err := viewer.IconStylesheet(`<i class="icon icon-diamond_ore">`)
		if err != nil {
			t.Errorf("iconset: %v", err)

			return
		}

		sheet = written

		opened, err := ultralight.Open("ultralight/sdk/resources")
		if err != nil {
			t.Errorf("open: %v", err)

			return
		}

		view := opened.View(64, 64)
		defer view.Close()

		started := time.Now()
		view.LoadHTML(`<style>` + sheet + `</style><body style="margin:0">` +
			`<i class="icon icon-diamond_ore" style="--icon:64px"></i>`)

		for range 800 {
			opened.Update()
			if !view.Loading() {
				break
			}
		}
		elapsed = time.Since(started)

		opened.Render()
		view.Pixels(func(pixels []byte, stride int) {
			for y := range 64 {
				for x := range 64 {
					if pixels[y*stride+x*4+3] > 0 {
						opaque++
					}
				}
			}
		})
	})

	t.Logf("sheet = %d KB, load = %v, opaque pixels = %d/4096", len(sheet)/1024, elapsed, opaque)

	if !strings.Contains(sheet, ".icon-diamond_ore{") {
		t.Fatal("the sheet has no rule for diamond ore")
	}
	if opaque < 500 {
		t.Fatalf("the sprite painted %d pixels, so the atlas never reached the page", opaque)
	}
}
