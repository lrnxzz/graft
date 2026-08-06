package viewer

import (
	"github.com/lrnxzz/go-craft/viewer/gpu"
)

const (
	suggestRows   = 8
	suggestScale  = 2
	suggestPad    = 4
	suggestBottom = 46
	suggestWidth  = 6
)

var (
	suggestBack   = gpu.RGBA(0.04, 0.05, 0.08, 0.92)
	suggestEdge   = gpu.RGBA(1, 1, 1, 0.10)
	suggestChosen = gpu.RGBA(0.49, 0.83, 0.99, 0.22)
	suggestText   = gpu.RGBA(0.91, 0.93, 0.96, 1)
	suggestDim    = gpu.RGBA(0.91, 0.93, 0.96, 0.45)
)

// Suggesting is the list under the chat while a command is being typed. It is a
// Layer like everything else the viewer draws, so it sits in the same stack as
// the hud and a plugin's own overlay.
type Suggesting struct {
	words  []string
	chosen int
	from   int
	to     int
}

// Offers replaces what is on show. The choice is kept when the list has not
// changed, so typing a letter that narrows nothing does not jump the highlight.
func (s *Suggesting) Offers(words []string, from, to int) {
	if same(s.words, words) {
		s.from, s.to = from, to

		return
	}

	s.words, s.from, s.to = words, from, to
	s.chosen = 0
}

func (s *Suggesting) Clear() {
	s.words = nil
	s.chosen = 0
}

func (s *Suggesting) Showing() bool {
	return len(s.words) > 0
}

func (s *Suggesting) Move(by int) {
	if len(s.words) == 0 {
		return
	}

	s.chosen = (s.chosen + by + len(s.words)) % len(s.words)
}

// Chosen is the word to splice in and the run it replaces
func (s *Suggesting) Chosen() (string, int, int, bool) {
	if len(s.words) == 0 {
		return "", 0, 0, false
	}

	return s.words[s.chosen], s.from, s.to, true
}

func (s *Suggesting) Draw(canvas *Canvas) {
	if len(s.words) == 0 {
		return
	}

	shown := s.words
	if len(shown) > suggestRows {
		shown = shown[:suggestRows]
	}

	line := float32(glyphSize * suggestScale)
	height := float32(len(shown))*line + 2*suggestPad

	widest := float32(0)
	for _, word := range shown {
		widest = max(widest, canvas.TextWidth(word, suggestScale))
	}

	screen := canvas.Screen()
	corner := gpu.At(suggestWidth, screen.Height()-suggestBottom-height)
	panel := gpu.RectAt(corner, widest+2*suggestPad+line, height)

	canvas.Fill(panel, suggestBack)
	canvas.Fill(gpu.RectAt(corner, panel.Width(), 1), suggestEdge)

	for row, word := range shown {
		at := corner.Offset(suggestPad, suggestPad+float32(row)*line)

		tint := suggestDim
		if row == s.chosen {
			canvas.Fill(gpu.RectAt(corner.Offset(0, at.Y-corner.Y), panel.Width(), line), suggestChosen)

			tint = suggestText
		}

		canvas.Shadowed(word, at, suggestScale, tint)
	}
}

func same(before, after []string) bool {
	if len(before) != len(after) {
		return false
	}

	for index := range before {
		if before[index] != after[index] {
			return false
		}
	}

	return true
}
