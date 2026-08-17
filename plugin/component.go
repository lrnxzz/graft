package plugin

import (
	"github.com/dop251/goja"

	"strconv"
	"strings"
)

const (
	glyphHeight = 8
	iconSize    = 16
	barHeight   = 6
	optionPad   = 2
	scrollWidth = 3

	defaultScale    = 1
	defaultBarWidth = 80
)

func componentWords() []Word {
	return []Word{
		component("Panel", buildStack),
		component("Column", buildStack),
		component("Row", buildRow),
		component("Text", buildText),
		component("Bar", buildBar),
		component("Icon", buildIcon),
		component("List", buildList),
		component("Option", buildOption),
		component("Raw", buildRaw),
	}
}

// component is one tag a plugin can write, building the node the canvas draws
func component(tag string, build func(reading) Node) Word {
	return Word{
		Name: tag,
		Build: func(r *Runtime, call goja.FunctionCall) goja.Value {
			return r.vm.ToValue(build(r.reading(call.Argument(0))))
		},
	}
}

type frame struct {
	gap      float32
	padding  float32
	fill     Color
	children []Node
}

func frameOf(given reading) frame {
	return frame{
		gap:      given.field("gap").number(),
		padding:  given.field("padding").number(),
		fill:     given.field("background").color(),
		children: childrenOf(given),
	}
}

func (f frame) down(p *placement) (float32, float32) {
	var width, height float32
	for index, child := range f.children {
		childWidth, childHeight := child.Measure(p)

		width = max(width, childWidth)
		height += childHeight
		if index > 0 {
			height += f.gap
		}
	}

	return width + 2*f.padding, height + 2*f.padding
}

func (f frame) across(p *placement) (float32, float32) {
	var width, height float32
	for index, child := range f.children {
		childWidth, childHeight := child.Measure(p)

		width += childWidth
		if index > 0 {
			width += f.gap
		}
		height = max(height, childHeight)
	}

	return width + 2*f.padding, height + 2*f.padding
}

func (f frame) stack(p *placement, at box) {
	pen := at.y + f.padding
	for _, child := range f.children {
		_, height := child.Measure(p)
		child.Paint(p, box{
			x:      at.x + f.padding,
			y:      pen,
			width:  at.width - 2*f.padding,
			height: height,
		})

		pen += height + f.gap
	}
}

func (f frame) line(p *placement, at box) {
	pen := at.x + f.padding
	for _, child := range f.children {
		width, height := child.Measure(p)
		child.Paint(p, box{
			x:      pen,
			y:      at.y + f.padding,
			width:  width,
			height: height,
		})

		pen += width + f.gap
	}
}

func (f frame) backdrop(p *placement, at box) {
	if f.fill.Transparent() {
		return
	}

	p.surface.Fill(at.x, at.y, at.width, at.height, f.fill)
}

type stack struct {
	frame
	anchor Anchor
}

func buildStack(given reading) Node {
	return stack{
		frame:  frameOf(given),
		anchor: Anchor(given.field("anchor").text()),
	}
}

func (s stack) Anchor() Anchor {
	return s.anchor
}

func (s stack) Measure(p *placement) (float32, float32) {
	return s.down(p)
}

func (s stack) Paint(p *placement, at box) {
	s.backdrop(p, at)
	s.stack(p, at)
}

type row struct {
	frame
	anchor Anchor
}

func buildRow(given reading) Node {
	return row{
		frame:  frameOf(given),
		anchor: Anchor(given.field("anchor").text()),
	}
}

func (r row) Anchor() Anchor {
	return r.anchor
}

func (r row) Measure(p *placement) (float32, float32) {
	return r.across(p)
}

func (r row) Paint(p *placement, at box) {
	r.backdrop(p, at)
	r.line(p, at)
}

type text struct {
	written string
	scale   float32
	color   Color
}

func buildText(given reading) Node {
	written := make([]string, 0, 4)
	for _, child := range childrenOf(given) {
		spoken, speaks := child.(text)
		if speaks {
			written = append(written, spoken.written)
		}
	}

	return text{
		written: strings.Join(written, " "),
		scale:   given.field("scale").number(),
		color:   given.field("color").color().or(white),
	}
}

func (t text) sized() float32 {
	if t.scale <= 0 {
		return defaultScale
	}

	return t.scale
}

func (t text) Measure(p *placement) (float32, float32) {
	return p.surface.Measure(t.written, t.sized()), glyphHeight * t.sized()
}

func (t text) Paint(p *placement, at box) {
	p.surface.Text(t.written, at.x, at.y, t.sized(), t.color)
}

type bar struct {
	value float32
	max   float32
	width float32
	color Color
	fill  Color
}

// parsed once: overlay rebuilds the tree every frame, and these never change
var (
	barGreen   = ParseColor("#3c3")
	barShade   = ParseColor("#0006")
	listRail   = ParseColor("#fff6")
	optionWash = ParseColor("#ffffff22")
)

func buildBar(given reading) Node {
	return bar{
		value: given.field("value").number(),
		max:   given.field("max").number(),
		width: given.field("width").number(),
		color: given.field("color").color().or(barGreen),
		fill:  barShade,
	}
}

func (b bar) Measure(*placement) (float32, float32) {
	if b.width <= 0 {
		return defaultBarWidth, barHeight
	}

	return b.width, barHeight
}

