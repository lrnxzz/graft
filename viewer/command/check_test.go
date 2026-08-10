package command_test

import (
	"testing"

	"github.com/lrnxzz/go-craft/viewer/command"
)

func read(t *testing.T, line string) command.Reading {
	t.Helper()

	var seen walked

	return tree(&seen).Check(line, command.Calling(nil, nil))
}

// what the run covers, so a test says which letters went red rather than which
// index they started at
func covered(line string, mark command.Mark) string {
	if mark.From < 0 || mark.To > len(line) || mark.From > mark.To {
		return ""
	}

	return line[mark.From:mark.To]
}

func wrong(t *testing.T, line string) []string {
	t.Helper()

	var red []string
	for _, mark := range read(t, line).Marks {
		if !mark.Good {
			red = append(red, covered(line, mark))
		}
	}

	return red
}

// A word still being typed is unfinished, not wrong. Colouring it red on the way
// to being right would put the line in red for most of every keystroke.
func TestAHalfTypedWordIsNotWrongYet(t *testing.T) {
	if red := wrong(t, "ro"); len(red) > 0 {
		t.Errorf("a half-typed command went red: %v", red)
	}
	if red := wrong(t, "mine iron_o"); len(red) > 0 {
		t.Errorf("a half-typed block went red: %v", red)
	}
}

// and the ghost is what finishing it would add
func TestTheGhostIsWhatFinishingAdds(t *testing.T) {
	if ghost := read(t, "ro").Ghost; ghost != "ute" {
		t.Errorf("the ghost is %q, want ute", ghost)
	}
	if ghost := read(t, "mine iron_o").Ghost; ghost != "re" {
		t.Errorf("the ghost is %q, want re", ghost)
	}
}

// a word nothing could grow into is wrong straight away
func TestAWordNothingMatchesIsWrong(t *testing.T) {
	red := wrong(t, "teleport")
	if len(red) != 1 || red[0] != "teleport" {
		t.Errorf("marked %v wrong, want teleport", red)
	}
}

// once something follows it, a word has stopped being typed and has to be right
func TestAFinishedWordIsJudged(t *testing.T) {
	red := wrong(t, "ro ")
	if len(red) != 1 || red[0] != "ro" {
		t.Errorf("marked %v wrong, want the finished ro", red)
	}
}

// the words that do parse stay green, and only the offending one goes red
func TestOnlyTheOffendingWordGoesRed(t *testing.T) {
	line := "route walk seven"

	var green, red []string
	for _, mark := range read(t, line).Marks {
		if mark.Good {
			green = append(green, covered(line, mark))

			continue
		}

		red = append(red, covered(line, mark))
	}

	if len(green) != 2 || green[0] != "route" || green[1] != "walk" {
		t.Errorf("kept %v green, want route and walk", green)
	}
	if len(red) != 1 || red[0] != "seven" {
		t.Errorf("marked %v red, want seven", red)
	}
}

// a word past everything the command declared can never become right
func TestAWordPastTheLastArgumentIsWrong(t *testing.T) {
	red := wrong(t, "route walk 3 extra")
	if len(red) != 1 || red[0] != "extra" {
		t.Errorf("marked %v wrong, want extra", red)
	}
}

// an empty line has nothing to say about
func TestAnEmptyLineReadsBlank(t *testing.T) {
	reading := read(t, "")

	if len(reading.Marks) != 0 || reading.Ghost != "" {
		t.Errorf("an empty line read as %+v", reading)
	}
}
