package command_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/lrnxzz/graft/viewer/command"
)

// The tree is what a line walks down, so the tests follow lines rather than
// methods: what a player types is the only interface it really has.
var (
	target = command.Spot("target")
	times  = command.Whole("times").Or(1)
	ore    = command.Block("ore")
	saying = command.Rest("message")
)

// walked records what each handler was given, so a test can tell a command that
// ran from one that only parsed
type walked struct {
	ran   string
	times int
	ore   string
}

func tree(seen *walked) *command.Tree {
	goTo := func(c command.Call) error {
		seen.ran = "goto"

		return nil
	}
	mark := func(c command.Call) error {
		seen.ran = "route mark"

		return nil
	}
	follow := func(c command.Call) error {
		seen.ran = "route walk"
		seen.times = times.Of(c)

		return nil
	}
	dig := func(c command.Call) error {
		seen.ran = "mine"
		seen.ore = string(ore.Of(c).Name)

		return nil
	}
	say := func(c command.Call) error {
		seen.ran = saying.Of(c)

		return nil
	}

	return command.Rooted(
		command.Word("goto").Says("walk to a block").Takes(target).Runs(goTo),
		command.Word("mine").Says("mine an ore").Takes(ore).Runs(dig),
		command.Word("say").Says("speak").Takes(saying).Runs(say),
		command.Word("route").Says("draw a path").Under(
			command.Word("mark").Says("mark the block aimed at").Runs(mark),
			command.Word("walk").Says("follow the route").Takes(times).Runs(follow),
		),
	)
}

func TestAWordRunsItsCommand(t *testing.T) {
	var seen walked

	claimed, err := tree(&seen).Run("mine iron_ore", command.Calling(nil, nil))
	if !claimed {
		t.Fatal("a declared word was not claimed")
	}
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if seen.ran != "mine" {
		t.Errorf("ran %q, want mine", seen.ran)
	}
	if seen.ore != "iron_ore" {
		t.Errorf("ore = %q, want iron_ore", seen.ore)
	}
}

// a word nobody declared is not this tree's business, and the caller has to be
// able to tell that from a command that failed
func TestAnUnknownWordIsNotClaimed(t *testing.T) {
	var seen walked

	claimed, err := tree(&seen).Run("teleport 1 2 3", command.Calling(nil, nil))
	if claimed {
		t.Error("an unknown word was claimed")
	}
	if err != nil {
		t.Errorf("an unknown word also failed: %v", err)
	}
}

// the tree is walked, not the string split: route walk is a node under a node
func TestASubcommandIsItsOwnNode(t *testing.T) {
	var seen walked

	if _, err := tree(&seen).Run("route mark", command.Calling(nil, nil)); err != nil {
		t.Fatalf("run: %v", err)
	}
	if seen.ran != "route mark" {
		t.Errorf("ran %q, want route mark", seen.ran)
	}
}

// each child keeps its own arguments, which is the whole reason for the tree
func TestASubcommandKeepsItsOwnArguments(t *testing.T) {
	var seen walked

	if _, err := tree(&seen).Run("route walk 4", command.Calling(nil, nil)); err != nil {
		t.Fatalf("run: %v", err)
	}
	if seen.times != 4 {
		t.Errorf("times = %d, want 4", seen.times)
	}
}

// an optional argument left off falls back rather than failing
func TestAnOptionalArgumentFallsBack(t *testing.T) {
	var seen walked

	if _, err := tree(&seen).Run("route walk", command.Calling(nil, nil)); err != nil {
		t.Fatalf("run: %v", err)
	}
	if seen.times != 1 {
		t.Errorf("times = %d, want the fallback of 1", seen.times)
	}
}

// a parent that only groups must say so rather than pretending to run
func TestAParentWithoutAHandlerSaysWhatItTakes(t *testing.T) {
	var seen walked

	claimed, err := tree(&seen).Run("route", command.Calling(nil, nil))
	if !claimed {
		t.Fatal("a declared word was not claimed")
	}
	if err == nil {
		t.Fatal("a grouping word ran as if it were a command")
	}
	if !strings.Contains(err.Error(), "mark") || !strings.Contains(err.Error(), "walk") {
		t.Errorf("the refusal never named the children: %v", err)
	}
}

// a name the registry never had is refused with the nearest one, because a typo
// among a thousand names is otherwise a guessing game
func TestAMisspelledBlockIsRefusedWithASuggestion(t *testing.T) {
	var seen walked

	_, err := tree(&seen).Run("mine iron_or", command.Calling(nil, nil))
	if err == nil {
		t.Fatal("a name the world never had was accepted")
	}

	var refused *command.Refusal
	if !errors.As(err, &refused) {
		t.Fatalf("the failure was not a refusal: %v", err)
	}
	if refused.Near != "iron_ore" {
		t.Errorf("suggested %q, want iron_ore", refused.Near)
	}
	if seen.ran != "" {
		t.Error("the handler ran on an argument that never parsed")
	}
}

// a failure carries the usage, since the player has no other way to learn it
func TestAFailureCarriesTheUsage(t *testing.T) {
	var seen walked

	_, err := tree(&seen).Run("mine", command.Calling(nil, nil))
	if err == nil {
		t.Fatal("a missing argument was accepted")
	}

	var failed *command.Failure
	if !errors.As(err, &failed) {
		t.Fatalf("the failure carried no usage: %v", err)
	}
	if failed.Usage != "mine <ore>" {
		t.Errorf("usage = %q, want mine <ore>", failed.Usage)
	}
}

// the rest of the line is one argument, spacing and all
func TestTheRestOfTheLineIsOneArgument(t *testing.T) {
	var seen walked

	if _, err := tree(&seen).Run("say hello there   world", command.Calling(nil, nil)); err != nil {
		t.Fatalf("run: %v", err)
	}
	if seen.ran != "hello there   world" {
		t.Errorf("message = %q, want the line as typed", seen.ran)
	}
}

// help walks the whole tree, deepest paths included
func TestHelpNamesEveryLeaf(t *testing.T) {
	var seen walked

	help := strings.Join(tree(&seen).Help(), "\n")

	for _, wanted := range []string{"goto <target>", "mine <ore>", "route mark", "route walk [times]"} {
		if !strings.Contains(help, wanted) {
			t.Errorf("help never mentions %q:\n%s", wanted, help)
		}
	}
}
