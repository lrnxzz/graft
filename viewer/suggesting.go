package viewer

import (
	"slices"
	"strconv"
	"strings"

	"github.com/lrnxzz/graft/viewer/gpu"
)

const (
	suggestRows   = 8
	suggestScale  = 2
	suggestPad    = 6
	suggestBottom = 46
	suggestLeft   = 6
	suggestEdge   = 1
	suggestBar    = 2
	suggestGap    = 8
)

var (
	suggestShade  = gpu.RGBA(0, 0, 0, 0.45)
	suggestBack   = gpu.RGBA(0.05, 0.06, 0.09, 0.94)
	suggestFrame  = gpu.RGBA(1, 1, 1, 0.09)
	suggestPick   = gpu.RGBA(0.49, 0.83, 0.99, 0.16)
	suggestAccent = gpu.RGBA(0.49, 0.83, 0.99, 1)
	suggestTyped  = gpu.RGBA(0.49, 0.83, 0.99, 0.9)
	suggestBright = gpu.RGBA(0.95, 0.96, 0.98, 1)
	suggestText   = gpu.RGBA(0.72, 0.75, 0.80, 1)
	suggestFaint  = gpu.RGBA(0.45, 0.48, 0.54, 1)
)

// Suggesting is the list under the chat while a command is being typed. It is a
// Layer like everything else the viewer draws, so it sits in the same stack as
// the hud and a plugin's own overlay.
type Suggesting struct {
	words  []string
	typed  string
	chosen int
	first  int
	moved  bool
	from   int
	to     int
}

// Offers replaces what is on show. The choice is kept when the list has not
// changed, so typing a letter that narrows nothing does not jump the highlight.
func (s *Suggesting) Offers(words []string, typed string, from, to int) {
	s.typed, s.from, s.to = typed, from, to

	if slices.Equal(s.words, words) {
		return
	}

	s.words = words
	s.chosen = 0
	s.first = 0
	s.moved = false
}

func (s *Suggesting) Clear() {
	s.words = nil
	s.chosen = 0
	s.first = 0
	s.moved = false
}

func (s *Suggesting) Showing() bool {
	return len(s.words) > 0
}

// Moved reports whether the player picked a row rather than being handed the
// first one, which is what tells Enter to accept instead of send
func (s *Suggesting) Moved() bool {
	return s.moved
}

func (s *Suggesting) Move(by int) {
	if len(s.words) == 0 {
		return
	}

	s.chosen = (s.chosen + by + len(s.words)) % len(s.words)
	s.moved = true

	s.scroll()
}

// scroll keeps the chosen row inside the window, since a list longer than the
// panel would otherwise highlight a row nobody can see
func (s *Suggesting) scroll() {
	if s.chosen < s.first {
		s.first = s.chosen
	}
	if s.chosen >= s.first+suggestRows {
		s.first = s.chosen - suggestRows + 1
	}
}

// Chosen is the word to splice in and the run it replaces
func (s *Suggesting) Chosen() (string, int, int, bool) {
	if len(s.words) == 0 {
		return "", 0, 0, false
	}

	return s.words[s.chosen], s.from, s.to, true
}

// Adds reports whether accepting would change the line at all. A word already
// typed in full is not worth a keystroke, so Enter sends instead.
func (s *Suggesting) Adds() bool {
	word, _, _, chosen := s.Chosen()

	return chosen && word != s.typed
}

func (s *Suggesting) Draw(canvas *Canvas) {
	if len(s.words) == 0 {
		return
	}

	shown := s.window()

	line := float32(glyphSize * suggestScale)
	rows := float32(len(shown))*line + 2*suggestPad

	height := rows
	if s.more() > 0 {
		height += line
	}

	widest := canvas.TextWidth(s.tally(), suggestScale)
	for _, word := range s.words {
		widest = max(widest, canvas.TextWidth(word, suggestScale))
	}

	screen := canvas.Screen()
	corner := gpu.At(suggestLeft, screen.Height()-suggestBottom-height)
	panel := gpu.RectAt(corner, widest+2*suggestPad+suggestBar+suggestGap, height)

	canvas.Fill(panel.Offset(2, 2), suggestShade)
	canvas.Fill(panel, suggestBack)
	frame(canvas, panel)

	for row, word := range shown {
		at := corner.Offset(suggestPad+suggestBar+suggestGap, suggestPad+float32(row)*line)

		if row+s.first == s.chosen {
			s.pick(canvas, gpu.RectAt(gpu.At(corner.X, at.Y-2), panel.Width(), line))
		}

		s.word(canvas, word, at, row+s.first == s.chosen)
	}

	if s.more() > 0 {
		canvas.Shadowed(s.tally(), corner.Offset(suggestPad+suggestBar+suggestGap, rows-suggestPad), suggestScale, suggestFaint)
	}
}

