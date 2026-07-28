package viewer

import (
	"sync"

	"github.com/lrnxzz/go-craft/viewer/gpu"
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
)

type chatLine struct {
	text string
	at   float64
}

// Chat is a Layer: it paints on whatever canvas the viewer hands it and owns no
// GPU resource of its own, the font belonging to the canvas
type Chat struct {
	mu     sync.Mutex
	clock  func() float64
	lines  []chatLine
	input  []rune
	typing bool
}

func NewChat(clock func() float64) *Chat {
	return &Chat{clock: clock}
}

func (c *Chat) Push(text string) {
	line := chatLine{
		text: text,
		at:   c.clock(),
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
		canvas.Shadowed(c.prompt(now), input.Min.Offset(chatMargin+chatScale, 2*chatScale), chatScale, textFill)

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

		canvas.Fill(line, chatBacking)
		canvas.Shadowed(lines[index].text, line.Min.Offset(chatMargin+chatScale, chatScale), chatScale, textFill)
		shown++
	}
}

func (c *Chat) prompt(now float64) string {
	prompt := chatPrompt + string(c.input)
	if int(now*caretHertz)%2 == 0 {
		prompt += chatCaret
	}

	return prompt
}
