package viewer

import (
	"strings"
	"testing"
)

// The chat shapes are what every command reports through, so they are checked
// for what reaches the line rather than for the escapes they happen to use.
func TestTheShapesReachTheChat(t *testing.T) {
	chat := NewChat(func() float64 { return 0 })
	says := Says(chat)

	says.Head("route")
	says.Row(1, "10 64 -3", "reachable")
	says.Pair("points", "3")
	says.Bar("walked", 2, 4)
	says.Keys("F", "leave the body", "Tab", "accept")

	lines := chat.Lines()
	if len(lines) != 5 {
		t.Fatalf("wrote %d lines, want 5", len(lines))
	}

	if !strings.Contains(lines[0], "ROUTE") {
		t.Errorf("the heading did not shout: %q", lines[0])
	}
	if !strings.Contains(lines[1], "10 64 -3") || !strings.Contains(lines[1], "reachable") {
		t.Errorf("the row lost part of itself: %q", lines[1])
	}
	if !strings.Contains(lines[3], "2/4") {
		t.Errorf("the bar never said the count: %q", lines[3])
	}
}

// the caret is the only shape that points, and it has to land under the run it
// was given however short the line is
func TestTheCaretLandsUnderTheRun(t *testing.T) {
	chat := NewChat(func() float64 { return 0 })

	Says(chat).Under("mine iron_or", 5, 12)

	lines := chat.Lines()
	if len(lines) != 2 {
		t.Fatalf("wrote %d lines, want the line and the caret", len(lines))
	}

	marks := strings.TrimLeft(stripped(lines[1]), " ")
	if len(marks) != 7 {
		t.Errorf("the caret is %d wide, want 7", len(marks))
	}
	if strings.Index(stripped(lines[1]), "^") != 5 {
		t.Errorf("the caret starts at %d, want 5", strings.Index(stripped(lines[1]), "^"))
	}
}

// a span past the end of the line must not panic
func TestACaretPastTheLineIsIgnored(t *testing.T) {
	chat := NewChat(func() float64 { return 0 })

	Says(chat).Under("short", 40, 50)

	if len(chat.Lines()) != 0 {
		t.Error("a caret past the line still drew")
	}
}

func stripped(line string) string {
	var kept strings.Builder

	runes := []rune(line)
	for at := 0; at < len(runes); at++ {
		if runes[at] == '§' {
			at++

			continue
		}

		kept.WriteRune(runes[at])
	}

	return kept.String()
}