// window is the run of words the panel has room for, following the choice
func (s *Suggesting) window() []string {
	last := min(s.first+suggestRows, len(s.words))

	return s.words[s.first:last]
}

func (s *Suggesting) more() int {
	return len(s.words) - len(s.window()) - s.first
}

func (s *Suggesting) tally() string {
	if s.more() <= 0 {
		return ""
	}

	return "+" + strconv.Itoa(s.more()) + " more"
}

// pick is the chosen row: a wash across the panel and a bar down its edge, so
// the choice reads even where the wash is too faint to see
func (s *Suggesting) pick(canvas *Canvas, row gpu.Rect) {
	canvas.Fill(row, suggestPick)
	canvas.Fill(gpu.RectAt(row.Min, suggestBar, row.Height()), suggestAccent)
}

// word draws what was typed apart from what completing it would add, which is
// the whole point of the list: you see what the keystroke buys you
func (s *Suggesting) word(canvas *Canvas, word string, at gpu.Point, chosen bool) {
	rest := suggestText
	if chosen {
		rest = suggestBright
	}

	if s.typed == "" || !strings.HasPrefix(word, s.typed) {
		canvas.Shadowed(word, at, suggestScale, rest)

		return
	}

	pen := canvas.Shadowed(s.typed, at, suggestScale, suggestTyped)
	canvas.Shadowed(word[len(s.typed):], pen, suggestScale, rest)
}

func frame(canvas *Canvas, panel gpu.Rect) {
	canvas.Fill(gpu.RectAt(panel.Min, panel.Width(), suggestEdge), suggestFrame)
	canvas.Fill(gpu.RectAt(panel.Min.Offset(0, panel.Height()-suggestEdge), panel.Width(), suggestEdge), suggestFrame)
	canvas.Fill(gpu.RectAt(panel.Min, suggestEdge, panel.Height()), suggestFrame)
	canvas.Fill(gpu.RectAt(panel.Min.Offset(panel.Width()-suggestEdge, 0), suggestEdge, panel.Height()), suggestFrame)
}

// suggest keeps the list under the chat current and answers whether a key was
// spent moving through it, so a stroke that picked a word never also submits
func (v *Viewer) suggest() bool {
	if v.stage.offering == nil {
		return false
	}

	line, cursor := v.chat.Line()

	words, from, to := v.stage.offering(line, cursor)
	v.suggesting.Offers(words, run(line, from, to), from, to)

	if !v.suggesting.Showing() {
		return false
	}

	if v.edges.key(gpu.KeyUp).started {
		v.suggesting.Move(-1)

		return true
	}
	if v.edges.key(gpu.KeyDown).started {
		v.suggesting.Move(1)

		return true
	}

	if v.edges.key(gpu.KeyTab).started {
		return v.accept()
	}

	// Enter takes the row once you have picked one, and otherwise only when the
	// word would add something. A word already typed in full is not worth a
	// keystroke, so the line goes instead of standing still.
	if v.edges.key(gpu.KeyEnter).started && (v.suggesting.Moved() || v.suggesting.Adds()) {
		return v.accept()
	}

	return false
}

func (v *Viewer) accept() bool {
	word, from, to, chosen := v.suggesting.Chosen()
	if !chosen {
		return false
	}

	v.chat.Splice(from, to, word)
	v.suggesting.Clear()

	return true
}

// run is what the offer would replace, which the list shows apart from the rest
// of each word
func run(line string, from, to int) string {
	if from < 0 || to > len(line) || from > to {
		return ""
	}

	return line[from:to]
}
