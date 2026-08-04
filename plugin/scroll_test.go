package plugin_test

import (
	"testing"

	"github.com/lrnxzz/go-craft/plugin"
)

func openedMenu(t *testing.T) (plugin.Node, *paper) {
	t.Helper()

	runtime := loadExample(t)

	menu, opened, err := runtime.Press("M", nil)
	if err != nil || !opened {
		t.Fatalf("press: opened=%v err=%v", opened, err)
	}

	tree, built, err := menu.Body()
	if err != nil || !built {
		t.Fatalf("menu body: built=%v err=%v", built, err)
	}

	return tree, &paper{width: 800, height: 600}
}

func TestScrollClampsToWhatTheListCanShow(t *testing.T) {
	tree, surface := openedMenu(t)

	var scroll plugin.Scroll
	plugin.Scrolled(tree, surface, &scroll)

	scroll.By(-1000)
	if scroll.Offset <= 0 {
		t.Fatal("a list taller than its box refused to scroll")
	}

	bottom := scroll.Offset
	scroll.By(-1000)
	if scroll.Offset != bottom {
		t.Errorf("offset = %v, want it to stop at %v", scroll.Offset, bottom)
	}

	scroll.By(1000)
	if scroll.Offset != 0 {
		t.Errorf("offset = %v, want it clamped back to the top", scroll.Offset)
	}
}

func TestScrollingMovesTheOptionsUnderTheCursor(t *testing.T) {
	tree, surface := openedMenu(t)

	var scroll plugin.Scroll

	resting := plugin.Scrolled(tree, surface, &scroll)
	if len(resting) == 0 {
		t.Fatal("the list produced nothing clickable")
	}
	_, before, _, _ := resting[0].Bounds()

	scroll.By(-30)
	moved := plugin.Scrolled(tree, surface, &scroll)
	if len(moved) == 0 {
		t.Fatal("scrolling emptied the list")
	}
	_, after, _, _ := moved[0].Bounds()

	if after >= before {
		t.Errorf("first option sat at %v and now at %v, want it to have moved up", before, after)
	}
}

func TestAListThatFitsDoesNotScroll(t *testing.T) {
	var scroll plugin.Scroll

	scroll.By(-1000)
	if scroll.Offset != 0 {
		t.Errorf("offset = %v, want nothing to scroll before a list reports its content", scroll.Offset)
	}
}
