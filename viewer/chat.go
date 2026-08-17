package viewer

import (
	"sync"

	"github.com/lrnxzz/graft/viewer/gpu"
)

const (
	chatScale     = 2
	chatVisible   = 10
	chatFade      = 10.0
	chatWidth     = 330
	chatLineSpan  = 9
	chatInputSpan = 12
	chatHistory   = 100
	chatMargin    = 2
	caretHertz    = 3

	chatPrompt = "> "
	chatCaret  = "_"
)

var (
	chatBacking  = gpu.RGBA(0, 0, 0, 0.4)
	inputBacking = gpu.RGBA(0, 0, 0, 0.5)
	textFill     = gpu.White
	textShadow   = gpu.Shade(0.25, 0.25, 0.25)

	// what the client says is backed and edged so it never reads as something
	// another player typed on the server
	clientBacking = gpu.RGBA(0.05, 0.09, 0.14, 0.62)
	clientEdge    = gpu.RGBA(0.49, 0.83, 0.99, 0.85)

	chatPromptFill = gpu.RGBA(0.49, 0.83, 0.99, 1)
	chatGhostFill  = gpu.RGBA(0.55, 0.58, 0.64, 1)
	chatGoodFill   = gpu.RGBA(0.62, 0.90, 0.62, 1)
	chatBadFill    = gpu.RGBA(0.98, 0.51, 0.49, 1)
)

// A Mark is a run of the line being typed and whether whoever reads it made
// sense of it. Marks are how the chat colours a command without knowing what a
// command is.
type Mark struct {
	From int
	To   int
	Good bool
}

// Reading is the line as its reader sees it, and what finishing the last word
// would add
type Reading struct {
	Marks []Mark
	Ghost string
}

type chatLine struct {
	text   string
	at     float64
	client bool
}

// Chat is a Layer: it paints on whatever canvas the viewer hands it and owns no
// GPU resource of its own, the font belonging to the canvas
type Chat struct {
	mu     sync.Mutex
	clock  func() float64
	lines  []chatLine
	input  []rune
	typing bool
	reads  func(line string) Reading
}

// Reads registers who judges the line being typed. Without one the chat draws
// plain text, which is what a line that is not a command should look like.
func (c *Chat) Reads(read func(line string) Reading) {
	c.reads = read
}

func NewChat(clock func() float64) *Chat {
	return &Chat{clock: clock}
}

// Push is a line from the server, drawn plain
func (c *Chat) Push(text string) {
	c.write(text, false)
}

// Says is the client speaking, drawn with a mark of its own so nobody mistakes
// it for another player
func (c *Chat) Says(text string) {
	c.write(text, true)
}

func (c *Chat) write(text string, client bool) {
	line := chatLine{
		text:   text,
		at:     c.clock(),
		client: client,
	}

	c.mu.Lock()
	c.lines = append(c.lines, line)
	if len(c.lines) > chatHistory {
		c.lines = c.lines[len(c.lines)-chatHistory:]
	}
	c.mu.Unlock()
}

func (c *Chat) Typing() bool {
	return c.typing
}

func (c *Chat) Open(prefill string) {
	c.typing = true
	c.input = []rune(prefill)
}

func (c *Chat) Type(typed []rune) {
	c.input = append(c.input, typed...)
}

// Lines is what has been said, oldest first
func (c *Chat) Lines() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	said := make([]string, 0, len(c.lines))
	for _, line := range c.lines {
		said = append(said, line.text)
	}

	return said
}

// Line is what is being typed and where the caret sits, which is what a
// completer needs to know what the player is in the middle of
func (c *Chat) Line() (string, int) {
	return string(c.input), len(c.input)
}

// Splice replaces a run of the line with a word, which is how an accepted
// suggestion lands without the player deleting what they already typed
func (c *Chat) Splice(from, to int, word string) {
	if from < 0 || to > len(c.input) || from > to {
		return
	}

	replaced := append([]rune(word), c.input[to:]...)
	c.input = append(c.input[:from], replaced...)
}

func (c *Chat) Erase() {
	if len(c.input) > 0 {
		c.input = c.input[:len(c.input)-1]
	}
}

func (c *Chat) Cancel() {
	c.typing = false
	c.input = nil
}

func (c *Chat) Submit() (string, bool) {
	message := string(c.input)
	c.Cancel()

	return message, message != ""
}

func (c *Chat) history() []chatLine {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.lines
}

func (c *Chat) Draw(canvas *Canvas) {
	now := c.clock()
	screen := canvas.Screen()
	bottom := screen.Max.Y - chatMargin

	if c.typing {
		input := gpu.RectAt(
			gpu.At(chatMargin, bottom-chatInputSpan*chatScale),
			screen.Width()-2*chatMargin, chatInputSpan*chatScale)

		canvas.Fill(input, inputBacking)
		c.typed(canvas, input.Min.Offset(chatMargin+chatScale, 2*chatScale), now)

		bottom = input.Min.Y - chatMargin
	}

	lines := c.history()

	shown := 0
	for index := len(lines) - 1; index >= 0 && shown < chatVisible; index-- {
		if !c.typing && now-lines[index].at > chatFade {
			break
		}

		line := gpu.RectAt(
			gpu.At(chatMargin, bottom-float32((shown+1)*chatLineSpan)*chatScale),
			chatWidth*chatScale, chatLineSpan*chatScale)

		said := lines[index]

		canvas.Fill(line, backing(said))
		if said.client {
			canvas.Fill(gpu.RectAt(line.Min, chatScale, line.Height()), clientEdge)
		}

		canvas.Shadowed(said.text, line.Min.Offset(chatMargin+2*chatScale, chatScale), chatScale, textFill)
		shown++
	}
}

func backing(line chatLine) gpu.Color {
	if line.client {
		return clientBacking
	}

	return chatBacking
}

// typed draws the line being written: coloured by what its reader made of it,
// with the rest of the word it is heading for behind the caret
func (c *Chat) typed(canvas *Canvas, at gpu.Point, now float64) {
	line := string(c.input)
	reading := c.reading(line)

	pen := canvas.Shadowed(chatPrompt, at, chatScale, chatPromptFill)
	pen = c.runs(canvas, line, reading.Marks, pen)

	// the caret keeps its place whether or not it is drawn this frame, or the
	// ghost behind it would jump every half second
	caret := canvas.TextWidth(chatCaret, chatScale)
	if int(now*caretHertz)%2 == 0 {
		canvas.Shadowed(chatCaret, pen, chatScale, textFill)
	}

	if reading.Ghost != "" {
		canvas.Shadowed(reading.Ghost, pen.Offset(caret, 0), chatScale, chatGhostFill)
	}
}

func (c *Chat) reading(line string) Reading {
	if c.reads == nil {
		return Reading{}
	}

	return c.reads(line)
}

// runs paints the line a mark at a time, leaving whatever no mark covers in the
// plain colour so a line that is not a command looks untouched
func (c *Chat) runs(canvas *Canvas, line string, marks []Mark, pen gpu.Point) gpu.Point {
	at := 0

	for _, mark := range marks {
		if mark.From < at || mark.To > len(line) || mark.From > mark.To {
			continue
		}

		pen = canvas.Shadowed(line[at:mark.From], pen, chatScale, textFill)
		pen = canvas.Shadowed(line[mark.From:mark.To], pen, chatScale, tint(mark))

		at = mark.To
	}

	return canvas.Shadowed(line[at:], pen, chatScale, textFill)
}

func tint(mark Mark) gpu.Color {
	if mark.Good {
		return chatGoodFill
	}

	return chatBadFill
}
