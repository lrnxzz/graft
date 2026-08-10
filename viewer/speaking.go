package viewer

import (
	"fmt"
	"strings"
)

// Minecraft's own colour codes, which the chat already renders. Naming them keeps
// the escape out of every call site.
const (
	Grey   = "§7"
	Faint  = "§8"
	White  = "§f"
	Green  = "§a"
	Red    = "§c"
	Yellow = "§e"
	Blue   = "§b"
	Bold   = "§l"
)

// Speaking writes to the chat in shapes rather than lines: a heading, a row, a
// bar. A command that reports through these reads the same as every other, and
// nothing has to remember which colour meant what.
type Speaking struct {
	chat *Chat
}

func Says(chat *Chat) Speaking {
	return Speaking{
		chat: chat,
	}
}

func (s Speaking) Line(text string) {
	s.chat.Says(text)
}

// Head is a title with a rule under it, for the top of a list
func (s Speaking) Head(title string) {
	s.chat.Says(Grey + Bold + strings.ToUpper(title))
}

// Note is something that went as expected
func (s Speaking) Note(text string) {
	s.chat.Says(Grey + text)
}

func (s Speaking) Good(text string) {
	s.chat.Says(Green + text)
}

func (s Speaking) Warn(text string) {
	s.chat.Says(Yellow + text)
}

// Bad is a refusal, and it is the only shape that shows a caret, because that is
// the only time the player needs to see where in their line it went wrong
func (s Speaking) Bad(text string) {
	s.chat.Says(Red + text)
}

// Under draws the offending run of a line with a marker beneath it
func (s Speaking) Under(line string, from, to int) {
	if to <= from {
		to = from + 1
	}
	if from > len(line) {
		return
	}
	if to > len(line) {
		to = len(line)
	}

	s.chat.Says(Faint + line)
	s.chat.Says(Red + strings.Repeat(" ", from) + strings.Repeat("^", to-from))
}

// Row is a numbered entry with a trailing note, which is what a list of things
// with a state looks like
func (s Speaking) Row(index int, text, note string) {
	s.chat.Says(fmt.Sprintf("%s%2d. %s%s  %s", Faint, index, White, text, note))
}

// Pair is a name and its value, lined up by the name being dim
func (s Speaking) Pair(name, value string) {
	s.chat.Says(Grey + name + "  " + White + value)
}

const barCells = 20

// Bar is a filled run, for a count against a total that a number alone would not
// show at a glance
func (s Speaking) Bar(name string, done, total int) {
	if total <= 0 {
		return
	}

	filled := done * barCells / total
	if filled > barCells {
		filled = barCells
	}

	drawn := Green + strings.Repeat("|", filled) + Faint + strings.Repeat("|", barCells-filled)

	s.chat.Says(fmt.Sprintf("%s%s %s %s%d/%d", Grey, name, drawn, White, done, total))
}

// Keys is the hint line at the foot of a mode, naming what each key does
func (s Speaking) Keys(pairs ...string) {
	if len(pairs)%2 != 0 {
		return
	}

	said := strings.Builder{}
	for at := 0; at < len(pairs); at += 2 {
		if at > 0 {
			said.WriteString(Faint + " · ")
		}

		said.WriteString(Yellow + pairs[at] + Grey + " " + pairs[at+1])
	}

	s.chat.Says(said.String())
}
