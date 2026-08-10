package viewer

import (
	"testing"
)

func offering(words ...string) *Suggesting {
	list := &Suggesting{}
	list.Offers(words, "", 0, 0)

	return list
}

// Enter has two jobs, and which one it does turns on whether the player picked a
// row. Once they have, the keystroke belongs to the list.
func TestEnterTakesTheRowOncePicked(t *testing.T) {
	list := offering("walk", "wait")

	if list.Moved() {
		t.Error("the list claims a row was picked before any key was pressed")
	}

	list.Move(1)

	if !list.Moved() {
		t.Error("the list forgot the row was picked")
	}
	if word, _, _, _ := list.Chosen(); word != "wait" {
		t.Errorf("picked %q, want wait", word)
	}
}

// a word already typed in full buys nothing, so the keystroke has to go to the
// line instead of standing still
func TestAWordAlreadyWrittenAddsNothing(t *testing.T) {
	list := &Suggesting{}
	list.Offers([]string{"list"}, "list", 0, 4)

	if list.Adds() {
		t.Error("a word typed in full still claimed the keystroke")
	}

	list.Offers([]string{"list"}, "li", 0, 2)

	if !list.Adds() {
		t.Error("a half-typed word gave up the keystroke")
	}
}

// narrowing the list starts over, since the row that was picked may not even be
// among the words any more
func TestNarrowingForgetsThePick(t *testing.T) {
	list := offering("walk", "wait", "wander")
	list.Move(2)

	list.Offers([]string{"walk"}, "wal", 0, 3)

	if list.Moved() {
		t.Error("the pick survived the list changing under it")
	}
	if word, _, _, _ := list.Chosen(); word != "walk" {
		t.Errorf("chose %q after narrowing, want the first word", word)
	}
}

// typing a letter that narrows nothing must not throw the pick away
func TestAnUnchangedListKeepsThePick(t *testing.T) {
	list := offering("walk", "wait")
	list.Move(1)

	list.Offers([]string{"walk", "wait"}, "wa", 0, 2)

	if word, _, _, _ := list.Chosen(); word != "wait" {
		t.Errorf("chose %q, want the pick to survive", word)
	}
}

// the panel holds eight rows, so a longer list has to follow the pick down
// rather than highlight something nobody can see
func TestTheWindowFollowsThePick(t *testing.T) {
	words := make([]string, 0, 12)
	for at := range 12 {
		words = append(words, string(rune('a'+at)))
	}

	list := offering(words...)

	if shown := list.window(); shown[0] != "a" || len(shown) != suggestRows {
		t.Fatalf("the window starts at %q with %d rows", shown[0], len(shown))
	}
	if list.more() != 4 {
		t.Errorf("counted %d words past the window, want 4", list.more())
	}

	for range suggestRows {
		list.Move(1)
	}

	shown := list.window()
	if shown[len(shown)-1] != "i" {
		t.Errorf("the window ends at %q, want the picked row", shown[len(shown)-1])
	}
	if list.more() != 3 {
		t.Errorf("counted %d words past the window, want 3", list.more())
	}

	// coming back up inside the window must not drag it along, or the list would
	// slide under a pick that was already on screen
	list.Move(-1)
	list.Move(-1)

	if shown := list.window(); shown[0] != "b" {
		t.Errorf("moving back up inside the window slid it to %q", shown[0])
	}
	if word, _, _, _ := list.Chosen(); word != "g" {
		t.Errorf("chose %q after moving back up, want g", word)
	}
}

// wrapping past the end lands on the first row with the window back at the top
func TestWrappingBringsTheWindowBack(t *testing.T) {
	list := offering("a", "b", "c", "d", "e", "f", "g", "h", "i", "j")

	list.Move(-1)

	if word, _, _, _ := list.Chosen(); word != "j" {
		t.Errorf("moving up from the first row chose %q, want the last", word)
	}
	if shown := list.window(); shown[len(shown)-1] != "j" {
		t.Errorf("the window ends at %q, want the last row", shown[len(shown)-1])
	}

	list.Move(1)

	if shown := list.window(); shown[0] != "a" {
		t.Errorf("wrapping left the window at %q, want the top", shown[0])
	}
}
