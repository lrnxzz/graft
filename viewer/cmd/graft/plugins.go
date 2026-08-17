package main

import (
	"log/slog"
	"sort"
	"strings"

	"github.com/lrnxzz/graft/plugin"
	"github.com/lrnxzz/graft/viewer"
)

// A line the player types is a plugin's before it is the server's, but only if a
// plugin claimed that word. Anything else goes through untouched, so the server's
// own commands keep working and a plugin cannot swallow them by accident.
type commands struct {
	plugins *plugin.Plugins
	chat    *viewer.Chat
}

// A slash also opens a line, since that is the prefix the server's own commands
// already taught the player's fingers.
const (
	prefix = viewer.Opener
	slash  = "/"
)

// Held takes whichever prefix opened the line, and reports the rest. Nothing
// else in the viewer should be testing the first letter itself.
func Held(line string) (string, bool) {
	for _, opener := range []string{prefix, slash} {
		if rest, cut := strings.CutPrefix(line, opener); cut {
			return rest, true
		}
	}

	return "", false
}

func (c commands) claim(line string) bool {
	word, arguments, is := split(line)
	if !is {
		return false
	}

	if word == "help" {
		c.help()

		return true
	}

	ran, err := c.plugins.Run(word, arguments)
	if !ran {
		return false
	}
	if err != nil {
		c.refuse(word, err)
	}

	return true
}

// split takes the leading slash and the words apart. Quoted runs stay together,
// so an argument may hold a space without a plugin having to rejoin it.
func split(line string) (string, []string, bool) {
	held, is := Held(strings.TrimSpace(line))
	if !is {
		return "", nil, false
	}

	words := quoted(held)
	if len(words) == 0 {
		return "", nil, false
	}

	return strings.ToLower(words[0]), words[1:], true
}

func quoted(line string) []string {
	var (
		words []string
		word  strings.Builder
		open  bool
	)

	for _, letter := range line {
		switch {
		case letter == '"':
			open = !open
		case letter == ' ' && !open:
			if word.Len() > 0 {
				words = append(words, word.String())
				word.Reset()
			}
		default:
			word.WriteRune(letter)
		}
	}

	if word.Len() > 0 {
		words = append(words, word.String())
	}

	return words
}

// a command that refused its arguments is answered with what it wanted, since
// the player has no other way to find out
func (c commands) refuse(word string, err error) {
	slog.Warn("plugin command", "command", word, "err", err)

	c.chat.Push(viewer.Red + err.Error())

	usage, known := c.plugins.Usage()[word]
	if known {
		c.chat.Push(viewer.Grey + prefix + usage)
	}
}

func (c commands) help() {
	usage := c.plugins.Usage()
	if len(usage) == 0 {
		c.chat.Push(viewer.Grey + "no plugin has claimed a command")

		return
	}

	words := make([]string, 0, len(usage))
	for word := range usage {
		words = append(words, word)
	}
	sort.Strings(words)

	c.chat.Push(viewer.Grey + "commands from plugins:")
	for _, word := range words {
		c.chat.Push(viewer.White + prefix + usage[word])
	}
}
