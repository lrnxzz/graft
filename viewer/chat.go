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
	untextured   = gpu.UV{U0: -1, V0: -1, U1: -1, V1: -1}
	chatBacking  = shade(0, 0, 0, 0.4)
	inputBacking = shade(0, 0, 0, 0.5)
	textFill     = shade(1, 1, 1, 1)
	textShadow   = shade(0.25, 0.25, 0.25, 1)
)

type chatLine struct {
	text string
	at   float64
}

type Chat struct {
	mu      sync.Mutex
	program *gpu.Program
	font    *Font
	clock   func() float64
	lines   []chatLine
	input   []rune
	typing  bool
}

func NewChat(clock func() float64) (*Chat, error) {
	program, err := gpu.NewProgram(hudVertexShader, hudFragmentShader)
	if err != nil {
		return nil, err
	}

	font, err := LoadFont()
	if err != nil {
		return nil, err
	}

	return &Chat{
		program: program,
		font:    font,
		clock:   clock,
	}, nil
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

func (c *Chat) Draw(screen gpu.Rect, now float64) {
	vertices := c.layout(screen, now)
	if len(vertices) == 0 {
		return
	}

	c.program.Use()
	c.program.Vec2("screen", screen.Width(), screen.Height())
	c.program.Int("icons", 0)
	c.font.Bind(0)

	batch := uploadHud(vertices)
	batch.Draw()
	batch.Delete()
}

func (c *Chat) layout(screen gpu.Rect, now float64) []float32 {
	var vertices []float32

	backdrop := func(area gpu.Rect, tint hudColor) {
		vertices = hudTintedQuad(vertices, area, untextured, tint)
	}
	shadowed := func(text string, pen gpu.Point) {
		vertices = c.font.Emit(vertices, text, pen.Offset(chatScale, chatScale), chatScale, textShadow)
		vertices = c.font.Emit(vertices, text, pen, chatScale, textFill)
	}

	bottom := screen.Max.Y - chatMargin
	if c.typing {
		input := gpu.RectAt(
			gpu.At(chatMargin, bottom-chatInputSpan*chatScale),
			screen.Width()-2*chatMargin, chatInputSpan*chatScale)
		backdrop(input, inputBacking)
		shadowed(c.prompt(now), input.Min.Offset(chatMargin+chatScale, 2*chatScale))

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
		backdrop(line, chatBacking)
		shadowed(lines[index].text, line.Min.Offset(chatMargin+chatScale, chatScale))
		shown++
	}

	return vertices
}

func (c *Chat) prompt(now float64) string {
	prompt := chatPrompt + string(c.input)
	if int(now*caretHertz)%2 == 0 {
		prompt += chatCaret
	}

	return prompt
}
