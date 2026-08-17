package command_test

import (
	"strings"
	"testing"

	"github.com/lrnxzz/graft/viewer/command"
)

// Completion runs the same walk and the same parameters as a real call, so these
// check the offers a player would see at each point of a line rather than the
// pieces underneath.
func offered(t *testing.T, line string) command.Offer {
	t.Helper()

	var seen walked

	return tree(&seen).Complete(line, len(line), command.Calling(nil, nil))
}

func TestAnEmptyLineOffersEveryCommand(t *testing.T) {
	offer := offered(t, "")

	for _, wanted := range []string{"goto", "mine", "route", "say"} {
		if !has(offer.Words, wanted) {
			t.Errorf("%q was never offered, got %v", wanted, offer.Words)
		}
	}
}

// a half-typed word narrows to what it could still become
func TestAHalfTypedCommandNarrows(t *testing.T) {
	offer := offered(t, "ro")

	if len(offer.Words) != 1 || offer.Words[0] != "route" {
		t.Errorf("offered %v, want just route", offer.Words)
	}
}

// after a grouping word the children are what may come next
func TestAGroupOffersItsChildren(t *testing.T) {
	offer := offered(t, "route ")

	if !has(offer.Words, "mark") || !has(offer.Words, "walk") {
		t.Errorf("offered %v, want the children of route", offer.Words)
	}
}

// the child's own arguments are offered once the child is named, which is what
// the tree buys over one command with a choice argument
func TestAChildOffersItsOwnArgument(t *testing.T) {
	offer := offered(t, "mine ")

	if len(offer.Words) == 0 {
		t.Fatal("a block argument offered nothing")
	}
	for _, word := range offer.Words {
		if strings.Contains(word, ":") {
			t.Errorf("offered %q with a namespace the player never types", word)
		}
	}
}

// a half-typed block name narrows against the registry
func TestAHalfTypedBlockNarrows(t *testing.T) {
	offer := offered(t, "mine iron_o")

	if !has(offer.Words, "iron_ore") {
		t.Errorf("offered %v, want iron_ore among them", offer.Words)
	}
	for _, word := range offer.Words {
		if !strings.HasPrefix(word, "iron_o") {
			t.Errorf("offered %q, which cannot follow what was typed", word)
		}
	}
}

// the span is what the chat replaces, so it has to cover the word being typed
// and nothing more
func TestTheOfferReplacesTheWordBeingTyped(t *testing.T) {
	line := "mine iron_o"
	offer := offered(t, line)

	if line[offer.From:offer.To] != "iron_o" {
		t.Errorf("the offer replaces %q, want iron_o", line[offer.From:offer.To])
	}
}

// a cursor sitting after a space starts the next argument, and must not be read
// as still editing the last one
func TestASpaceStartsTheNextArgument(t *testing.T) {
	var seen walked

	line := "route walk "
	offer := tree(&seen).Complete(line, len(line), command.Calling(nil, nil))

	if offer.From != len(line) || offer.To != len(line) {
		t.Errorf("the offer spans %d..%d, want the empty spot at %d", offer.From, offer.To, len(line))
	}
}

// a word nobody declared offers nothing rather than guessing
func TestAnUnknownCommandOffersNothing(t *testing.T) {
	if offer := offered(t, "teleport "); !offer.Empty() {
		t.Errorf("an unknown command offered %v", offer.Words)
	}
}

// the argument after a filled one is the next parameter, not the same one again
func TestCompletionCountsWhatWasAlreadyGiven(t *testing.T) {
	var seen walked

	line := "route walk 3"
	offer := tree(&seen).Complete(line, len(line), command.Calling(nil, nil))

	// times is the only argument route walk takes, and it is already written
	if !offer.Empty() {
		t.Errorf("offered %v after every argument was given", offer.Words)
	}
}

func has(among []string, wanted string) bool {
	for _, word := range among {
		if word == wanted {
			return true
		}
	}

	return false
}
