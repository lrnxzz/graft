package command_test

import (
	"testing"

	"github.com/lrnxzz/graft/viewer/command"
)

// The lexer is what every other piece reads, so it is checked shape by shape
// rather than by the sentences it happens to appear in.
func TestReadsEachShape(t *testing.T) {
	type reading struct {
		line   string
		shapes []command.Shape
		texts  []string
	}

	readings := []reading{
		{
			line:   "goto 10 64 -3",
			shapes: []command.Shape{command.Bare, command.Numeral, command.Numeral, command.Numeral},
			texts:  []string{"goto", "10", "64", "-3"},
		},
		{
			line:   "goto ~ ~5 ~-2",
			shapes: []command.Shape{command.Bare, command.Relative, command.Relative, command.Relative},
			texts:  []string{"goto", "~", "~5", "~-2"},
		},
		{
			line:   "goto ^ ^ ^10",
			shapes: []command.Shape{command.Bare, command.Local, command.Local, command.Local},
			texts:  []string{"goto", "^", "^", "^10"},
		},
		{
			line:   "mine iron_ore",
			shapes: []command.Shape{command.Bare, command.Bare},
			texts:  []string{"mine", "iron_ore"},
		},
		{
			line:   "goto 10,64,-3",
			shapes: []command.Shape{command.Bare, command.Numeral, command.Comma, command.Numeral, command.Comma, command.Numeral},
			texts:  []string{"goto", "10", ",", "64", ",", "-3"},
		},
		{
			line:   "say 2.5 blocks",
			shapes: []command.Shape{command.Bare, command.Numeral, command.Bare},
			texts:  []string{"say", "2.5", "blocks"},
		},
	}

	for _, want := range readings {
		tokens := command.Read(want.line)

		if tokens.Len() != len(want.shapes) {
			t.Errorf("%q read as %v, want %d tokens", want.line, tokens, len(want.shapes))

			continue
		}

		for index, token := range tokens.All() {
			if token.Shape != want.shapes[index] {
				t.Errorf("%q token %d is %s, want %s", want.line, index, token.Shape, want.shapes[index])
			}
			if token.Text != want.texts[index] {
				t.Errorf("%q token %d is %q, want %q", want.line, index, token.Text, want.texts[index])
			}
		}
	}
}

// a quoted run is one argument however many spaces it holds, and the escapes it
// carried are gone by the time a parameter sees it
func TestQuotesHoldSpacesAndEscapes(t *testing.T) {
	tokens := command.Read(`say "he said \"hello\" twice" now`)

	if tokens.Len() != 3 {
		t.Fatalf("read as %v, want three tokens", tokens)
	}

	quoted, _ := tokens.At(1)
	if quoted.Shape != command.Quoted {
		t.Errorf("shape = %s, want quoted text", quoted.Shape)
	}
	if want := `he said "hello" twice`; quoted.Text != want {
		t.Errorf("text = %q, want %q", quoted.Text, want)
	}
}

// half a line is what completion is asked about, so an unfinished quote must not
// throw away everything the player already typed
func TestAnUnfinishedQuoteStillReads(t *testing.T) {
	tokens := command.Read(`say "still typ`)

	if tokens.Len() != 2 {
		t.Fatalf("read as %v, want two tokens", tokens)
	}

	quoted, _ := tokens.At(1)
	if quoted.Text != "still typ" {
		t.Errorf("text = %q, want the part already typed", quoted.Text)
	}
}

// the span is what puts the caret under the right characters
func TestEveryTokenKnowsWhereItSat(t *testing.T) {
	line := "goto 10 64"
	tokens := command.Read(line)

	for _, token := range tokens.All() {
		if token.From < 0 || token.To > len(line) || token.From >= token.To {
			t.Errorf("%q claims the span %d..%d", token.Text, token.From, token.To)
		}
	}

	second, _ := tokens.At(1)
	if line[second.From:second.To] != "10" {
		t.Errorf("the span of %q cuts %q out of the line", second.Text, line[second.From:second.To])
	}
}

// a cursor inside a token is editing it; a cursor in the blank after one is
// starting the next argument, and completion turns on that difference
func TestTheCursorKnowsWhichArgumentItIsIn(t *testing.T) {
	tokens := command.Read("mine iron")

	index, inside := tokens.Under(7)
	if !inside || index != 1 {
		t.Errorf("a cursor inside the second token reported %d, inside=%v", index, inside)
	}

	index, inside = tokens.Under(len("mine iron") + 1)
	if inside {
		t.Errorf("a cursor past the line reported it was inside token %d", index)
	}
	if index != 2 {
		t.Errorf("a cursor after the line starts argument %d, want 2", index)
	}
}

// an offset carries its number, and a bare mark carries none
func TestOffsetsCarryTheirNumber(t *testing.T) {
	tokens := command.Read("goto ~ ~-2.5")

	bare, _ := tokens.At(1)
	if _, carried := bare.Offset(); carried {
		t.Error("a bare ~ carried an offset")
	}

	moved, _ := tokens.At(2)

	offset, carried := moved.Offset()
	if !carried || offset != "-2.5" {
		t.Errorf("offset = %q carried=%v, want -2.5", offset, carried)
	}
}

func TestAnEmptyLineReadsAsNothing(t *testing.T) {
	if tokens := command.Read("   "); tokens.Len() != 0 {
		t.Errorf("blank read as %v", tokens)
	}
}

// the caret is drawn under the span, which is how an error shows what it could
// not read
func TestTheCaretPointsAtTheSpan(t *testing.T) {
	line := "goto 10 64 x"
	tokens := command.Read(line)

	last, _ := tokens.At(3)
	drawn := tokens.Caret(last.From, last.To)

	want := line + "\n" + "           ^"
	if drawn != want {
		t.Errorf("caret drew\n%s\nwant\n%s", drawn, want)
	}
}
