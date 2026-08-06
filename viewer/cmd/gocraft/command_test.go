package main

import (
	"reflect"
	"testing"
)

// A line only stops here when a plugin claimed that word. Everything else has to
// reach the server untouched, or the bot would silently swallow /tp and /gamemode.
func TestOnlyASlashStartsACommand(t *testing.T) {
	for _, line := range []string{"hello there", " ", "", "not/a/command"} {
		_, _, is := split(line)
		if is {
			t.Errorf("%q was taken for a command", line)
		}
	}

	word, arguments, is := split("/mine iron_ore 32")
	if !is {
		t.Fatal("a slashed line was not taken for a command")
	}
	if word != "mine" {
		t.Errorf("word = %q, want mine", word)
	}
	if want := []string{"iron_ore", "32"}; !reflect.DeepEqual(arguments, want) {
		t.Errorf("arguments = %q, want %q", arguments, want)
	}
}

// the word is matched however it was typed, since a plugin declares it once
func TestTheWordIsCaseless(t *testing.T) {
	word, _, _ := split("/MiNe stone")
	if word != "mine" {
		t.Errorf("word = %q, want mine", word)
	}
}

// an argument with a space in it is one argument, so a plugin never has to
// rejoin what the player meant as a single thing
func TestQuotedArgumentsSurviveWhole(t *testing.T) {
	_, arguments, _ := split(`/say "bom dia a todos" agora`)

	want := []string{"bom dia a todos", "agora"}
	if !reflect.DeepEqual(arguments, want) {
		t.Errorf("arguments = %q, want %q", arguments, want)
	}
}

func TestExtraSpacesAreNotArguments(t *testing.T) {
	word, arguments, is := split("   /goto    10   64   -3  ")
	if !is || word != "goto" {
		t.Fatalf("word = %q is = %v", word, is)
	}

	want := []string{"10", "64", "-3"}
	if !reflect.DeepEqual(arguments, want) {
		t.Errorf("arguments = %q, want %q", arguments, want)
	}
}

// a lone slash is what the viewer types to open the chat, and it must not be
// read as a command with an empty name
func TestALoneSlashIsNotACommand(t *testing.T) {
	if _, _, is := split("/"); is {
		t.Error("a lone slash was taken for a command")
	}
}