func (b bar) Paint(p *placement, at box) {
	p.surface.Fill(at.x, at.y, at.width, at.height, b.fill)

	if b.max <= 0 {
		return
	}

	filled := at.width * min(b.value/b.max, 1)
	p.surface.Fill(at.x, at.y, filled, at.height, b.color)
}

type icon struct {
	sprite Sprite
	badge  string
	color  Color
}

func buildIcon(given reading) Node {
	return icon{
		sprite: Sprite(given.field("item").text()),
		badge:  given.field("badge").text(),
		color:  white,
	}
}

func (i icon) Measure(*placement) (float32, float32) {
	return iconSize, iconSize
}

func (i icon) Paint(p *placement, at box) {
	p.surface.Icon(i.sprite, at.x, at.y, iconSize)

	if i.badge == "" {
		return
	}

	width := p.surface.Measure(i.badge, defaultScale)
	p.surface.Text(i.badge, at.x+iconSize-width, at.y+iconSize-glyphHeight, defaultScale, i.color)
}

type list struct {
	frame
	height float32
	rail   Color
}

func buildList(given reading) Node {
	return list{
		frame:  frameOf(given),
		height: given.field("height").number(),
		rail:   listRail,
	}
}

func (l list) Measure(p *placement) (float32, float32) {
	width, _ := l.down(p)

	return width + scrollWidth, l.height
}

func (l list) Paint(p *placement, at box) {
	l.backdrop(p, at)

	scrolled := p.scroll.at()
	pen := at.y - scrolled
	for _, child := range l.children {
		_, height := child.Measure(p)

		if pen+height > at.y && pen < at.y+at.height {
			child.Paint(p, box{
				x:      at.x,
				y:      pen,
				width:  at.width - scrollWidth,
				height: height,
			})
		}

		pen += height + l.gap
	}

	content := pen + scrolled - at.y
	p.scroll.fits(content, at.height)

	l.rails(p, at, scrolled, content)
}

func (l list) rails(p *placement, at box, scrolled, content float32) {
	if content <= at.height {
		return
	}

	span := at.height * at.height / content
	offset := at.height * scrolled / content

	p.surface.Fill(at.x+at.width-scrollWidth, at.y+offset, scrollWidth, span, l.rail)
}

type option struct {
	frame
	selected  bool
	pick      func()
	highlight Color
}

func buildOption(given reading) Node {
	return option{
		frame:     frameOf(given),
		selected:  given.field("selected").flag(),
		pick:      given.field("onPick").callback(),
		highlight: optionWash,
	}
}

func (o option) Measure(p *placement) (float32, float32) {
	width, height := o.across(p)

	return width + 2*optionPad, height + 2*optionPad
}

func (o option) Paint(p *placement, at box) {
	if o.selected {
		p.surface.Fill(at.x, at.y, at.width, at.height, o.highlight)
	}
	if o.pick != nil {
		p.picks = append(p.picks, Pick{
			at: at,
			on: o.pick,
		})
	}

	o.line(p, box{
		x:      at.x + optionPad,
		y:      at.y + optionPad,
		width:  at.width - 2*optionPad,
		height: at.height - 2*optionPad,
	})
}

type raw struct {
	width  float32
	height float32
	draw   func(Surface)
}

func buildRaw(given reading) Node {
	return raw{
		width:  given.field("width").number(),
		height: given.field("height").number(),
		draw:   given.field("draw").painter(),
	}
}

func (r raw) Measure(*placement) (float32, float32) {
	return r.width, r.height
}

func (r raw) Paint(p *placement, _ box) {
	if r.draw != nil {
		r.draw(p.surface)
	}
}

func childrenOf(given reading) []Node {
	return flattenChildren(given.field("children"))
}

func flattenChildren(given reading) []Node {
	if given.missing() {
		return nil
	}

	node, built := nodeOf(given)
	if built {
		only := []Node{node}

		return only
	}

	items := given.items()
	if items == nil {
		return asText(given)
	}

	var found []Node
	for _, item := range items {
		found = append(found, flattenChildren(item)...)
	}

	return found
}

func asText(given reading) []Node {
	written := strings.TrimSpace(given.text())
	if written == "" {
		return nil
	}

	return []Node{text{
		written: written,
		color:   white,
	}}
}

var white = Color{
	Red:   1,
	Green: 1,
	Blue:  1,
	Alpha: 1,
}

func ParseColor(written string) Color {
	digits := strings.TrimPrefix(strings.TrimSpace(written), "#")

	switch len(digits) {
	case 3, 4:
		digits = widen(digits)
	case 6, 8:
	default:
		return Color{}
	}

	channels := [4]float32{0, 0, 0, 1}
	for index := range len(digits) / 2 {
		value, err := strconv.ParseUint(digits[index*2:index*2+2], 16, 8)
		if err != nil {
			return Color{}
		}

		channels[index] = float32(value) / 255
	}

	return Color{
		Red:   channels[0],
		Green: channels[1],
		Blue:  channels[2],
		Alpha: channels[3],
	}
}

func widen(digits string) string {
	var wide strings.Builder
	for _, digit := range digits {
		wide.WriteRune(digit)
		wide.WriteRune(digit)
	}

	return wide.String()
}
